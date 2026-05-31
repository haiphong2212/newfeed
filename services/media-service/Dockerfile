FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY services ./services
RUN go build -o /out/media-service ./services/media-service/cmd/media-service

FROM alpine:3.20
COPY --from=build /out/media-service /usr/local/bin/media-service
EXPOSE 8008 50057
ENTRYPOINT ["media-service"]
