FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o org-api ./cmd/api

FROM alpine:latest

RUN apk add --no-cache ca-certificates && adduser -D appuser
WORKDIR /app
COPY --from=builder /app/org-api .
COPY migrations ./migrations

USER appuser
EXPOSE 8080
CMD ["./org-api"]
