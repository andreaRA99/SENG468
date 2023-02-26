import base64
from urllib.parse import parse_qs

from starlette.requests import Request
from starlette.status import HTTP_400_BAD_REQUEST


def urlsafe_encode(bs: bytes) -> str:
    """
    Encodes bytes using base64 with the URL-safe alphabet and without padding.
    """
    return base64.urlsafe_b64encode(bs).rstrip(b"=").decode("ascii")


def urlsafe_decode(s: str) -> bytes:
    """
    Decodes a base64-encoded string in the URL-safe alphabet and without padding to bytes.
    """
    return base64.urlsafe_b64decode(
        s + "=="
    )  # it doesn’t care if there’s too much padding, only too little :(


FORM_MIME_TYPE = "application/x-www-form-urlencoded"


class BadRequest(Exception):
    def __init__(
        self,
        error: str,
        message: str,
        *,
        status_code=HTTP_400_BAD_REQUEST,
        headers=None,
        **additional_data,
    ) -> None:
        super().__init__(message)
        self.error = error
        self.additional_data = additional_data
        self.status_code = status_code
        self.headers = headers


class Body:
    __slots__ = ("_d",)

    def __init__(self, d: dict[str, list[str]]) -> None:
        self._d = d

    def one(self, key: str) -> str:
        value = self.optional(key)

        if value is None:
            raise BadRequest(
                "request-body.required-missing",
                f"Missing required field {key!r}",
                field=key,
            )

        return value

    def optional(self, key: str) -> str | None:
        values = self._d.get(key)

        if values is None:
            return None

        if len(values) != 1:
            raise BadRequest(
                "request-body.multiple-values",
                f"Unexpected multiple values for field {key!r}",
                field=key,
            )

        return values[0]

    def many(self, key: str) -> list[str]:
        return self._d.get(key, None) or []


async def read_body(request: Request) -> Body:
    if not (content_type := request.headers.get("Content-Type")):
        raise BadRequest("request-body.content-type-missing", "No Content-Type header")

    if content_type.split(";")[0] != FORM_MIME_TYPE:
        raise BadRequest(
            "request-body.content-type-invalid", f"Content-Type is not {FORM_MIME_TYPE}"
        )

    body = await request.body()

    try:
        parsed_body = parse_qs(
            body.decode(),
            keep_blank_values=True,
            strict_parsing=True,
            errors="strict",
        )
    except (UnicodeDecodeError, ValueError) as e:
        raise BadRequest(
            "request-body.body-invalid", "Invalid URL-encoded form data"
        ) from e

    return Body(parsed_body)
