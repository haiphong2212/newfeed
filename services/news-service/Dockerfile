FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY gen ./gen
COPY services ./services
RUN go build -o /out/news-service ./services/news-service/cmd/news-service

FROM alpine:3.20
COPY --from=build /out/news-service /usr/local/bin/news-service
EXPOSE 8003 50053
ENTRYPOINT ["news-service"]
