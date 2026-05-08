# syntax=docker/dockerfile:1.7
# ─── Build stage ─────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Install build deps used by some Go modules (git for VCS info, ca-certs).
RUN apk add --no-cache git ca-certificates

# Cache dependencies.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy sources and build a static binary.
COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" \
        -o /out/gateway ./cmd

# ─── Runtime stage ───────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /app
COPY --from=builder /out/gateway /app/gateway

# Public proxy + admin API.
EXPOSE 8080 9090

USER nonroot:nonroot
ENTRYPOINT ["/app/gateway"]
