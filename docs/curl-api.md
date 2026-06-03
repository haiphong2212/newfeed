# NewFeed API Curl

Set gateway URL:

```bash
export API_URL=http://localhost:8000
export ACCESS_TOKEN=replace_with_access_token
```

## User Profile

Upload avatar or cover media:

```bash
curl -X POST "$API_URL/v1/media/upload" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "owner_id=00000000-0000-4000-8000-000000000001" \
  -F "bucket=profile-media" \
  -F "files=@./avatar.png"
```

Get profile:

```bash
curl -X GET "$API_URL/v1/users/00000000-0000-4000-8000-000000000001/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Upsert profile:

```bash
curl -X PUT "$API_URL/v1/users/00000000-0000-4000-8000-000000000001/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Nguyen Van A",
    "bio": "Community technology writer",
    "headline": "Senior Backend Engineer",
    "education": "BSc Computer Science",
    "occupation": "Software Engineer",
    "location": "Ho Chi Minh City",
    "website_url": "https://newfeed.site",
    "avatar_object_id": "00000000-0000-4000-8000-000000000010",
    "cover_object_id": "00000000-0000-4000-8000-000000000011"
  }'
```

Update avatar:

```bash
curl -X PATCH "$API_URL/v1/users/00000000-0000-4000-8000-000000000001/profile/avatar" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"object_id":"00000000-0000-4000-8000-000000000010"}'
```

Update cover image:

```bash
curl -X PATCH "$API_URL/v1/users/00000000-0000-4000-8000-000000000001/profile/cover" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"object_id":"00000000-0000-4000-8000-000000000011"}'
```

## Articles

List published articles by user:

```bash
curl -X GET "$API_URL/v1/users/00000000-0000-4000-8000-000000000001/articles?limit=20" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Create comment:

```bash
curl -X POST "$API_URL/v1/articles/00000000-0000-4000-8000-000000000100/comments" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "00000000-0000-4000-8000-000000000001",
    "body": "Bai viet nay can cap nhat them nguon tham khao."
  }'
```

Reply to comment:

```bash
curl -X POST "$API_URL/v1/articles/00000000-0000-4000-8000-000000000100/comments" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "00000000-0000-4000-8000-000000000001",
    "parent_id": "00000000-0000-4000-8000-000000000200",
    "body": "Dong y, minh bo sung them thong tin."
  }'
```

List comments:

```bash
curl -X GET "$API_URL/v1/articles/00000000-0000-4000-8000-000000000100/comments?limit=50" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Share article to profile:

```bash
curl -X POST "$API_URL/v1/articles/00000000-0000-4000-8000-000000000100/share" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "00000000-0000-4000-8000-000000000001",
    "caption": "Bai nay dang doc cho cong dong backend.",
    "visibility": "public"
  }'
```

List user shares:

```bash
curl -X GET "$API_URL/v1/users/00000000-0000-4000-8000-000000000001/shares?limit=20" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Search

Index article:

```bash
curl -X POST "$API_URL/v1/search/articles" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "article_id": "00000000-0000-4000-8000-000000000100",
    "title": "OpenAI Releases GPT-6",
    "content": "Official community news content",
    "category": "ai",
    "tags": ["openai", "gpt", "ai"]
  }'
```

Search articles:

```bash
curl -X GET "$API_URL/v1/search/articles?q=openai&tag=openai&category=ai&limit=20" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```
