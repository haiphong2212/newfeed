.PHONY: test migrate proto run-auth run-news compose-config

test:
	go test ./...

migrate:
	go run ./cmd/migrate

proto:
	"C:/Program Files/protoc/bin/protoc.exe" -I api/proto -I services/auth-service/api/proto --go_out=. --go_opt=module=github.com/newfeed/community-news --go-grpc_out=. --go-grpc_opt=module=github.com/newfeed/community-news api/proto/user/v1/user.proto api/proto/news/v1/news.proto api/proto/search/v1/search.proto services/auth-service/api/proto/auth.proto

run-auth:
	go run ./services/auth-service/cmd/auth-service

run-news:
	go run ./services/news-service/cmd/news-service

compose-config:
	docker compose config --quiet
