FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

RUN apk add --no-cache curl ca-certificates && \
    addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -D appuser

WORKDIR /app

COPY --from=builder /app/main ./
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./main"]
