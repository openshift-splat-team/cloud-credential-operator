# CloudCredentials

**API Group**: `operator.openshift.io/v1`
**Kind**: `CloudCredentials`
**Scope**: Cluster (singleton: `cluster`)

## Purpose

Operator configuration resource controlling CCO's credentials management mode cluster-wide. Set by the installer at install time; can be changed post-install with restrictions.

**Key Principle**: One CR controls the behavior of all CredentialsRequests across the entire cluster.

## Spec Structure

```go
type CloudCredentialsSpec struct {
    ManagementState ManagementState       // Managed | Unmanaged | Removed
    CredentialsMode CloudCredentialsMode  // "" | Mint | Passthrough | Manual
    LogLevel        LogLevel             // Normal | Debug | Trace | TraceAll
    OperatorLogLevel LogLevel
}
```

### CredentialsMode Values

| Mode | Behavior | Use Case |
|------|----------|----------|
| `""` (empty) | CCO auto-detects best mode based on root credential permissions | Default install |
| `Mint` | CCO mints scoped credentials per CredentialsRequest | Mutable root credential with admin permissions |
| `Passthrough` | CCO copies root credential to each requester's Secret | Limited-permission root credential |
| `Manual` | CCO does not manage credentials; operator or admin pre-provisions | Disconnected, STS, Workload Identity |

## Key Concepts

### Auto-Detection (Empty Mode)

When `credentialsMode` is empty, `secretannotator` controller probes the root credential secret for cloud permissions and sets the `cloudcredential.openshift.io/mode` annotation. The `credentialsrequest` controller reads this annotation to determine actual behavior.

### Mode Constraints

- **Mint → Manual**: Supported with `ccoctl` pre-provisioning
- **Passthrough → Manual**: Supported
- **Manual → Mint/Passthrough**: Not supported post-install (root credential may not have mint permissions)
- Azure: Mint mode removed in 4.16+ (Azure Workload Identity is the default STS path)

### ManagementState

`Removed` causes CCO to delete all managed credentials and the operator itself uninstalls. Used when migrating fully to manual/STS.

## Lifecycle

1. **Creation**: Installer sets mode based on install-config and detected cloud capabilities
2. **Update**: `oc patch cloudcredentials cluster --patch '{"spec":{"credentialsMode":"Manual"}}' --type merge`
3. **Never deleted** in normal operation (cluster singleton)

## Example

```yaml
apiVersion: operator.openshift.io/v1
kind: CloudCredentials
metadata:
  name: cluster
spec:
  managementState: Managed
  credentialsMode: Manual
```

## Related Concepts

- [CredentialsRequest](./credentialsrequest.md) - Per-component credential request
- [architecture/components.md](../architecture/components.md) - secretannotator controller details
