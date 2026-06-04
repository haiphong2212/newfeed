# Event Contracts

Exchange: `newfeed.events`

Routing keys:

- `article.published` -> `ArticlePublished`
- `user.mentioned` -> `UserMentioned`
- `comment.created` -> `CommentCreated`
- `follow.topic.created` -> `FollowTopicCreated`

Consumer rules:

- Every event must include `event_id`, `event_name`, `occurred_at`, and `payload`.
- Consumers must write `event_id` to `processed_events` before applying side effects or in the same transaction as the side effect.
- Failed deliveries are retried with exponential backoff.
- Poison messages go to service-specific dead-letter queues named `<service>.dlq`.

Payload fields are owned by the publishing bounded context. Consumers must reject payloads that do not include the required business identifiers for the event they handle.
