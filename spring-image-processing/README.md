# Spring Image Processing API

A reactive image processing HTTP API built with **Spring Boot WebFlux**. Downloads a source image from a remote URL, applies one or more transformations, caches the result, and returns a PNG response.

## Features

- **Reactive stack**: Spring WebFlux + Project Reactor — non-blocking I/O throughout
- **Operations**: `rotate-90`, `rotate-180`, `rotate-270`, `resize-WxH` (cover-crop)
- **Caching**: Caffeine in-memory cache with configurable idle TTL
- **Metrics**: Prometheus counters and timers via Micrometer at `/metrics`
- **Health probes**: `/health` and `/ready`
- **CORS**: All origins allowed (`GET`, `OPTIONS`)
- **Error placeholders**: Color-coded PNG returned on error (orange=4xx, red=5xx)
- **Security**: Source size cap (50 MB), max output 1400×1400 px

## API

### `GET /image`

| Parameter | Required | Description |
|-----------|----------|-------------|
| `url`     | Yes      | Source image URL (`http`/`https`) |
| `op`      | Yes      | Comma-separated operations, e.g. `rotate-90,resize-400x300` |

**Operations:**

| Format | Description |
|--------|-------------|
| `rotate-90` / `rotate-180` / `rotate-270` | Clockwise rotation |
| `resize-WxH` | Scale + center-crop to W×H (max 1400×1400) |

**Responses:**

| Condition | Status | Body |
|-----------|--------|------|
| Success | 200 | PNG image |
| Missing/invalid parameter | 400 | Orange PNG placeholder |
| Download failure | 502 | Orange PNG placeholder |
| Internal error | 500 | Red PNG placeholder |

### `GET /health`

Returns `{"status":"ok"}` — liveness probe.

### `GET /ready`

Returns `{"status":"ready"}` — readiness probe.

### `GET /metrics`

Prometheus metrics endpoint.

## Configuration

`src/main/resources/application.yml`:

```yaml
app:
  max-source-size: 52428800   # 50 MB max source image download
  max-output-width: 1400      # max output width in pixels
  max-output-height: 1400     # max output height in pixels
  cache-ttl: PT5M             # idle TTL (ISO-8601 duration)
```

## Running Locally

### With Maven

```bash
./mvnw spring-boot:run
```

### With Docker Compose

```bash
docker compose up --build
```

API is available at `http://localhost:8080`.

## Building

```bash
# Run tests + quality gates (JaCoCo ≥85%, SpotBugs)
./mvnw verify

# Build JAR only
./mvnw package -DskipTests

# Build Docker image
docker build -t spring-image-processing .
```

## Testing

```bash
# All tests with coverage report
./mvnw verify

# Coverage report at:
# target/site/jacoco/index.html
```

Test suite covers:

| Class | Tests |
|-------|-------|
| `ImageProcessorServiceTest` | 25 — parse, rotate, resize, applyAll |
| `ImageCacheServiceTest` | 9 — TTL, eviction, concurrency |
| `ImageDownloaderServiceTest` | 10 — JPEG/PNG/GIF, size limit, errors, timeout |
| `PlaceholderGeneratorTest` | 11 — colors, dimensions, PNG validity |
| `ImageControllerTest` | 14 — WebTestClient integration tests |
| `ImageProcessingApplicationTests` | 1 — context loads |

## CI/CD

| Trigger | Pipeline |
|---------|----------|
| Every push / PR | Tests, JaCoCo ≥85%, SpotBugs, Docker build, Trivy scan |
| Git tag `vX.Y.Z` | Multi-arch Docker image pushed to GHCR |

Container images are published to `ghcr.io/steviee/spring-image-processing`.

## Architecture

```
WebFlux Router
    └── ImageController
            ├── ImageDownloaderService  (WebClient, non-blocking)
            ├── ImageProcessorService   (Thumbnailator, boundedElastic)
            ├── ImageCacheService       (Caffeine)
            └── PlaceholderGenerator   (AWT Graphics2D)
```

## License

MIT
