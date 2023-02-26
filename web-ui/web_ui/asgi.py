import os
import stat
import tomllib
from datetime import datetime
from datetime import timezone
from logging import getLogger
from urllib.parse import quote as urlquote

from httpx import AsyncClient as HttpxClient
from starlette.applications import Starlette
from starlette.authentication import requires
from starlette.endpoints import HTTPEndpoint
from starlette.middleware import Middleware
from starlette.middleware.authentication import AuthenticationMiddleware
from starlette.middleware.trustedhost import TrustedHostMiddleware
from starlette.requests import Request
from starlette.responses import PlainTextResponse
from starlette.responses import RedirectResponse
from starlette.responses import Response
from starlette.routing import Mount
from starlette.routing import Route
from starlette.routing import Router
from starlette.staticfiles import StaticFiles
from starlette.status import HTTP_303_SEE_OTHER
from starlette.status import HTTP_400_BAD_REQUEST
from starlette.templating import Jinja2Templates
from uvicorn.middleware.proxy_headers import ProxyHeadersMiddleware

from .authentication import Session
from .authentication import SessionAuthBackend
from .config import AppConfig
from .middleware import ContentSecurityPolicyMiddleware
from .middleware import CsrfMiddleware
from .middleware import HeadersMiddleware
from .middleware import SessionMiddleware
from .util import read_body


CONFIG_PATH = "config.toml"

# Prevent many kinds of XSS-style injection and exfiltration.
CSP_MIDDLEWARE = Middleware(
    ContentSecurityPolicyMiddleware,
    policy_template="default-src 'none'; style-src 'nonce-{csp_nonce}'; font-src 'self'; frame-ancestors 'none'; block-all-mixed-content",
)

SECURITY_HEADERS_MIDDLEWARE = Middleware(
    HeadersMiddleware,
    headers={
        # Limit attacks that involve loading content from this origin as embedded.
        "Cross-Origin-Embedder-Policy": "require-corp",
        "Cross-Origin-Opener-Policy": "same-origin",
        "Cross-Origin-Resource-Policy": "same-origin",
        "X-Content-Type-Options": "nosniff",
        # Since we don't use the Referer header at all, we *would* use `no-referrer` here, but that causes POST requests to use `Origin: null`, breaking CSRF protection.
        "Referrer-Policy": "strict-origin",
    },
)

APP_HEADERS_MIDDLEWARE = Middleware(
    HeadersMiddleware,
    headers={
        "Cache-Control": "private, max-age=0, must-revalidate, no-store",
    },
)

STATIC_HEADERS_MIDDLEWARE = Middleware(
    HeadersMiddleware,
    headers={
        "Cache-Control": "public, max-age=0",
    },
)

logger = getLogger(__name__)

templates = Jinja2Templates(directory="templates")


class Home(HTTPEndpoint):
    @requires("user", redirect="Login")
    async def get(self, request: Request) -> Response:
        httpx_client: HttpxClient = request.app.state.httpx_client

        data = (
            await httpx_client.get(f"/displaysummary/{urlquote(request.user.username)}")
        ).json()
        balance = data["accStatus"]["cash_balance"]
        pending_transactions = data["limitOrders"] or []

        data["transactions"] = [
            t for t in data["transactions"] if t["command"] != "DISPLAY_SUMMARY"
        ]
        transaction_history = []
        pending_transactions = []
        last_buy_or_sell = None

        for t in data["transactions"]:
            timestamp = datetime.fromtimestamp(
                t["Timestamp"] / 1000
                if t["Timestamp"] > 1000000000000
                else t["Timestamp"],
                timezone.utc,
            )
            t["timestamp"] = timestamp.strftime("%Y-%m-%d %H:%M:%S %Z")

            if t["command"] in {"BUY", "SELL"}:
                t["type"] = t["command"]
                if t["command"] == "BUY":
                    t["quantity"] = t["funds"]
                else:
                    t["total-price"] = t["funds"]

                if last_buy_or_sell is not None:
                    logger.warn(
                        "Pending transactions are out of sync: too many buy/sell"
                    )
                last_buy_or_sell = t
            elif t["command"] in {
                "COMMIT_BUY",
                "COMMIT_SELL",
                "CANCEL_BUY",
                "CANCEL_SELL",
            }:
                if last_buy_or_sell is None:
                    logger.warn(
                        "Pending transactions are out of sync: too many commit/cancel"
                    )
                else:
                    if t["command"].startswith("COMMIT_"):
                        transaction_history.append(last_buy_or_sell)
                    else:
                        assert t["command"].startswith("CANCEL_")
                    last_buy_or_sell = None
            elif (
                t["action"] == "add" and t["funds"] != 0
            ):  # == 0 indicates result of sale
                t["type"] = "Add funds"
                t["total-price"] = t["funds"]
                transaction_history.append(t)

        if last_buy_or_sell is not None:
            last_buy_or_sell["can-commit"] = True
            pending_transactions.append(last_buy_or_sell)

        transaction_history.sort(key=lambda t: t["timestamp"], reverse=True)
        pending_transactions.sort(key=lambda t: t["timestamp"], reverse=True)

        return templates.TemplateResponse(
            "home.html",
            {
                "response": data,
                "request": request,
                "balance": balance,
                "pending_transactions": pending_transactions,
                "transaction_history": transaction_history,
            },
        )


class Login(HTTPEndpoint):
    @requires("guest")
    async def get(self, request: Request) -> Response:
        return templates.TemplateResponse("login.html", {"request": request})

    @requires("guest")
    async def post(self, request: Request) -> Response:
        body = await read_body(request)
        username = body.one("username")

        if (
            not 1 <= len(username) <= 40
            or not username.isascii()
            or not username.isalnum()
        ):
            return PlainTextResponse(
                "invalid username", status_code=HTTP_400_BAD_REQUEST
            )

        request.state.session = Session.create(username)
        return RedirectResponse("/", status_code=HTTP_303_SEE_OTHER)


class Logout(HTTPEndpoint):
    async def post(self, request: Request) -> Response:
        request.state.session = None
        return RedirectResponse(
            "/login",
            status_code=HTTP_303_SEE_OTHER,
            headers={
                "Clear-Site-Data": '"cache", "cookies", "storage", "executionContexts"',
            },
        )


class Fund(HTTPEndpoint):
    @requires("user", redirect="Login")
    async def get(self, request: Request) -> Response:
        httpx_client: HttpxClient = request.app.state.httpx_client
        data = (
            await httpx_client.get(f"/displaysummary/{urlquote(request.user.username)}")
        ).json()
        balance = data["accStatus"]["cash_balance"]
        return templates.TemplateResponse(
            "fund.html",
            {
                "request": request,
                "balance": balance,
            },
        )

    @requires("user", redirect="Login")
    async def post(self, request: Request) -> Response:
        httpx_client: HttpxClient = request.app.state.httpx_client
        body = await read_body(request)
        amount = float(body.one("amount"))

        response = await httpx_client.put(
            "/users/addBal",
            json={
                "id": request.user.username,
                "amount": amount,
            },
        )
        response.raise_for_status()
        logger.debug("Added balance: %s", response.text)

        return RedirectResponse("/", status_code=HTTP_303_SEE_OTHER)


class Trade(HTTPEndpoint):
    @requires("user", redirect="Login")
    async def get(self, request: Request) -> Response:
        httpx_client: HttpxClient = request.app.state.httpx_client
        stock_symbol = request.query_params["stock"].upper()
        automatic = request.query_params.get("automatic") == "yes"

        user = (
            await httpx_client.get(f"/displaysummary/{urlquote(request.user.username)}")
        ).json()
        owned = next(
            (
                stock["quantity"]
                for stock in (user["accStatus"]["stocks"] or ())
                if stock["symbol"].upper() == stock_symbol
            ),
            0,
        )

        response = await httpx_client.get(
            f"/users/{urlquote(request.user.username)}/quote/{urlquote(stock_symbol)}"
        )
        response.raise_for_status()
        data = response.json()

        return templates.TemplateResponse(
            "trade.html",
            {
                "request": request,
                "owned": owned,
                "stock": data["Stock"],
                "price": data["Price"],
                "automatic": automatic,
            },
        )

    @requires("user", redirect="Login")
    async def post(self, request: Request) -> Response:
        httpx_client: HttpxClient = request.app.state.httpx_client

        body = await read_body(request)

        match body.one("do"):
            case "buy":
                response = await httpx_client.post(
                    "/users/buy",
                    json={
                        "id": request.user.username,
                        "stock": request.query_params["stock"].upper(),
                        "amount": int(body.one("amount")),
                    },
                )

                if response.status_code != 200:
                    return PlainTextResponse(
                        response.json(), status_code=response.status_code
                    )

                logger.info("Set up a buy: %s", response.json())

                return RedirectResponse("/", status_code=HTTP_303_SEE_OTHER)

            case "sell":
                response = await httpx_client.post(
                    "/users/sell",
                    json={
                        "id": request.user.username,
                        "stock": request.query_params["stock"].upper(),
                        "amount": int(body.one("amount"))
                        * float(
                            body.one("price")
                        ),  # XXX: yes, $ when selling and # when buying
                    },
                )

                if response.status_code != 200:
                    return PlainTextResponse(
                        response.json(), status_code=response.status_code
                    )

                logger.info("Set up a sell: %s", response.json())

                return RedirectResponse("/", status_code=HTTP_303_SEE_OTHER)

            case "commit-buy":
                response = await httpx_client.post(
                    "/users/buy/commit",
                    json={
                        "id": request.user.username,
                    },
                )
                response.raise_for_status()

                logger.info("Committed a buy: %s", response.json())

                return RedirectResponse("/", status_code=HTTP_303_SEE_OTHER)

            case "cancel-buy":
                response = await httpx_client.delete(
                    f"/users/{urlquote(request.user.username)}/buy/cancel"
                )
                response.raise_for_status()

                logger.info("Cancelled a buy: %r", response.text)

                return RedirectResponse("/", status_code=HTTP_303_SEE_OTHER)

            case "commit-sell":
                response = await httpx_client.post(
                    "/users/sell/commit",
                    json={
                        "id": request.user.username,
                    },
                )
                response.raise_for_status()

                logger.info("Committed a sell: %s", response.json())

                return RedirectResponse("/", status_code=HTTP_303_SEE_OTHER)

            case "cancel-sell":
                response = await httpx_client.delete(
                    f"/users/{urlquote(request.user.username)}/sell/cancel"
                )
                response.raise_for_status()

                logger.info("Cancelled a sell: %r", response.text)

                return RedirectResponse("/", status_code=HTTP_303_SEE_OTHER)

            case "set-buy":
                response = await httpx_client.post(
                    "/users/set/buy",
                    json={
                        "ID": request.user.username,
                        "stock": request.query_params["stock"].upper(),
                        "amount": int(body.one("amount")),
                    },
                )
                response.raise_for_status()
                logger.info("Set a buy amount: %r", response.text)

                response = await httpx_client.post(
                    "/users/set/buy/trigger",
                    json={
                        "ID": request.user.username,
                        "stock": request.query_params["stock"].upper(),
                        "price": float(body.one("price")),
                    },
                )
                response.raise_for_status()
                logger.info("Set a buy trigger: %r", response.text)

                return RedirectResponse("/", status_code=HTTP_303_SEE_OTHER)

            case "set-sell":
                response = await httpx_client.post(
                    "/users/set/sell",
                    json={
                        "ID": request.user.username,
                        "stock": request.query_params["stock"].upper(),
                        "amount": int(body.one("amount")) * float(body.one("price")),
                    },
                )
                response.raise_for_status()
                logger.info("Set a sell amount: %r", response.text)

                response = await httpx_client.post(
                    "/users/set/sell/trigger",
                    json={
                        "ID": request.user.username,
                        "stock": request.query_params["stock"].upper(),
                        "price": float(body.one("price")),
                    },
                )
                response.raise_for_status()
                logger.info("Set a sell trigger: %r", response.text)

                return RedirectResponse("/", status_code=HTTP_303_SEE_OTHER)

            case _:
                return PlainTextResponse(
                    "Unrecognized action", status_code=HTTP_400_BAD_REQUEST
                )


async def healthcheck(_request: Request) -> Response:
    return PlainTextResponse("ok")


APP_ROUTES = (
    Route("/", Home),
    Route("/login", Login),
    Route("/logout", Logout),
    Route("/fund", Fund),
    Route("/trade", Trade),
    Route("/health", healthcheck),
)


def load_config() -> AppConfig:
    with open(CONFIG_PATH, "rb") as f:
        mode = os.fstat(f.fileno()).st_mode

        if mode & (stat.S_IROTH | stat.S_IRGRP) != 0:
            logger.warn(
                f"{CONFIG_PATH} contains secret keys and is more than user-readable"
            )

        config_root = tomllib.load(f)

    return AppConfig.deserialize(config_root)


def make_app(config: AppConfig | None = None) -> Starlette:
    if config is None:
        config = load_config()

    routes = []

    if config.serve_static_files:
        routes += [
            Mount(
                "/css",
                StaticFiles(directory="static/css"),
                middleware=(STATIC_HEADERS_MIDDLEWARE,),
            ),
        ]

    routes.append(
        Mount(
            "/",
            Router(APP_ROUTES),
            middleware=(
                APP_HEADERS_MIDDLEWARE,
                Middleware(
                    SessionMiddleware,
                    secret_keys=config.sessions.secret_keys,
                    cookie_name=config.sessions.cookie_name,
                    cookie_secure=config.sessions.cookie_secure,
                    session_type=Session,
                ),
                Middleware(
                    AuthenticationMiddleware,
                    backend=SessionAuthBackend(),
                ),
            ),
        ),
    )

    app = Starlette(
        routes=routes,
        middleware=(
            Middleware(
                TrustedHostMiddleware,
                allowed_hosts=config.allowed_hosts,
            ),
            Middleware(
                ProxyHeadersMiddleware,
                trusted_hosts=config.proxy_trusted_hosts,
            ),
            CSP_MIDDLEWARE,
            SECURITY_HEADERS_MIDDLEWARE,
            Middleware(
                CsrfMiddleware,
                allowed_origins=config.csrf_allowed_origins,
            ),
        ),
    )

    app.state.config = config
    app.state.httpx_client = HttpxClient(base_url=os.environ["TRANSACTION_SERVICE"])

    return app
