FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY gen ./gen
COPY services ./services
RUN go build -o /out/chat-service ./services/chat-service/cmd/chat-service

FROM alpine:3.20
COPY --from=build /out/chat-service /usr/local/bin/chat-service
EXPOSE 8006 50054
ENTRYPOINT ["chat-service"]
