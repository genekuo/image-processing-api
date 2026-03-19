# Claude AI Assistant - Project Workflow Rules

This document defines the mandatory workflow rules for AI assistants working on this project.

## Core Principles

**GitHub Issues and README.md are the SINGLE SOURCE OF TRUTH for project progress and documentation.**

All work must maintain complete traceability through issues and comprehensive documentation.

## Mandatory Workflow

### 1. Feature Branch & Pull Request Workflow

- **NEVER commit directly to `main`**
- **ALWAYS create feature branches** for any work: `feature/issue-N-description`
- **ALWAYS work through Pull Requests** for merging to main
- Branch naming convention: `feature/issue-<number>-<short-description>`
  - Example: `feature/issue-5-http-server-setup`

### 2. Documentation Requirements

**Documentation must be updated with EVERY commit:**

- **Issue Descriptions**: Update the relevant GitHub Issue with implementation notes, decisions, and progress
- **README.md**: Update if any user-facing functionality, API, deployment, or architecture changes
- **Code Comments**: Add clear comments for complex logic
- **Commit Messages**: Follow conventional commits format:
  ```
  type(scope): description

  - Detailed explanation if needed
  - Reference: #issue-number
  ```

### 3. Pull Request Requirements

A PR can ONLY be merged when ALL checks are GREEN:

- ✅ All unit tests passing
- ✅ All integration tests passing
- ✅ JaCoCo coverage threshold met (85% minimum instruction coverage)
- ✅ SpotBugs passing (no bugs)
- ✅ Trivy/Snyk security scan passing (no high/critical vulnerabilities)
- ✅ Documentation updated (README.md and Issue descriptions)

### 4. GitHub Issues Management

- **Create issues FIRST** before implementing features
- **Reference README.md** in issue descriptions for global context
- **Use labels** for categorization:
  - `infrastructure`, `core-api`, `error-handling`, `observability`, `quality`, `documentation`
  - `priority: critical`, `priority: high`, `priority: medium`, `priority: low`
- **Document dependencies** in issue descriptions:
  - "Depends on: #X" for blocking dependencies
  - "Blocks: #Y" for issues that depend on this one
- **Update issues** as you work:
  - Add implementation notes
  - Document decisions made
  - Link to related PRs
- **Close issues** via PR commit messages: `Closes #N`

### 5. CI/CD Pipeline

**Continuous Integration (on every PR):**
- Run all tests (unit + integration)
- JaCoCo coverage threshold check (85%)
- SpotBugs static analysis
- Trivy security scanning
- Build Docker image (test build)

**Continuous Deployment (on git tags):**
- Build multi-architecture Docker image (linux/amd64, linux/arm64)
- Push to GitHub Container Registry (ghcr.io)
- Tag with semantic version: `X.Y.Z`, `X.Y`, `X`, `latest`

### 6. Versioning Strategy

**Semantic Versioning (SemVer):**
- Format: `vMAJOR.MINOR.PATCH` (e.g., `v1.2.3`)
- MAJOR: Breaking API changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes (backward compatible)

**Release Process:**
1. Ensure all PRs merged to main
2. Update README.md with release notes
3. Create annotated git tag: `git tag -a v1.0.0 -m "Release description"`
4. Push tag: `git push origin v1.0.0`
5. CD pipeline automatically builds and publishes to ghcr.io

### 7. Code Quality Standards

- **Test Coverage**: Minimum 85% instruction coverage for merging (enforced by JaCoCo)
- **Static Analysis**: Zero SpotBugs errors
- **Security**: No high/critical vulnerabilities (Trivy)
- **Error Handling**: All errors must be handled gracefully; return PNG placeholders on failure
- **Logging**: Use SLF4J structured logging for all operations
- **Reactive**: Blocking operations (ImageIO, Thumbnailator) must run on `Schedulers.boundedElastic()`

### 8. Development Workflow Example

```bash
# 1. Start with an issue
# Create GitHub Issue #5: "Add image rotation feature"

# 2. Create feature branch
git checkout -b feature/issue-5-image-rotation

# 3. Implement with tests and documentation
# - Write code in src/main/java/...
# - Write tests in src/test/java/...
# - Update README.md if needed
# - Update Issue #5 with progress notes

# 4. Commit with conventional format
git add .
git commit -m "feat(image): add rotation operations

- Implement rotate-90, rotate-180, rotate-270 via AffineTransform
- Add unit tests with >90% coverage
- Update API documentation in README.md

Closes #5"

# 5. Push and create PR
git push -u origin feature/issue-5-image-rotation
gh pr create --title "feat: Add image rotation operations" \
  --body "Implements rotate-90, rotate-180, rotate-270 operations.

Closes #5

## Changes
- Added rotation functions in ImageProcessorService
- Tests with >90% coverage
- Updated README.md API docs

## Testing
All tests pass, JaCoCo coverage above 85% threshold"

# 6. Wait for all CI checks to pass (green)
# 7. Merge PR (auto-closes issue #5)
# 8. Delete feature branch
```

## Project-Specific Rules

### Technology Stack
- **Runtime**: Java 21 (LTS), Spring Boot 3.x
- **HTTP**: Spring WebFlux (reactive, non-blocking)
- **Image processing**: Thumbnailator (resize/crop), Java AWT (rotation, placeholder)
- **Cache**: Caffeine with `expireAfterAccess` (idle TTL)
- **Metrics**: Micrometer + Prometheus
- **Testing**: JUnit 5, AssertJ, Mockito, Reactor Test, MockWebServer

### API Constraints
- Max source image size: 50 MB
- Max output dimensions: 1400×1400 px
- Cache TTL: 5 minutes idle time (configurable)
- Supported operations: `rotate-90`, `rotate-180`, `rotate-270`, `resize-WxH`
- Allowed URL schemes: `http`, `https` only

### Error Handling
- 4xx errors: Orange placeholder (#FF8C00) with error code
- 5xx errors: Red placeholder (#DC143C) with error code
- Other/unknown: Gray placeholder (#808080)
- All placeholders respect requested dimensions (from last `resize` op)

### Monitoring
- Prometheus metrics at `/metrics`
- Liveness at `/health`
- Readiness at `/ready`
- CORS: Allow all origins

### Blocking I/O Rule
All blocking operations (ImageIO, Thumbnailator, stream reads) **must** be wrapped in
`Mono.fromCallable(...).subscribeOn(Schedulers.boundedElastic())` to avoid blocking
the Netty event loop.

## References

- **Project Documentation**: See [README.md](README.md)
- **GitHub Repository**: https://github.com/steviee/spring-image-processing
- **Container Registry**: ghcr.io/steviee/spring-image-processing

---

**Remember**: Issues + README = Source of Truth. Always keep them updated!
