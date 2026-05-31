FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY services ./services
RUN go build -o /out/api-gateway ./services/api-gateway/cmd/api-gateway

FROM alpine:3.20
COPY --from=build /out/api-gateway /usr/local/bin/api-gateway
EXPOSE 8000
ENTRYPOINT ["api-gateway"]
