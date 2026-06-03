FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY gen ./gen
COPY services ./services
RUN go build -o /out/search-service ./services/search-service/cmd/search-service

FROM alpine:3.20
COPY --from=build /out/search-service /usr/local/bin/search-service
EXPOSE 8004 50056
ENTRYPOINT ["search-service"]
