FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY gen ./gen
COPY services ./services
RUN go build -o /out/analytics-service ./services/analytics-service/cmd/analytics-service

FROM alpine:3.20
COPY --from=build /out/analytics-service /usr/local/bin/analytics-service
EXPOSE 8007 50058
ENTRYPOINT ["analytics-service"]
