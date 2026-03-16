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

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /build/server /usr/local/bin/server

EXPOSE 8080

USER appuser

ENTRYPOINT ["server"]
