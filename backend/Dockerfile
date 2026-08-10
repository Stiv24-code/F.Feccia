FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

# poppler-utils fornisce pdftoppm/pdftotext, il motore PDF degli endpoint
# /api/v1/pdf/* (internal/services/pdfengine) — senza, quegli endpoint
# rispondono 422 ma il resto del servizio funziona.
RUN apk add --no-cache curl ca-certificates poppler-utils && \
    addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -D appuser

WORKDIR /app

COPY --from=builder /app/main ./
# /data ospita il token Microsoft Graph del mail scraper (DATA_DIR, volume in
# docker-compose): va creato nell'immagine con l'owner giusto, altrimenti il
# named volume nasce di root e appuser non può scriverci.
RUN mkdir -p /data && chown -R appuser:appgroup /app /data

USER appuser

EXPOSE 8080

CMD ["./main"]
