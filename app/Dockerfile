FROM golang:1.25.4-alpine3.22 AS builder

ARG VERSION=${VERSION}

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy backend source code
COPY . .

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