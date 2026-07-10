# syntax=docker/dockerfile:1.7
FROM node:24-alpine AS web-builder

WORKDIR /web/frontend
ENV CI=true

RUN corepack enable

# Cache pnpm install
COPY web/frontend/package.json web/frontend/pnpm-lock.yaml web/frontend/.npmrc web/frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

# Copy the rest of the frontend sources
COPY web/frontend/ ./

# Build to /web/dist (vite.config.ts outputs to ../dist)
RUN pnpm build

FROM golang:1.26.4-alpine AS builder

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

FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /vcv-backend /app/vcv-backend
COPY --from=builder --chown=nonroot:nonroot /app/config /app/config

EXPOSE 52000

CMD ["/app/vcv-backend"]
