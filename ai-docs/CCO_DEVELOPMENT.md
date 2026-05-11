# CCO Development Guide

> **Generic Go standards and controller-runtime patterns**: See [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns)

## Quick Start

**Prerequisites**: Go 1.21+, `oc` CLI, access to an OpenShift cluster (for E2E)

```bash
make build          # compile operator + ccoctl binaries
make test           # unit tests
make verify         # generated code + formatting check
make update         # regenerate deepcopy, bindata, CRD schema
```

## Repository Structure

```text
cmd/
├── cloud-credential-operator/  # Operator entrypoint
└── ccoctl/                     # Off-cluster CLI entrypoint

pkg/
├── apis/cloudcredential/v1/    # CRD type definitions (edit here, then make update)
│   ├── types_credentialsrequest.go
│   ├── types_aws.go            # AWSProviderSpec
│   ├── types_azure.go          # AzureProviderSpec
│   ├── types_gcp.go            # GCPProviderSpec
│   └── types_*.go              # Per-provider specs
├── operator/
│   ├── controller.go           # Manager setup + actuator selection
│   ├── credentialsrequest/     # Main reconciler
│   │   ├── credentialsrequest_controller.go
│   │   └── actuator/actuator.go   # Actuator interface
│   ├── status/                 # ClusterOperator status
│   ├── secretannotator/        # Root credential mode detection
│   ├── podidentity/            # Pod identity webhook
│   ├── cleanup/                # Stale CR removal
│   ├── metrics/                # Prometheus metrics
│   └── platform/               # Infrastructure detection
├── aws/                        # AWS actuator + ccoctl AWS commands
├── azure/                      # Azure actuator + ccoctl Azure commands
├── gcp/                        # GCP actuator + ccoctl GCP commands
└── util/                       # Infra helpers, scheme registration

bindata/                        # Static assets embedded in binary (default CRs)
manifests/                      # Kubernetes manifests for deploying CCO
test/e2e/                       # E2E tests (build-tag gated)
```

## Development Workflow

### Local Build

```bash
make build
# Produces: ./cloud-credential-operator (operator) and ./ccoctl (CLI)
```

### Run Unit Tests

```bash
make test
# Or for a specific package:
go test -v ./pkg/aws/...
go test -v ./pkg/operator/credentialsrequest/...
```

### Verify Generated Code

After modifying types in `pkg/apis/cloudcredential/v1/`:

```bash
make update       # regenerates deepcopy, bindata, updates CRD manifests
make verify       # fails if generated code is out of sync
```

### On-Cluster Testing

```bash
# Build and push image
make images IMAGE_REGISTRY=quay.io/myuser

# Replace the running operator (dev override)
oc set image deployment/cloud-credential-operator \
  cloud-credential-operator=quay.io/myuser/cloud-credential-operator:latest \
  -n openshift-cloud-credential-operator

# Watch logs
oc logs -f deployment/cloud-credential-operator \
  -n openshift-cloud-credential-operator
```

### Debug Credentials Issues

```bash
# Check operator mode
oc get cloudcredentials cluster -o jsonpath='{.spec.credentialsMode}'

# Check all CredentialsRequests and their status
oc get credentialsrequests -A
oc describe credentialsrequest <name> -n <namespace>

# Check root credential annotation
oc get secret <root-secret> -n kube-system \
  -o jsonpath='{.metadata.annotations}'
```

## Code Organization

### Adding a New Cloud Provider

See [docs/adding-new-cloud-provider.md](../../docs/adding-new-cloud-provider.md).

Key steps:
1. Add `types_<cloud>.go` in `pkg/apis/cloudcredential/v1/` with ProviderSpec + ProviderStatus
2. Create `pkg/<cloud>/` package implementing `actuator.Actuator`
3. Register in `pkg/operator/controller.go` switch statement
4. Add `ccoctl <cloud>` subcommands if supporting Manual/STS mode
5. Run `make update` to regenerate deepcopy

### Adding a New Controller

1. Create `pkg/operator/<name>/` package with `Add(manager.Manager, ...) error`
2. Register in `pkg/operator/controller.go` `init()`:
   ```go
   AddToManagerFuncs = append(AddToManagerFuncs, mycontroller.Add)
   ```

### Modifying CredentialsRequest Types

1. Edit `pkg/apis/cloudcredential/v1/types_credentialsrequest.go`
2. Run `make update` (regenerates zz_generated.deepcopy.go)
3. Update CRD: `make update-vendored-crds` if CRD is from openshift/api

## Common Tasks

| Task | Command |
|------|---------|
| Update dependencies | `go mod tidy && go mod vendor` |
| Update direct deps | `make update-go-modules-direct` |
| Update indirect deps | `make update-go-modules-indirect` |
| Regenerate bindata | `make update` |
| Run coverage report | `make coverage` (outputs to hack/codecov.sh) |

## Component-Specific Notes

- **Bindata**: Default CredentialsRequests (for GCP read-only, AWS IAM read-only) are embedded via `bindata/`. Edit source files in `bindata/bootstrap/`, then `make update` regenerates `pkg/assets/`.
- **Vendor**: All dependencies are vendored. Use `go mod vendor` after dependency changes.
- **Build tags**: E2E tests use `//go:build e2e` tag. Unit tests have no tags.
- **Log levels**: Controlled dynamically via `CloudCredentials.spec.logLevel`. Use `log.WithField("cr", cr.Name)` pattern from logrus.
