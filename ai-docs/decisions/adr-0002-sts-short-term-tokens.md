# ADR-0002: STS/OIDC Short-Term Token Architecture

**Date**: 2021-06-01
**Status**: Accepted

## Context

Long-term cloud credentials (IAM users with access keys, service account JSON keys) are a security liability: keys don't expire, rotation is manual, and credential leakage has large blast radius. Cloud providers increasingly offer token-based authentication (AWS STS, Azure Workload Identity, GCP Workload Identity Federation) where a Kubernetes Service Account JWT is exchanged for a short-lived cloud token.

CCO needed a way to support these flows without requiring CCO itself to have cloud API access (supporting air-gapped and restricted environments).

## Decision

Introduce `ccoctl` as an **off-cluster CLI tool** that:
1. Reads CredentialsRequests from manifests
2. Creates OIDC provider configuration in the cloud
3. Establishes trust relationships (IAM role trust policy, Azure federated identity, GCP Workload Identity Pool)
4. Emits pre-generated Secret manifests containing token-based config

CCO in `Manual` mode applies no changes; components use the pre-generated Secrets with a mounted ServiceAccount JWT to obtain short-term tokens at runtime.

`spec.cloudTokenPath` on CredentialsRequest signals that the requester expects a token-based Secret.

## Rationale

Separating the provisioning step (ccoctl, offline) from runtime (the pod itself exchanges tokens) means CCO never needs cloud credentials in STS mode. This enables air-gapped installs and satisfies security requirements that mandate no long-term credentials in the cluster.

### Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| ccoctl + Manual mode (chosen) | No cloud creds in cluster, air-gap friendly | Requires pre-install step; more complex for users |
| CCO creates STS roles online | Single tool, no pre-install | CCO needs elevated permissions at runtime |
| Cloud-native pod identity (IRSA only) | AWS-native | Not portable across clouds |

## Consequences

**Positive**: No long-term credentials needed in the cluster at runtime. Satisfies FedRAMP and similar compliance requirements. Enables air-gapped installs.

**Negative**: Requires a pre-install step with `ccoctl`. Debugging is harder when the JWT exchange fails (cloud-side configuration vs. cluster-side token mounting). Rotation requires re-running `ccoctl` and reapplying Secrets.

## References

- [STS Documentation](../../docs/sts.md)
- [Azure Workload Identity](../../docs/azure_workload_identity.md)
- [GCP Workload Identity](../../docs/gcp_workload_identity.md)
