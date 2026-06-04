import asyncio
from concurrent import futures

import grpc
from grpc_health.v1 import health, health_pb2, health_pb2_grpc

from .config import settings


async def serve_grpc(stop_event: asyncio.Event) -> None:
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=4))
    health_servicer = health.HealthServicer()
    health_pb2_grpc.add_HealthServicer_to_server(health_servicer, server)
    health_servicer.set("", health_pb2.HealthCheckResponse.SERVING)
    server.add_insecure_port(f"[::]:{settings.grpc_port}")
    await server.start()
    await stop_event.wait()
    await server.stop(grace=5)
