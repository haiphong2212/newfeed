FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY gen ./gen
COPY services ./services
RUN go build -o /out/user-service ./services/user-service/cmd/user-service

FROM alpine:3.20
COPY --from=build /out/user-service /usr/local/bin/user-service
EXPOSE 8002 50052
ENTRYPOINT ["user-service"]
