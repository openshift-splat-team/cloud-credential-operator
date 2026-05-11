# CredentialsRequest

**API Group**: `cloudcredential.openshift.io/v1`
**Kind**: `CredentialsRequest`
**Scope**: Namespaced (typically `openshift-*` namespaces; CCO watches all namespaces)

## Purpose

Declarative request for cloud credentials. Components create a CredentialsRequest describing what cloud permissions they need; CCO provisions the credentials and stores them in the referenced Secret.

**Key Principle**: Callers declare *what* (permissions), not *how* (IAM user vs role vs OIDC token). CCO chooses the mechanism based on the operator's mode.

## Spec Structure

```go
type CredentialsRequestSpec struct {
    SecretRef          ObjectReference    // Where to store the resulting credentials Secret
    ProviderSpec       *RawExtension      // Cloud-specific permissions (AWSProviderSpec, AzureProviderSpec, etc.)
    ServiceAccountNames []string          // SAs that will use these credentials (for STS/OIDC trust policy)
    CloudTokenPath     string             // JWT mount path for STS/Workload Identity flows
}
```

### Provider Specs

Each cloud provider embeds its own type in `ProviderSpec`:

| Cloud | Type | Key Fields |
|-------|------|------------|
| AWS | `AWSProviderSpec` | `statementEntries` (IAM policy statements) |
| Azure | `AzureProviderSpec` | `roleBindings` (scope + role) |
| GCP | `GCPProviderSpec` | `predefinedRoles`, `skipServiceCheck` |
| OpenStack | `OpenStackProviderSpec` | `secretNamespace`, `secretName` |
| vSphere | `VSphereProviderSpec` | `permissions` |
| KubeVirt | `KubevirtProviderSpec` | (inherits from infra cluster) |
| IBM Cloud | `IBMCloudProviderSpec` | `policies` |
| PowerVS | `PowerVSProviderSpec` | `policies` |
| Nutanix | `NutanixProviderSpec` | `apiAccess` |

## Status Structure

```go
type CredentialsRequestStatus struct {
    Provisioned                          bool        // True once credentials first provisioned
    LastSyncTimestamp                    *Time       // Last reconciliation time
    LastSyncGeneration                   int64       // Generation when last synced
    LastSyncCloudCredsSecretResourceVersion string   // Detect root credential rotation
    LastSyncInfrastructureResourceVersion  string   // Detect infra tag changes
    ProviderStatus                        *RawExtension // Cloud-specific status (e.g., IAM user ARN)
    Conditions                            []CredentialsRequestCondition
}
```

### Condition Types

| Condition | Meaning |
|-----------|---------|
| `InsufficientCloudCreds` | Root credential lacks permissions to mint or passthrough |
| `MissingTargetNamespace` | SecretRef.Namespace doesn't exist yet |
| `CredentialsProvisionFailure` | Cloud API error during provisioning |
| `CredentialsDeprovisionFailure` | Cloud API error during cleanup |
| `Ignored` | ProviderSpec platform doesn't match cluster platform (normal) |
| `StaleCredentials` | CredentialsRequest removed from release image; pending cleanup |
| `OrphanedCloudResource` | Azure App Registration couldn't be deleted during mode transition |

## Key Concepts

### Mode-Driven Behavior

The same CredentialsRequest is handled differently by mode:

- **Mint**: CCO creates a scoped IAM user/service account in the cloud, stores long-term credentials in SecretRef
- **Passthrough**: CCO copies the root credential directly into SecretRef
- **Manual/STS**: CCO does nothing (pre-provisioned by `ccoctl`); operator owns the Secret

### Finalization

`cloudcredential.openshift.io/deprovision` finalizer ensures cloud resources (IAM users, Azure app registrations) are deleted before the CredentialsRequest is removed from etcd.

### Ignored CredentialsRequests

The release image ships CredentialsRequests for *all* clouds. CCO sets `Ignored=True` for providers that don't match the cluster's infrastructure type. This is expected and healthy.

### STS/OIDC Short-Term Tokens

When `spec.cloudTokenPath` is set alongside an STS-capable `providerSpec` field (e.g., `stsIAMRoleARN` for AWS), CCO recognizes this as a short-term token request. `ccoctl` pre-creates the trust relationship; the resulting Secret contains token-based config, not long-term keys.

## Lifecycle

1. **Creation**: Reconciler detects new CR, calls actuator `Create()` → cloud resource provisioned → Secret written → `Provisioned=true`
2. **Update**: Generation change or cloud credential rotation triggers resync; actuator `Update()` called
3. **Deletion**: Finalizer invokes actuator `Delete()` → cloud resource removed → finalizer released

## Example: AWS STS

```yaml
apiVersion: cloudcredential.openshift.io/v1
kind: CredentialsRequest
metadata:
  name: openshift-image-registry
  namespace: openshift-cloud-credential-operator
spec:
  secretRef:
    name: installer-cloud-credentials
    namespace: openshift-image-registry
  cloudTokenPath: /var/run/secrets/openshift/serviceaccount/token
  providerSpec:
    apiVersion: cloudcredential.openshift.io/v1
    kind: AWSProviderSpec
    stsIAMRoleARN: arn:aws:iam::123456789:role/openshift-image-registry
    statementEntries:
    - effect: Allow
      action:
      - s3:GetObject
      - s3:PutObject
      resource: "arn:aws:s3:::*"
```

## Labels and Annotations on Target Secret

CCO adds these to the provisioned Secret:

| Key | Value | Purpose |
|-----|-------|---------|
| `cloudcredential.openshift.io/credentials-request` label | `"true"` | Marks CCO-managed secrets |
| `cloudcredential.openshift.io/credentials-request` annotation | `namespace/name` | Back-pointer to owning CR |
| `cloudcredential.openshift.io/aws-policy-last-applied` | JSON | Avoid unnecessary AWS calls |

## Related Concepts

- [CloudCredentials](./cloudcredentials.md) - Operator config controlling CCO's mode
- [architecture/components.md](../architecture/components.md) - How the actuator pattern works
