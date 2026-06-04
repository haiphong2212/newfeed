FROM python:3.12-slim AS build
WORKDIR /app
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
RUN apt-get update && apt-get install -y --no-install-recommends gcc libxml2-dev libxslt1-dev && rm -rf /var/lib/apt/lists/*
COPY services/crawl-service/requirements.txt ./requirements.txt
RUN pip install --no-cache-dir -r requirements.txt
COPY api/proto /app/api/proto
COPY services/crawl-service /app/services/crawl-service
RUN mkdir -p /app/services/crawl-service/app/generated && \
    python -m grpc_tools.protoc -I /app/api/proto \
      --python_out=/app/services/crawl-service/app/generated \
      --grpc_python_out=/app/services/crawl-service/app/generated \
      /app/api/proto/crawl/v1/crawl.proto

FROM python:3.12-slim
WORKDIR /app
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
COPY --from=build /usr/local /usr/local
COPY --from=build /app/services/crawl-service /app
EXPOSE 8010 50059
CMD ["python", "-m", "app.main"]
