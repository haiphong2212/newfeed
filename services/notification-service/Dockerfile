FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY services ./services
RUN go build -o /out/notification-service ./services/notification-service/cmd/notification-service

FROM alpine:3.20
COPY --from=build /out/notification-service /usr/local/bin/notification-service
EXPOSE 8005 50055
ENTRYPOINT ["notification-service"]
