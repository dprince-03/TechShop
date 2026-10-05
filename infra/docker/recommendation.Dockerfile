# TechShop recommendation system (batch + optional serving). Build from the repo root:
#   docker build -f infra/docker/recommendation.Dockerfile recommendation/
FROM python:3.13-slim
RUN useradd --create-home --uid 10001 recs
WORKDIR /app
COPY pyproject.toml README.md ./
COPY src ./src
COPY configs ./configs
RUN pip install --no-cache-dir . && mkdir -p /app/artifacts && chown recs /app/artifacts
USER recs
ENTRYPOINT ["techshop-recs"]
