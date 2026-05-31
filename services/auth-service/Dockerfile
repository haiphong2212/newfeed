FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY services ./services
RUN go build -o /out/auth-service ./services/auth-service/cmd/auth-service

FROM alpine:3.20
COPY --from=build /out/auth-service /usr/local/bin/auth-service
EXPOSE 8001 50051
ENTRYPOINT ["auth-service"]
