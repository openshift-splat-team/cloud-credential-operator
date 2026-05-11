# Cloud Credential Operator - Agentic Documentation

**Component**: Cloud Credential Operator (CCO)
**Repository**: openshift/cloud-credential-operator
**Documentation Tier**: 2 (Component-specific)

> **Generic Platform Patterns**: See [Tier 1 Ecosystem Hub](https://github.com/openshift/enhancements/tree/master/ai-docs) for operator patterns, testing practices, security guidelines, and cross-repo ADRs.

> **Retrieval-first**: Read `ai-docs/domain/credentialsrequest.md` before answering questions about CredentialsRequest fields or CCO modes.

## What is CCO?

Manages cloud provider credentials for OpenShift components. Operators request credentials via `CredentialsRequest` CRs; CCO provisions the resulting `Secret` based on the cluster's mode (Mint, Passthrough, or Manual/STS).

**Key Principle**: Every component declaring what cloud access it needs — CCO fulfills it without callers knowing the underlying credential mechanism.

## Core Components

- **credentialsrequest controller**: Reconciles CredentialsRequests → cloud credentials + Secrets
- **secretannotator**: Detects root credential permissions, sets mode annotation on root Secret
- **status controller**: Rolls up CredentialsRequest conditions → ClusterOperator status
- **podidentity**: Deploys pod identity webhook for short-term token (STS/Workload Identity) flows
- **cleanup**: Removes stale CredentialsRequests no longer in the release image
- **ccoctl**: Off-cluster CLI for Manual/STS credential pre-provisioning

**Quick Start**: `oc describe clusteroperator/cloud-credential-operator` | `oc get credentialsrequests -A`

## Documentation Structure

```text
ai-docs/
├── domain/                        # CCO-specific CRDs
│   ├── credentialsrequest.md      # Primary CRD (all fields, modes, conditions)
│   └── cloudcredentials.md        # Operator config CRD (mode selection)
├── architecture/
│   └── components.md              # Actuator pattern, controller wiring, provider map
├── decisions/                     # CCO-specific ADRs
├── exec-plans/active/             # Active feature implementation plans
├── references/
│   └── ecosystem.md               # Links to Tier 1 patterns
├── CCO_DEVELOPMENT.md             # Build, dev workflow, code organization
└── CCO_TESTING.md                 # Unit, integration, E2E test suites
```

**Exec-Plans**: Use `active/` for new features. See [Tier 1 Guide](https://github.com/openshift/enhancements/tree/master/ai-docs/workflows/exec-plans).

**Platform Patterns (Tier 1)**: [Operator](https://github.com/openshift/enhancements/tree/master/ai-docs/platform/operator-patterns) | [Testing](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/testing) | [Security](https://github.com/openshift/enhancements/tree/master/ai-docs/practices/security)

## Knowledge Graph

```text
                        [AGENTS.md] ← Start here
                              │
           ┌──────────────────┼──────────────────┐
           │                  │                  │
      [domain/]        [architecture/]      [decisions/]
  CredentialsRequest   Actuator pattern     Mode choices
  CloudCredentials     Controller wiring    Provider ADRs
           │                  │                  │
           └──────────────────┼──────────────────┘
                              │
                    [references/ecosystem]
                      Links to Tier 1
```

**AI Agent Path**: domain/ → architecture/ → decisions/ → CCO_DEVELOPMENT.md

## Operator Modes (CCO-Specific)

| Mode | Behavior | Credential Type |
|------|----------|----------------|
| `Mint` | CCO creates scoped IAM user/SA per request | Long-term keys |
| `Passthrough` | CCO copies root credential to each Secret | Long-term keys |
| `Manual` | Admin/ccoctl pre-provisions; CCO does nothing | Any |
| Manual + STS | ccoctl wires OIDC trust; pod uses JWT → short-term token | Short-term |

## Commit Format

```
<subsystem>: <description>

Assisted-by: <AI Model Name>
```

**Subsystems**: `ccoctl`, `aws`, `azure`, `gcp`, `openstack`, `vsphere`, `kubevirt`, `ibmcloud`, `powervs`, `operator`

## External References

[CCO Docs](https://docs.openshift.com/container-platform/latest/authentication/managing_cloud_provider_credentials/about-cloud-credential-operator.html) | [STS Flow](docs/sts.md) | [ccoctl Guide](docs/ccoctl.md) | [Adding Cloud Provider](docs/adding-new-cloud-provider.md)

---

**Tier 1 Hub**: https://github.com/openshift/enhancements/tree/master/ai-docs
