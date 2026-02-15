# ═══════════════════════════════════════════════════════════
# Stage 1: build
# ═══════════════════════════════════════════════════════════
ARG GO_VERSION=1.25-alpine
FROM golang:${GO_VERSION} AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# ── Кэш зависимостей ────────────────────────────────────
COPY go.mod go.sum ./
RUN go mod download

# ── Сборка ───────────────────────────────────────────────
COPY . .

ARG APP_VERSION=dev
ARG BUILD_TIME=""
ARG COMMIT_SHA=""

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
        -X main.version=${APP_VERSION} \
        -X main.buildTime=${BUILD_TIME} \
        -X main.commit=${COMMIT_SHA}" \
    -o /bin/app ./cmd/app

# ═══════════════════════════════════════════════════════════
# Stage 2: runtime
# ═══════════════════════════════════════════════════════════
FROM gcr.io/distroless/static-debian12

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /bin/app /app
COPY config/ /config/

ENV APP_ENV=production

ENTRYPOINT ["/app"]
CMD ["serve"]