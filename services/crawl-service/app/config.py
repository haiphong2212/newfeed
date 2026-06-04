from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    service_name: str = "crawl-service"
    http_host: str = "0.0.0.0"
    http_port: int = 8010
    grpc_port: int = 50059
    database_url: str = "postgresql://postgres:postgres@postgres:5432/newfeed"
    redis_url: str = "redis://redis:6379/0"
    rabbitmq_url: str = "amqp://guest:guest@rabbitmq:5672/"
    auth_validate_url: str = "http://auth-service:8001/v1/auth/validate"
    crawl_cache_ttl_seconds: int = 900
    crawl_timeout_seconds: int = 12
    max_candidates: int = 5
    admin_bypass_token: str = ""


settings = Settings()
