import dataclasses
import struct
import time
from dataclasses import dataclass
from logging import getLogger
from typing import Self

from starlette.authentication import AuthCredentials
from starlette.authentication import AuthenticationBackend
from starlette.authentication import BaseUser
from starlette.authentication import SimpleUser
from starlette.authentication import UnauthenticatedUser
from starlette.requests import HTTPConnection
from starlette.requests import Request


logger = getLogger(__name__)


@dataclass(frozen=True, slots=True)
class Session:
    created: int
    """UNIX timestamp for the when this session was created."""

    refreshed: int
    """Number of seconds *since `created`* of the most recent refresh of this session."""

    username: str

    should_refresh: bool

    def serialize(self) -> bytes:
        return struct.pack("<IH", self.created, self.refreshed) + self.username.encode()

    def refresh(self) -> Self | None:
        if not self.should_refresh:
            return None

        return dataclasses.replace(
            self,
            refreshed=max(int(time.time() - self.created), 0),
        )

    @classmethod
    def deserialize(cls, request: Request, bs: bytes) -> Self | None:
        header_len = struct.calcsize("<IH")
        created, refreshed = struct.unpack("<IH", bs[:header_len])
        username = bs[header_len:].decode()
        config = request.app.state.config

        hard_expires = created + 60 * config.sessions.max_lifetime_minutes
        refresh_expires = created + refreshed + 60 * config.sessions.timeout_minutes
        now = time.time()

        if now > min(hard_expires, refresh_expires):
            return None

        return cls(
            created=created,
            refreshed=refreshed,
            username=username,
            should_refresh=True,
        )

    @classmethod
    def create(cls, username) -> Self:
        return cls(
            username=username,
            created=int(time.time()),
            refreshed=0,
            should_refresh=True,
        )


class SessionAuthBackend(AuthenticationBackend):
    async def authenticate(
        self, conn: HTTPConnection
    ) -> tuple[AuthCredentials, BaseUser] | None:
        session: Session | None = conn.state.session

        if session is None:
            return AuthCredentials(("guest",)), UnauthenticatedUser()

        return AuthCredentials(("user",)), SimpleUser(session.username)
