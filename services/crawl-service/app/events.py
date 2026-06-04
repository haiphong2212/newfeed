import json
import aio_pika

from .config import settings


async def publish_event(event_name: str, payload: dict) -> None:
    try:
        conn = await aio_pika.connect_robust(settings.rabbitmq_url)
        async with conn:
            channel = await conn.channel()
            exchange = await channel.declare_exchange("newfeed.events", aio_pika.ExchangeType.TOPIC, durable=True)
            body = json.dumps({"event_name": event_name, "payload": payload}, default=str).encode("utf-8")
            await exchange.publish(aio_pika.Message(body=body, delivery_mode=aio_pika.DeliveryMode.PERSISTENT), routing_key="crawl.post.approved")
    except Exception:
        return
