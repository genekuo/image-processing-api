# Image Processing API

A stateless, Dockerized REST API written in Go that acts as an image proxy with on-the-fly transformations. It downloads images from any public URL, applies requested operations (rotate, resize with cover/crop), converts to PNG, and returns the result — with intelligent in-memory caching.

## Features

- **Image proxy**: Fetch any publicly accessible image by URL
- **Format conversion**: Automatically converts source images (JPEG, GIF, WebP, BMP, TIFF) to PNG
- **Rotate**: 90°, 180°, 270° clockwise
- **Resize with cover/crop**: Scale to fill the target dimensions while preserving aspect ratio, then center-crop to exact size
- **Operation chaining**: Apply multiple operations in sequence (e.g., rotate then resize)
- **In-memory cache**: Processed images are cached with a 5-minute idle TTL (resets on access)
- **Error placeholders**: On failure, returns a PNG placeholder in the requested dimensions with the HTTP error code (orange for 4xx, red for 5xx)
- **Observability**: Prometheus metrics, health and readiness endpoints
- **Dockerized**: Fully self-contained, no external storage required

## API

### Process Image

```
GET /image?url=<source_url>&op=<operations>
```

**Parameters:**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `url`     | Yes      | Publicly accessible image URL |
| `op`      | Yes      | Comma-separated list of operations |

**Supported Operations:**

| Operation         | Description |
|-------------------|-------------|
| `rotate-90`       | Rotate 90° clockwise |
| `rotate-180`      | Rotate 180° |
| `rotate-270`      | Rotate 270° clockwise |
| `resize-WxH`      | Resize to WxH using cover/crop (e.g., `resize-800x600`) |

**Examples:**

```bash
# Rotate an image 90°
curl "http://localhost:8080/image?url=https://picsum.photos/800/600&op=rotate-90" -o rotated.png

# Resize to 400x300 (cover/crop)
curl "http://localhost:8080/image?url=https://picsum.photos/800/600&op=resize-400x300" -o resized.png

# Chain operations: rotate then resize
curl "http://localhost:8080/image?url=https://example.com/photo.jpg&op=rotate-180,resize-1200x800" -o result.png
```

**Response:**

- `Content-Type: image/png`
- The processed image as a PNG binary

**Error Responses:**

On error, the API returns a PNG placeholder image instead of JSON:

- **4xx errors**: Orange placeholder with the error code displayed
- **5xx errors**: Red placeholder with the error code displayed
- Placeholders respect the requested dimensions (from `resize-WxH` if present, otherwise a default size)

### Health & Readiness

```bash
GET /health    # Liveness probe — always 200 if the process is running
GET /ready     # Readiness probe — 200 when the service is ready to accept traffic
```

### Metrics

```bash
GET /metrics   # Prometheus-compatible metrics endpoint
```

**Exposed metrics include:**

- `http_requests_total` — total requests by method, path, status
- `http_request_duration_seconds` — request duration histogram
- `image_cache_hits_total` / `image_cache_misses_total` — cache effectiveness
- `image_processing_duration_seconds` — image operation latency
- `image_cache_entries` — current number of cached entries

## Constraints

| Constraint | Value |
|------------|-------|
| Max source image size | 50 MB |
| Max output dimensions | 1400 × 1400 px |
| Cache TTL | 5 minutes idle time |
| Supported source formats | JPEG, PNG, GIF, WebP, BMP, TIFF |
| Output format | Always PNG |

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT`   | `8080`  | HTTP listen port |
| `MAX_SOURCE_SIZE` | `52428800` | Max source image size in bytes (50 MB) |
| `MAX_OUTPUT_WIDTH` | `1400` | Max output width in pixels |
| `MAX_OUTPUT_HEIGHT` | `1400` | Max output height in pixels |
| `CACHE_TTL` | `5m` | Cache idle TTL duration |

## Architecture

```
┌─────────────────────────────────────────────────┐
│                   HTTP Server                    │
│  ┌──────────┐  ┌───────────┐  ┌──────────────┐  │
│  │  Router   │→│ Middleware │→│   Handlers    │  │
│  │ /image    │  │  CORS     │  │  ImageHandler │  │
│  │ /health   │  │  Metrics  │  │  HealthHandler│  │
│  │ /ready    │  │  Logging  │  │              │  │
│  │ /metrics  │  │           │  │              │  │
│  └──────────┘  └───────────┘  └──────┬───────┘  │
│                                      │           │
│  ┌───────────────────────────────────┘           │
│  │                                               │
│  ▼                                               │
│  ┌──────────┐  ┌───────────┐  ┌──────────────┐  │
│  │  Cache    │→│ Downloader│→│  Processor    │  │
│  │ (in-mem)  │  │ (HTTP GET)│  │ rotate/resize│  │
│  │ 5min TTL  │  │ ≤50MB     │  │ chain ops    │  │
│  └──────────┘  └───────────┘  └──────────────┘  │
│                                                   │
│  ┌──────────────────────────────────────────────┐ │
│  │         Error Placeholder Generator          │ │
│  │  4xx → orange │ 5xx → red │ shows error code │ │
│  └──────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
```

## Project Structure

```
image-processing-api/
├── main.go                 # Application entry point
├── go.mod / go.sum
├── Dockerfile
├── docker-compose.yml
├── CLAUDE.md               # AI assistant workflow rules
├── README.md               # This file
├── .github/
│   └── workflows/
│       └── ci.yml          # CI/CD pipeline
└── internal/
    ├── server/
    │   ├── server.go       # HTTP server setup
    │   ├── routes.go       # Route definitions
    │   └── middleware.go   # CORS, logging, metrics middleware
    ├── handler/
    │   └── image.go        # Image request handler
    ├── service/
    │   ├── downloader.go   # Image download + format detection
    │   └── processor.go    # Image operations (rotate, resize)
    ├── cache/
    │   └── cache.go        # In-memory cache with idle TTL
    ├── placeholder/
    │   └── placeholder.go  # Error placeholder image generator
    └── config/
        └── config.go       # Environment-based configuration
```

## Running

### Docker (recommended)

```bash
docker build -t image-processing-api .
docker run -p 8080:8080 image-processing-api
```

### Docker Compose

```bash
docker-compose up
```

### Local Development

```bash
go run main.go
```

## Development

### Prerequisites

- Go 1.25+
- Docker (for containerized builds)
- `gh` CLI (for GitHub workflow)

### Running Tests

```bash
go test ./... -v -race -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Linting

```bash
golangci-lint run ./...
```

## GitHub Workflow

This project follows a strict GitHub Issues + PR workflow. See [CLAUDE.md](CLAUDE.md) for the complete workflow rules.

- All work is tracked via GitHub Issues
- Feature branches: `feature/issue-N-description`
- All merges via Pull Requests with passing CI checks
- Conventional commit messages referencing issue numbers

## License

MIT
