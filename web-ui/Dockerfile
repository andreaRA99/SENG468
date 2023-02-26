# syntax=docker/dockerfile:1

FROM python:3.11-bullseye AS builder

WORKDIR /app

COPY poetry-requirements.txt ./
RUN python -m venv .poetry-venv && .poetry-venv/bin/pip install -r poetry-requirements.txt

COPY pyproject.toml poetry.lock ./
RUN \
    --mount=type=cache,id=poetry,target=/var/cache/poetry,sharing=shared \
    python -m venv .venv && \
    POETRY_CACHE_DIR=/var/cache/poetry .poetry-venv/bin/poetry install --only main

COPY . .
RUN mkdir target && mv .venv static templates web_ui target/


FROM python:3.11-bullseye

WORKDIR /app
COPY --from=builder /app/target/ ./

CMD .venv/bin/gunicorn --workers 4 --worker-class uvicorn.workers.UvicornWorker --bind 0.0.0.0:8000 'web_ui.asgi:make_app()'
EXPOSE 8000
HEALTHCHECK CMD [".venv/bin/python", "-c", \
    "import httpx; httpx.get('http://daytrading.localhost:8000/health').raise_for_status()"]
