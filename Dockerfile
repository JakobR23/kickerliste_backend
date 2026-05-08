FROM golang:1.26.2-alpine AS builder
WORKDIR /app

# Download dependencies first so this layer is cached unless go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd

# ── Final image ──────────────────────────────────────────────────────────────
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
