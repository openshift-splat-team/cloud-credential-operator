# CCO Testing Guide

> **Test pyramid philosophy and E2E framework patterns**: See [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/testing)

## Test Organization

```text
pkg/
└── **/*_test.go              # Unit tests (co-located with source)

test/
├── e2e/
│   ├── aws/sts/              # AWS STS E2E tests (build tag: e2e)
│   └── azure/azident/        # Azure Workload Identity E2E tests (build tag: e2e)
└── extend/                   # OTE extension tests (external test binary)

cmd/cloud-credential-tests-ext/  # E2E test binary entrypoint
```

## Unit Tests

**Location**: `pkg/**/*_test.go`

```bash
make test                           # all unit tests
go test -v ./pkg/aws/...            # AWS actuator tests
go test -v ./pkg/azure/...          # Azure actuator tests
go test -v ./pkg/operator/credentialsrequest/...  # Controller tests
go test -v ./pkg/operator/status/... # Status controller tests
make coverage                       # coverage report via hack/codecov.sh
```

**CCO-Specific Unit Test Patterns**:

- **Actuator mocking**: Tests inject a mock `Actuator` implementation into the credentialsrequest controller; no real cloud calls
- **Fake clients**: Use `sigs.k8s.io/controller-runtime/pkg/client/fake` for Kubernetes API interactions
- **Provider spec serialization**: Provider specs are `*runtime.RawExtension`; tests exercise JSON round-trip via codec
- **Mode-specific paths**: Tests cover Mint, Passthrough, and Manual mode branches in the controller

**Key test files**:
- `pkg/operator/credentialsrequest/credentialsrequest_controller_test.go` — main controller tests
- `pkg/operator/credentialsrequest/credentialsrequest_controller_azure_test.go` — Azure-specific paths
- `pkg/operator/credentialsrequest/credentialsrequest_controller_gcp_test.go` — GCP-specific paths
- `pkg/operator/credentialsrequest/credentialsrequest_controller_vsphere_test.go` — vSphere paths

## E2E Tests

**Prerequisites**: Running OpenShift cluster with appropriate cloud credentials

### AWS STS E2E

```bash
# Requires: AWS cluster with STS configured, KUBECONFIG set
make test-e2e-sts
# Equivalent:
go test -mod=vendor -race -tags e2e ./test/e2e/aws/sts/...
```

**Tests cover**: STS role assumption, short-term token refresh, CredentialsRequest lifecycle in STS mode

### Azure Workload Identity E2E

```bash
# Requires: Azure cluster with Workload Identity configured
make test-e2e-azident
# Equivalent:
go test -mod=vendor -race -tags e2e ./test/e2e/azure/azident/...
```

**Tests cover**: Azure federated identity credential creation, managed identity token exchange

### Extended Tests (OTE)

```bash
# Build the external test binary
make cloud-credential-tests-ext

# Run via openshift-tests (typical CI invocation)
openshift-tests run --provider aws ./cloud-credential-tests-ext
```

## Test Coverage

**Current gaps** (areas needing more coverage):
- IBM Cloud, PowerVS, Nutanix actuators (minimal unit tests)
- `ccoctl` subcommands (mostly manual testing)
- Mode transition scenarios (Mint → Manual)

## Debugging Tests

### Unit Test Failures

```bash
# Run with verbose output
go test -v -run TestMyController ./pkg/operator/credentialsrequest/...

# Check for race conditions
go test -race ./pkg/...
```

### E2E Test Failures

```bash
# Check operator logs
oc logs -f deployment/cloud-credential-operator \
  -n openshift-cloud-credential-operator

# Check CredentialsRequest status
oc describe credentialsrequest -A

# Must-gather for full state
oc adm must-gather -- /usr/bin/gather
```

### Common Failures

| Symptom | Likely Cause |
|---------|-------------|
| `InsufficientCloudCreds` condition | Root credential lacks mint permissions |
| `CredentialsProvisionFailure` | Cloud API error (check actuator logs with `-v=4`) |
| `MissingTargetNamespace` | Target namespace not yet created (retry expected) |
| E2E STS test fails token exchange | OIDC provider thumbprint mismatch or role trust policy issue |

## CI Integration

- Unit tests run on every PR via `make test`
- E2E tests run in cloud-specific CI lanes (`e2e-aws-cco-manual`, etc.)
- `make verify` blocks PRs with out-of-date generated code
