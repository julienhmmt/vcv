# syntax=docker/dockerfile:1.9
FROM node:24-alpine@sha256:a34ab6d7f6f4813f4d0c5469d74685604b962fbdbf5793a7bc7b4998ecb867c8 AS web-builder

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

FROM golang:1.27.3-alpine@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 AS builder

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

FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329
WORKDIR /app
USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /vcv-backend /app/vcv-backend
COPY --from=builder --chown=nonroot:nonroot /app/config /app/config

EXPOSE 52000

CMD ["/app/vcv-backend"]
