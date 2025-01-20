# build stage
FROM golang:1.23-alpine as builder

RUN mkdir -p /app/

WORKDIR /app

# install build dependencies
RUN apk add --no-cache git

# install goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY .. .

RUN go mod download

# build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/doublee ./cmd/doubleed

# final stage
FROM alpine:3.20

LABEL maintainer="j0lvera"

WORKDIR /app

# install bash for entrypoint script
RUN apk add --no-cache bash

# create a non-root user
RUN addgroup -S app && adduser -S -G app app

# Copy goose, binary, migrations and entrypoint
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/doublee .
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations
COPY docker/entrypoint.sh .
RUN chmod +x entrypoint.sh

RUN ls -la /app/internal/db/migrations

EXPOSE 8080

USER app

ENTRYPOINT ["./entrypoint.sh", "./doublee"]