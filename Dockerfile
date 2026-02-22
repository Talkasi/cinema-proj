FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

COPY . .

RUN swag init --parseDependency --parseInternal -g ./cmd/api/main.go

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api

FROM alpine:latest

RUN apk add --no-cache postgresql-client

WORKDIR /root/

COPY --from=builder /app/bin/api .
COPY --from=builder /app/sql ./sql
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/scripts ./scripts

RUN chmod -R 777 ./scripts/k6/*.js

COPY --from=builder /app/scripts/docker/docker-entrypoint.sh ./docker-entrypoint.sh
RUN chmod +x ./docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./docker-entrypoint.sh"]
