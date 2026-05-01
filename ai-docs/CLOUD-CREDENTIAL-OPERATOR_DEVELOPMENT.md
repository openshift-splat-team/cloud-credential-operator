# cloud-credential-operator Development Guide

**Last Updated:** 2026-05-01

---

## Overview

This guide covers the development workflow for cloud-credential-operator.

**Tech Stack:** **Languages:** Go  
**Frameworks:** Kubernetes, controller-runtime  
**Build Systems:** Make, Docker

---

## Prerequisites

**Required:**
- Go 1.21+ (for Go projects) or appropriate language runtime
- Git
- Make
- Docker (for containerized testing)

**Optional:**
- kubectl (for Kubernetes testing)
- podman (alternative to Docker)

---

## Repository Setup

### Clone Repository

```bash
git clone https://github.com/openshift-splat-team/cloud-credential-operator.git
cd cloud-credential-operator
```

### Install Dependencies

```bash
# For Go projects
go mod download
go mod vendor  # if vendoring is used

# For Python projects
pip install -r requirements.txt
pip install -r requirements-dev.txt

# For JavaScript/TypeScript
npm install
```

---

## Building

### Local Build

```bash
# For Go projects
make build

# Or directly
go build -o bin/cloud-credential-operator ./cmd/...
```

### Build Container Image

```bash
make docker-build

# Or with podman
podman build -t cloud-credential-operator:latest .
```

---

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/my-feature
```

### 2. Make Changes

- Follow project coding conventions
- Add/update tests for your changes
- Update documentation as needed

### 3. Run Tests Locally

```bash
# Unit tests
make test

# Integration tests (if applicable)
make test-integration

# E2E tests (if applicable)
make test-e2e
```

### 4. Verify Build

```bash
# Lint
make lint

# Verify formatting
make verify

# Build
make build
```

### 5. Commit Changes

Follow team commit conventions (see `../../team/knowledge/commit-convention.md`).

### 6. Open Pull Request

- Push branch to fork
- Open PR against main branch
- Request review from team
- Address review feedback
- Wait for CI to pass

---

## Running Locally

### As Standalone Binary

```bash
# Build
make build

# Run
./bin/cloud-credential-operator --help
```

### In Kubernetes Cluster

```bash
# Build and push image
make docker-build docker-push

# Deploy to cluster
kubectl apply -f deploy/
```

### With Operator SDK (if applicable)

```bash
# Run locally (watches cluster)
make run
```

---

## Debugging

### Enable Debug Logging

```bash
# Set log level
export LOG_LEVEL=debug

# Or via command line
./bin/cloud-credential-operator --log-level=debug
```

### Attach Debugger (Go)

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug ./cmd/cloud-credential-operator
```

### Common Issues

**Build failures:**
- Check Go version: `go version`
- Verify dependencies: `go mod verify`
- Clean build cache: `go clean -cache`

**Test failures:**
- Check test environment setup
- Review test logs for specific errors
- Run individual test: `go test -v -run TestName ./pkg/...`

---

## Project Structure

```
cloud-credential-operator/
├── cmd/                    # Command-line entry points
├── pkg/                    # Library code
│   ├── controllers/        # Controllers (if operator)
│   ├── api/               # API types and CRDs
│   └── ...
├── config/                # Configuration (CRDs, RBAC, etc.)
├── hack/                  # Build and development scripts
├── test/                  # Test suites
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── docs/                  # Project documentation
├── Makefile              # Build automation
└── go.mod                # Go dependencies
```

See [Components Overview](architecture/components.md) for architectural details.

---

## Code Conventions

### Naming

- **Packages**: lowercase, single word if possible
- **Files**: lowercase with underscores (snake_case)
- **Types**: PascalCase
- **Functions**: camelCase (exported) or PascalCase (unexported)

### Error Handling

- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Return errors, don't panic
- Log errors at appropriate level

### Testing

- Unit tests in same package: `*_test.go`
- Table-driven tests preferred
- Mock external dependencies
- Aim for 80%+ code coverage

---

## Helpful Make Targets

```bash
make help              # Show all targets
make build            # Build binaries
make test             # Run unit tests
make test-integration # Run integration tests
make test-e2e         # Run e2e tests
make lint             # Run linters
make fmt              # Format code
make verify           # Verify formatting and generated files
make docker-build     # Build container image
make deploy           # Deploy to cluster
```

---

## CI/CD

### Prow Jobs (OpenShift)

This project uses OpenShift Prow for CI/CD.

**Pre-submit jobs:**
- `pull-ci-*-unit` - Unit tests
- `pull-ci-*-e2e` - E2E tests
- `pull-ci-*-verify` - Linting and verification

**Post-submit jobs:**
- `branch-ci-*-images` - Build and push images

See `.ci-operator.yaml` and `ci-operator/config/` for Prow configuration.

### GitHub Actions (if applicable)

See `.github/workflows/` for GitHub Actions configuration.

---

## Related Documentation

- [Testing Guide](cloud-credential-operator_TESTING.md) - Test suites and strategies
- [Components](architecture/components.md) - Architecture overview
- [Team Workflows](../../team/ai-docs/workflows/) - Team-level processes

---

**Questions?** See `../../team/HUMAN-REVIEW-GUIDE.md` for how to escalate issues.
