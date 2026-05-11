# CCO Ecosystem References

Links to Tier 1 platform documentation. All generic patterns live there; this file is CCO's index into that hub.

> **Tier 1 Hub**: https://github.com/openshift/enhancements/tree/master/ai-docs

## Operator Patterns (Tier 1)

| Pattern | Link | CCO Usage |
|---------|------|-----------|
| Controller-runtime reconcile loop | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | credentialsrequest controller |
| Status conditions semantics | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns/status-conditions.md) | CredentialsRequestConditionType |
| Finalizer pattern | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | `cloudcredential.openshift.io/deprovision` |
| ClusterOperator status | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | status controller |
| Webhook patterns | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | pod identity webhook |

## Testing Practices (Tier 1)

| Practice | Link | CCO Usage |
|----------|------|-----------|
| Test pyramid (60/30/10) | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/testing) | Unit: pkg/*_test.go, E2E: test/e2e/ |
| Mock vs real strategies | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/testing) | Actuator mock interface in unit tests |
| E2E framework patterns | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/testing) | test/e2e uses Go test with build tags |

## Security Practices (Tier 1)

| Practice | Link | CCO Relevance |
|----------|------|---------------|
| Threat modeling (STRIDE) | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/security) | Credential leakage, privilege escalation |
| RBAC guidelines | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/security) | CCO needs broad Secret write access |
| Secrets management | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/security) | CCO is the secrets manager for cloud credentials |

## Reliability Practices (Tier 1)

| Practice | Link | CCO Usage |
|----------|------|-----------|
| SLO framework | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/reliability) | Prometheus metrics in pkg/operator/metrics |
| Degraded state handling | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/reliability) | status controller → ClusterOperator Degraded |

## Kubernetes / OpenShift Fundamentals (Tier 1)

| Concept | Link | CCO Usage |
|---------|------|-----------|
| ClusterOperator | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/domain/openshift) | CCO reports health via `cloud-credential-operator` CO |
| Secret | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/domain/kubernetes) | Target of CredentialsRequest.spec.secretRef |
| ServiceAccount | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/domain/kubernetes) | Referenced in spec.serviceAccountNames for STS |
| OIDC / ServiceAccount token | [Tier 1](https://github.com/openshift/enhancements/tree/master/ai-docs/domain/kubernetes) | JWT mounted at spec.cloudTokenPath |

## Cross-Repo ADRs (Tier 1)

| ADR | Relevance |
|-----|-----------|
| CVO operator lifecycle | CCO is managed by CVO; CredentialsRequests sourced from release image |
| Release image bootstrapping | CCO's default CredentialsRequests are embedded in the release image |

## CCO-Specific External Docs

- [STS Flow Diagram](../../docs/sts.md) — AWS STS setup and token exchange
- [Azure Workload Identity](../../docs/azure_workload_identity.md)
- [GCP Workload Identity](../../docs/gcp_workload_identity.md)
- [ccoctl Reference](../../docs/ccoctl.md) — Off-cluster credential provisioning CLI
- [Adding a Cloud Provider](../../docs/adding-new-cloud-provider.md)
- [Metrics Reference](../../docs/metrics.md)
