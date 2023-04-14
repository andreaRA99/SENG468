import hashlib
import hmac
import secrets
from logging import getLogger
from typing import Awaitable
from typing import Callable
from typing import Generic
from typing import Iterable
from typing import Mapping
from typing import Protocol
from typing import Self
from typing import TypeVar

from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import PlainTextResponse
from starlette.responses import Response
from starlette.status import HTTP_403_FORBIDDEN

from .util import urlsafe_decode
from .util import urlsafe_encode


SAFE_METHODS = frozenset({"GET", "HEAD", "OPTIONS", "TRACE"})

logger = getLogger(__name__)


class CsrfMiddleware(BaseHTTPMiddleware):
    """
    Protects against CSRF by checking the `Origin` header of requests.

    This is more efficient and reliable than CSRF tokens, and supported by all modern browsers.
    """

    allowed_origins: frozenset[str]

    def __init__(self, app, dispatch=None, *, allowed_origins: Iterable[str]) -> None:
        super().__init__(app, dispatch)
        self.allowed_origins = frozenset(allowed_origins)

    async def dispatch(
        self,
        request: Request,
        call_next: Callable[[Request], Awaitable[Response]],
    ) -> Response:
        if (
            request.method in SAFE_METHODS
            or request.headers.get("origin") in self.allowed_origins
        ):
            return await call_next(request)
        else:
            return PlainTextResponse(
                "CSRF check failed",
                status_code=HTTP_403_FORBIDDEN,
            )


class ContentSecurityPolicyMiddleware(BaseHTTPMiddleware):
    """
    Adds a `Content-Security-Policy` to every response, formatting in a randomly generated nonce.
    """

    policy_template: str

    def __init__(self, app, dispatch=None, *, policy_template: str) -> None:
        super().__init__(app, dispatch)
        self.policy_template = policy_template

    async def dispatch(
        self, request: Request, call_next: Callable[[Request], Awaitable[Response]]
    ) -> Response:
        csp_nonce = secrets.token_urlsafe(16)
        request.state.csp_nonce = csp_nonce
        response = await call_next(request)
        response.headers.append(
            "Content-Security-Policy", self.policy_template.format(csp_nonce=csp_nonce)
        )
        return response


class HeadersMiddleware(BaseHTTPMiddleware):
    """
    Adds the same set of headers to every response.
    """

    headers: dict[str, str]

    def __init__(self, app, dispatch=None, *, headers: Mapping) -> None:
        super().__init__(app, dispatch)
        self.headers = {**headers}

    async def dispatch(
        self, request: Request, call_next: Callable[[Request], Awaitable[Response]]
    ) -> Response:
        response = await call_next(request)
        response.headers.update(self.headers)
        return response


SESSION_COOKIE_KEY_ID_SIZE = 4
SESSION_COOKIE_HMAC_SIZE = 16


def _hmac(base: hmac.HMAC, message: bytes) -> bytes:
    h = base.copy()
    h.update(message)
    return h.digest()[:SESSION_COOKIE_HMAC_SIZE]


class SessionProtocol(Protocol):
    def refresh(self) -> Self | None:
        ...

    def serialize(self) -> bytes:
        ...


SessionT = TypeVar("SessionT", bound=SessionProtocol, covariant=True)


class SessionFactoryProtocol(Protocol, Generic[SessionT]):
    def deserialize(self, request: Request, bs: bytes) -> SessionT | None:
        ...

    def empty(self) -> SessionT:
        ...


class SessionMiddleware(BaseHTTPMiddleware):
    _sign: tuple[hmac.HMAC]
    _signing_secret_key_id: bytes
    _key_index_by_id: dict[bytes, int]

    cookie_name: str
    cookie_secure: bool
    session_type: SessionFactoryProtocol

    def __init__(
        self,
        app,
        dispatch=None,
        *,
        secret_keys: Iterable[bytes],
        cookie_name: str,
        cookie_secure: bool,
        session_type: SessionFactoryProtocol,
    ) -> None:
        super().__init__(app, dispatch)

        sign = []
        secret_key_ids = []

        for secret_key in secret_keys:
            sign.append(hmac.new(secret_key, None, "sha256"))
            secret_key_ids.append(
                hashlib.sha256(secret_key).digest()[:SESSION_COOKIE_KEY_ID_SIZE]
            )
            del secret_key

        self._sign = tuple(sign)
        self._signing_secret_key_id = secret_key_ids[0]
        self._key_index_by_id = dict(zip(secret_key_ids, range(len(secret_key_ids))))

        if len(self._key_index_by_id) != len(secret_key_ids):
            raise Exception("secret key id collision")

        self.cookie_name = cookie_name
        self.cookie_secure = cookie_secure
        self.session_type = session_type

    async def dispatch(
        self, request: Request, call_next: Callable[[Request], Awaitable[Response]]
    ) -> Response:
        encoded_cookie = request.cookies.get(self.cookie_name)
        session: SessionProtocol | None = None
        rekey = False

        if encoded_cookie is not None:
            try:
                cookie = urlsafe_decode(encoded_cookie)
            except ValueError as e:
                logger.warn("Session cookie was not encoded correctly: %s", e)
            else:
                if len(cookie) >= SESSION_COOKIE_KEY_ID_SIZE + SESSION_COOKIE_HMAC_SIZE:
                    key_index = self._key_index_by_id.get(
                        cookie[:SESSION_COOKIE_KEY_ID_SIZE]
                    )

                    if key_index is not None:
                        message = cookie[
                            SESSION_COOKIE_KEY_ID_SIZE + SESSION_COOKIE_HMAC_SIZE :
                        ]
                        provided_hmac = cookie[
                            SESSION_COOKIE_KEY_ID_SIZE : SESSION_COOKIE_KEY_ID_SIZE
                            + SESSION_COOKIE_HMAC_SIZE
                        ]
                        expected_hmac = _hmac(self._sign[key_index], message)

                        if hmac.compare_digest(provided_hmac, expected_hmac):
                            session = self.session_type.deserialize(request, message)
                            rekey = key_index != 0
                    else:
                        logger.warn(
                            "Session cookie was signed with unavailable secret key"
                        )

        request.state.session = session
        response = await call_next(request)
        session = request.state.session

        if session is None:
            response.delete_cookie(
                self.cookie_name,
                secure=self.cookie_secure,
                httponly=True,
                samesite="strict",
            )
        elif rekey or (session := session.refresh()) is not None:
            message = session.serialize()
            message_hmac = _hmac(self._sign[0], message)
            encoded_cookie = urlsafe_encode(
                self._signing_secret_key_id + message_hmac + message
            )
            response.set_cookie(
                self.cookie_name,
                encoded_cookie,
                secure=self.cookie_secure,
                httponly=True,
                samesite="strict",
            )

        return response
