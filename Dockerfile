# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy dependency manifests first for layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and compile a static binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/server .

# Stage 2: Runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /build/server /usr/local/bin/server

EXPOSE 8080

USER appuser

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD ["wget", "-qO-", "http://localhost:8080/health"]

ENTRYPOINT ["server"]
