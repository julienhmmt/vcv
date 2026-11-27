# syntax=docker/dockerfile:1.9
FROM node:26-alpine@sha256:6bd9ab43c0ec0190ca7b2d193906cb8b7edf6e7f4afa43ddeccd6452710bffb5 AS web-builder

WORKDIR /web/frontend
ENV CI=true

RUN npm install -g pnpm@10

# Cache pnpm install
COPY web/frontend/package.json web/frontend/pnpm-lock.yaml web/frontend/.npmrc web/frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

# Copy the rest of the frontend sources
COPY web/frontend/ ./

# Build to /web/dist (vite.config.ts outputs to ../dist)
RUN pnpm build

FROM golang:1.27.4-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS builder

ARG VERSION=${VERSION}

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY web/web.go ./web/web.go

# Pull the compiled Svelte bundle into the Go embed target.
COPY --from=web-builder /web/dist ./web/dist

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-w -s \
    -X vcv/internal/version.Version=${VERSION}" \
    -o /vcv-backend ./cmd/server

# Create config directory in builder stage (has shell)
RUN mkdir -p /app/config

FROM gcr.io/distroless/static-debian13:nonroot@sha256:aafe993665bf6200db38a02c11a9351e782a84b6a45e7f33b6695638c88e3931
WORKDIR /app
USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /vcv-backend /app/vcv-backend
COPY --from=builder --chown=nonroot:nonroot /app/config /app/config

EXPOSE 52000

CMD ["/app/vcv-backend"]
