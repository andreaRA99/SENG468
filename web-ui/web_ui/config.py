import base64
from dataclasses import dataclass
from logging import getLogger
from typing import Any
from typing import Self
from typing import TypeVar


T = TypeVar("T")


logger = getLogger(__name__)


def _expect_type(value: Any, expected_type: type[T], path: str) -> T:
    if type(value) is not expected_type:
        raise TypeError(
            f"expected type {expected_type.__name__} for {path} but got {type(value).__name__} instead"
        )

    return value


class _ConfigDeserializer:
    _root: Any
    _path: tuple[str]
    _unused_keys: dict[str, None]

    def __init__(self, root: Any, path: tuple[str, ...]) -> None:
        self._root = root
        self._path = path
        self._unused_keys = dict.fromkeys(root, None)

    def use(self, key: str) -> Any:
        self._unused_keys.pop(key, None)
        return self._root[key]

    def expect(self, key: str, type_: type[T]) -> T:
        return _expect_type(self.use(key), type_, ".".join(self._path + (key,)))

    def expect_list(self, key: str, element_type: type[T]) -> tuple[T]:
        list_ = self.expect(key, list)

        for i, element in enumerate(list_):
            _expect_type(element, str, ".".join(self._path + (key,)) + f"[{i}]")

        return tuple(list_)

    def expect_table(self, key: str) -> "_ConfigDeserializer":
        return _ConfigDeserializer(self.use(key), self._path + (key,))

    @property
    def unused_keys(self) -> tuple[str]:
        return tuple(".".join(self._path + (key,)) for key in self._unused_keys)


@dataclass(frozen=True, slots=True)
class SessionsConfig:
    secret_keys: tuple[bytes]
    cookie_name: str
    cookie_secure: bool
    max_lifetime_minutes: int
    timeout_minutes: int


@dataclass(frozen=True, slots=True)
class AppConfig:
    allowed_hosts: tuple[str]
    csrf_allowed_origins: tuple[str]
    proxy_trusted_hosts: tuple[str]
    serve_static_files: bool
    sessions: SessionsConfig

    @classmethod
    def deserialize(cls: type[Self], root: Any) -> Self:
        de = _ConfigDeserializer(root, ())

        sessions_de = de.expect_table("sessions")
        encoded_secret_keys = sessions_de.expect_list("secret_keys", str)

        if not encoded_secret_keys:
            raise TypeError("sessions.secret_keys cannot be empty")

        secret_keys: list[bytes] = []

        for i, encoded_secret_key in enumerate(encoded_secret_keys):
            if len(encoded_secret_key) != 22:
                raise TypeError(f"sessions.secret_keys[{i}] has incorrect length")

            secret_key = base64.urlsafe_b64decode(encoded_secret_key + "==")
            assert len(secret_key) == 16

            secret_keys.append(secret_key)

        result = cls(
            allowed_hosts=de.expect_list("allowed_hosts", str),
            proxy_trusted_hosts=de.expect_list("proxy_trusted_hosts", str),
            csrf_allowed_origins=de.expect_list("csrf_allowed_origins", str),
            serve_static_files=de.expect("serve_static_files", bool),
            sessions=SessionsConfig(
                secret_keys=tuple(secret_keys),
                cookie_name=sessions_de.expect("cookie_name", str),
                cookie_secure=sessions_de.expect("cookie_secure", bool),
                max_lifetime_minutes=sessions_de.expect("max_lifetime_minutes", int),
                timeout_minutes=sessions_de.expect("timeout_minutes", int),
            ),
        )

        if unused_keys := [*de.unused_keys, *sessions_de.unused_keys]:
            logger.warn("Unused keys in configuration: %s", unused_keys)

        return result
