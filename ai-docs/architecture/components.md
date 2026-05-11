# CCO Component Architecture

## Overview

CCO is structured around an **actuator pattern**: a single `credentialsrequest` controller orchestrates lifecycle, while cloud-specific `Actuator` implementations handle cloud API calls.

## Component Map

```text
cmd/
├── cloud-credential-operator/   # Operator binary (controller manager)
└── ccoctl/                      # Off-cluster CLI tool

pkg/
├── apis/cloudcredential/v1/     # CRD types (CredentialsRequest + provider specs)
├── operator/                    # Controller implementations
│   ├── credentialsrequest/      # Main reconciler + Actuator interface
│   ├── secretannotator/         # Root credential mode detection
│   ├── status/                  # ClusterOperator status rollup
│   ├── podidentity/             # Pod identity webhook lifecycle
│   ├── cleanup/                 # Stale CredentialsRequest removal
│   ├── loglevel/                # Dynamic log level from CloudCredentials CR
│   ├── metrics/                 # Prometheus metrics
│   ├── platform/                # Infrastructure platform detection
│   └── constants/               # Shared constants (namespace, annotation keys)
├── aws/                         # AWS actuator (IAM users, STS roles)
├── azure/                       # Azure actuator (App Registrations, Workload Identity)
├── gcp/                         # GCP actuator (Service Accounts, Workload Identity)
├── openstack/                   # OpenStack actuator (application credentials)
├── vsphere/                     # vSphere actuator (passthrough/manual only)
├── kubevirt/                    # KubeVirt actuator (delegates to infra cluster)
├── ovirt/                       # oVirt actuator
├── ibmcloud/                    # IBM Cloud actuator
└── util/                        # Infra detection, scheme registration
```

## Actuator Interface

Every cloud provider implements:

```go
type Actuator interface {
    Create(ctx, *CredentialsRequest) error
    Delete(ctx, *CredentialsRequest) error
    Update(ctx, *CredentialsRequest) error
    Exists(ctx, *CredentialsRequest) (bool, error)
    GetCredentialsRootSecretLocation() types.NamespacedName
    IsTimedTokenCluster(client, ctx, logger) (bool, error)
    Upgradeable(CloudCredentialsMode) *ClusterOperatorStatusCondition
    GetCredentialsRootSecret(ctx, *CredentialsRequest) (*corev1.Secret, error)
}
```

At startup, `pkg/operator/controller.go` detects the infrastructure platform type and initializes the matching actuator. A `DummyActuator` is used for unsupported platforms (no-op).

## Controller Wiring

```
Manager startup
    └── AddToManager()
            ├── metrics.Add()           → Prometheus metrics collector
            ├── secretannotator.Add()   → watches root credential Secret
            ├── podidentity.Add()       → manages pod identity webhook Deployment
            ├── status.Add()            → watches all CredentialsRequests → ClusterOperator
            ├── loglevel.Add()          → watches CloudCredentials for log level changes
            ├── cleanup.Add()           → periodic scan for stale CredentialsRequests
            └── credentialsrequest.AddWithActuator(actuator)
                    └── watches CredentialsRequests, Secrets, CloudCredentials
```

## Dual Manager Architecture

CCO runs **two** controller-runtime managers:
- **`m`** (tenant manager): watches the component cluster (normal operation)
- **`rootM`** (root manager): watches the root cluster where root credentials live

This supports HyperShift hosted clusters where the root credential may be in a different cluster from the workload.

## Data Flow: Mint Mode

```
CredentialsRequest created
    → credentialsrequest controller reconciles
    → actuator.Exists() → false
    → actuator.Create()
        → parse ProviderSpec (e.g., AWSProviderSpec)
        → call cloud API (create IAM user + policy)
        → write credentials into SecretRef
    → update Status.Provisioned = true
    → update Status.ProviderStatus (e.g., IAM user ARN)
```

## Data Flow: STS/OIDC (Manual Mode)

```
ccoctl pre-runs (off-cluster):
    → reads CredentialsRequests from manifests
    → creates OIDC trust relationships in cloud
    → generates Secret manifests with token-based config

Cluster bootstraps:
    → admin applies pre-generated Secret manifests
    → CCO in Manual mode: sees CredentialsRequest, does nothing
    → component pod mounts JWT → exchanges with cloud STS → short-term token
```

## Secret Annotator

Detects what the root credential can do, sets annotation on root Secret:

```
cloudcredential.openshift.io/mode: mint | passthrough | insufficient
```

This annotation drives actuator behavior when `CloudCredentials.spec.credentialsMode` is empty.

## Status Rollup

`status` controller aggregates all CredentialsRequests:
- Any failure condition → `Degraded=True` on ClusterOperator
- All provisioned + no failures → `Available=True`
- Pending provisioning → `Progressing=True`
- Mint mode with sufficient perms → `Upgradeable=True`

## Provider-Specific Notes

| Provider | Special Behavior |
|----------|-----------------|
| AWS | STS role ARN extracted from `spec.providerSpec.stsIAMRoleARN`; `ccoctl aws` subcommands |
| Azure | Mint mode removed 4.16+; Workload Identity via `ccoctl azure` |
| GCP | Workload Identity Pool via `ccoctl gcp`; `skipServiceCheck` for disconnected |
| KubeVirt | Delegates to underlying infra cluster (nested virtualization) |
| vSphere | Passthrough/Manual only (no credential minting API) |
