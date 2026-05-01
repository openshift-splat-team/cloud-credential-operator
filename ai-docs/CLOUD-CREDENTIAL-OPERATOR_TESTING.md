# cloud-credential-operator Testing Guide

**Last Updated:** 2026-05-01

---

## Overview

This guide covers all test suites for cloud-credential-operator and how to run them.

**Testing Philosophy:**
- Unit tests for business logic
- Integration tests for component interactions
- E2E tests for critical user workflows
- Aim for 80%+ code coverage

---

## Test Suites

### Unit Tests

**Purpose:** Test individual functions and methods in isolation

**Location:** `pkg/*/` (co-located with source)

**Run:**

```bash
go test ./...
```

**With coverage:**
```bash
go test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out
```

**Example:**
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "test",
            want:  "result",
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

### Integration Tests

**Purpose:** Test interactions between components

**Location:** `test/integration/`

**Run:**
```bash
make test-integration

# Or directly
go test ./test/integration/... -tags=integration
```

**Requirements:**
- May require local Kubernetes cluster (kind, minikube)
- External dependencies (databases, message queues)

**Example:**
```go
// +build integration

func TestControllerReconciliation(t *testing.T) {
    // Setup test cluster
    testEnv := setupTestEnvironment(t)
    defer testEnv.Cleanup()

    // Create test resource
    resource := createTestResource(testEnv)

    // Wait for reconciliation
    eventually(t, func() bool {
        return resource.Status.Ready == true
    }, 30*time.Second)
}
```

---

### E2E Tests

**Purpose:** Test critical user workflows end-to-end

**Location:** `test/e2e/`

**Run:**
```bash
make test-e2e

# Or with specific cluster
export KUBECONFIG=/path/to/kubeconfig
go test ./test/e2e/... -timeout 30m
```

**Requirements:**
- Real or realistic Kubernetes cluster
- Project deployed to cluster
- May require cloud credentials (for cloud-specific features)

**Example:**
```go
func TestUserWorkflow(t *testing.T) {
    // Deploy application
    deployApp(t)

    // Perform user actions
    createResource(t, testResource)

    // Verify expected outcomes
    verifyResourceCreated(t, testResource)
    verifyStatusUpdated(t, testResource)

    // Cleanup
    deleteResource(t, testResource)
}
```

---

## Test Organization

### Table-Driven Tests

Preferred pattern for unit tests:

```go
tests := []struct {
    name    string
    input   InputType
    want    OutputType
    wantErr bool
}{
    {name: "case1", input: ..., want: ...},
    {name: "case2", input: ..., want: ...},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test logic
    })
}
```

### Test Fixtures

Reusable test data:

```go
// test/fixtures/resources.go
func NewTestResource(name string) *MyResource {
    return &MyResource{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: "test",
        },
        Spec: MyResourceSpec{
            // Defaults
        },
    }
}
```

### Test Helpers

Common test utilities:

```go
// test/helpers/assertions.go
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration) {
    t.Helper()
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if condition() {
            return
        }
        time.Sleep(100 * time.Millisecond)
    }
    t.Fatal("condition not met within timeout")
}
```

---

## Mocking

### Interface-Based Mocking

```go
// Define interface
type MyClient interface {
    Get(ctx context.Context, key string) (string, error)
}

// Mock implementation for tests
type mockClient struct {
    getFunc func(ctx context.Context, key string) (string, error)
}

func (m *mockClient) Get(ctx context.Context, key string) (string, error) {
    return m.getFunc(ctx, key)
}

// Use in test
func TestWithMock(t *testing.T) {
    mock := &mockClient{
        getFunc: func(ctx context.Context, key string) (string, error) {
            return "mocked-value", nil
        },
    }

    result := functionUnderTest(mock)
    // Assertions...
}
```

### Using testify/mock (if applicable)

```go
import "github.com/stretchr/testify/mock"

type MockClient struct {
    mock.Mock
}

func (m *MockClient) Get(ctx context.Context, key string) (string, error) {
    args := m.Called(ctx, key)
    return args.String(0), args.Error(1)
}

func TestWithTestify(t *testing.T) {
    mockClient := new(MockClient)
    mockClient.On("Get", mock.Anything, "key").Return("value", nil)

    result := functionUnderTest(mockClient)

    mockClient.AssertExpectations(t)
}
```

---

## Test Coverage

### Generate Coverage Report

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./pkg/...

# View HTML report
go tool cover -html=coverage.out

# View summary
go tool cover -func=coverage.out
```

### Coverage Goals

- **Minimum:** 70% overall coverage
- **Target:** 80%+ overall coverage
- **Critical paths:** 90%+ coverage (controllers, reconcilers, business logic)

### Excluding from Coverage

```go
// This function intentionally not tested
// Coverage: ignore
func helperFunction() {
    // ...
}
```

---

## CI Test Execution

### Prow Jobs

**Pre-submit tests (run on PRs):**
- `pull-ci-cloud-credential-operator-unit` - Unit tests
- `pull-ci-cloud-credential-operator-integration` - Integration tests (if enabled)
- `pull-ci-cloud-credential-operator-e2e-*` - E2E test suites

**Post-submit tests (run on merge):**
- `branch-ci-cloud-credential-operator-unit` - Unit tests
- `branch-ci-cloud-credential-operator-e2e-*` - Full E2E suite

### Debugging CI Failures

1. **Check Prow logs**
   - Find job in PR checks
   - Click "Details" → view logs

2. **Reproduce locally**
   ```bash
   # Match CI environment
   export CI=true
   make test
   ```

3. **Run specific test**
   ```bash
   go test -v -run TestFailingTest ./pkg/...
   ```

---

## Test Best Practices

### DO

✅ Write tests before fixing bugs (TDD for bugs)
✅ Test both success and error paths
✅ Use table-driven tests for multiple scenarios
✅ Mock external dependencies
✅ Keep tests fast (unit tests < 1s, integration < 10s)
✅ Use meaningful test names describing the scenario
✅ Clean up resources in test cleanup functions

### DON'T

❌ Test implementation details (test behavior, not internals)
❌ Write flaky tests (tests that randomly fail)
❌ Skip cleanup (use `t.Cleanup()` or `defer`)
❌ Use sleeps (use eventually/wait helpers instead)
❌ Test third-party code (trust their tests)
❌ Ignore test failures ("it works on my machine")

---

## Test Utilities

### Common Test Helpers

```bash
# Run specific test
go test -run TestName ./pkg/path

# Run tests in specific package
go test ./pkg/controllers/...

# Run tests with race detector
go test -race ./pkg/...

# Run tests with timeout
go test -timeout 5m ./test/e2e/...

# Verbose output
go test -v ./pkg/...

# Run tests matching pattern
go test -run "Test.*Controller" ./pkg/...
```

### Environment Variables

```bash
# Enable debug logging in tests
export LOG_LEVEL=debug

# Use specific kubeconfig for tests
export KUBECONFIG=/path/to/test-cluster-config

# Skip slow tests
export SKIP_SLOW_TESTS=true

# CI mode (stricter timeouts, no interactive)
export CI=true
```

---

## Related Documentation

- [Development Guide](cloud-credential-operator_DEVELOPMENT.md) - Build and development workflow
- [Components](architecture/components.md) - Architecture to understand what to test
- [Team Testing Practices](../../team/ai-docs/practices/testing.md) - Team-wide testing guidelines

---

**Questions?** See test-specific issues in GitHub or ask in team channel.
