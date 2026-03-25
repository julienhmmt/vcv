FROM golang:1.26.1-alpine AS builder

ARG VERSION=${VERSION}

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy only necessary source files (exclude img/, docs/, etc.)
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY web/ ./web/

# Build the application with version info
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-w -s \
    -X vcv/internal/version.Version=${VERSION}" \
    -o /vcv-backend ./cmd/server

# Final stage
FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /vcv-backend /app/vcv-backend

EXPOSE 52000

CMD ["/app/vcv-backend"]