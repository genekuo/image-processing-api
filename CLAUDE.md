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
- ✅ golangci-lint passing (no linting errors)
- ✅ Trivy/Snyk security scan passing (no high/critical vulnerabilities)
- ✅ Code coverage threshold met (85% minimum)
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
- Run golangci-lint
- Run Trivy security scanning
- Check code coverage threshold
- Build Docker image (test build)

**Continuous Deployment (on git tags):**
- Build multi-architecture Docker image
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

- **Test Coverage**: Minimum 85% for merging
- **Linting**: Zero golangci-lint errors
- **Security**: No high/critical vulnerabilities
- **Error Handling**: All errors must be handled gracefully
- **Logging**: Use structured logging for all operations
- **Documentation**: All public functions must have godoc comments

### 8. Development Workflow Example

```bash
# 1. Start with an issue
# Create GitHub Issue #42: "Add image rotation feature"

# 2. Create feature branch
git checkout -b feature/issue-42-image-rotation

# 3. Implement with tests and documentation
# - Write code
# - Write unit tests
# - Update README.md if needed
# - Update Issue #42 with progress notes

# 4. Commit with conventional format
git add .
git commit -m "feat(image): add rotation operations

- Implement rotate-90, rotate-180, rotate-270
- Add unit tests with 95% coverage
- Update API documentation in README.md

Closes #42"

# 5. Push and create PR
git push -u origin feature/issue-42-image-rotation
gh pr create --title "feat: Add image rotation operations" \
  --body "Implements rotate-90, rotate-180, rotate-270 operations.

Closes #42

## Changes
- Added rotation functions
- 95% test coverage
- Updated README.md API docs

## Testing
All tests pass, coverage above threshold"

# 6. Wait for all CI checks to pass (green)
# 7. Merge PR (auto-closes issue #42)
# 8. Delete feature branch
```

## Project-Specific Rules

### API Constraints
- Max source image size: 50MB
- Max output dimensions: 1400x1400px
- Cache TTL: 5 minutes idle time
- Supported operations: rotate-90, rotate-180, rotate-270, resize-WxH

### Error Handling
- 4xx errors: Orange placeholder with error code
- 5xx errors: Red placeholder with error code
- All placeholders respect requested dimensions

### Monitoring
- Prometheus metrics exposed on `/metrics`
- Health check on `/health`
- Readiness check on `/ready`
- CORS: Allow all origins

## References

- **Project Documentation**: See [README.md](README.md)
- **GitHub Repository**: https://github.com/steviee/github-workflow-article
- **Container Registry**: ghcr.io/steviee/github-workflow-article

---

**Remember**: Issues + README = Source of Truth. Always keep them updated!
- Always use the golang-pro sub-agent for working with Golang code.