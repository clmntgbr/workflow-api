# ============================================
# Base stage - Common dependencies
# ============================================
FROM golang:1.25-alpine AS base

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .


# ============================================
# Development stage - With Air hot reload
# ============================================
FROM base AS development

RUN go install github.com/air-verse/air@v1.67.1
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
ENV PATH="/go/bin:${PATH}"

EXPOSE 3000

CMD ["air", "-c", ".air.toml"]


# ============================================
# Builder stage - Build binaries
# ============================================
FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o api \
    ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o worker \
    ./cmd/worker

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o cli \
    ./cmd/cli


# ============================================
# Production stage - Minimal runtime
# ============================================
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

COPY --from=builder --chown=appuser:appuser /app/api .
COPY --from=builder --chown=appuser:appuser /app/worker .
COPY --from=builder --chown=appuser:appuser /app/cli .

USER appuser

EXPOSE 3000

CMD ["./api"]
