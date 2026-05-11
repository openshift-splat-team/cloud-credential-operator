# ADR-0003: Dual Controller-Runtime Manager for HyperShift

**Date**: 2022-03-01
**Status**: Accepted

## Context

HyperShift separates the control plane (management cluster) from the data plane (hosted cluster). In a hosted cluster, the root cloud credential (the admin credential used to mint or inspect permissions) lives in the **management cluster**, while the CredentialsRequests and workload Secrets live in the **hosted cluster**.

A single controller-runtime manager can only watch one cluster, making the original single-manager architecture incompatible with HyperShift.

## Decision

Run two controller-runtime managers simultaneously:
- **`m`** (tenant manager): connected to the hosted/component cluster; watches CredentialsRequests and writes Secrets there
- **`rootM`** (root manager): connected to the management cluster; reads the root credential Secret

Actuators receive both clients and use them appropriately.

## Rationale

Reusing the existing actuator interface with an additional `rootM` client required minimal changes to each actuator (pass through the root client when reading the root credential). The alternative of a single manager with dual kubeconfigs would require more invasive changes to controller-runtime internals.

### Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| Dual managers (chosen) | Minimal actuator changes, clean separation | Startup complexity, two informer caches |
| Single manager + multi-cluster client | Single manager lifecycle | Requires custom client factory, less idiomatic |
| Separate CCO instances per cluster | Maximum isolation | Operational complexity, no shared state |

## Consequences

**Positive**: Enables CCO to operate in HyperShift hosted clusters. Each manager has independent leader election and health checks.

**Negative**: Startup sequence must ensure both managers are ready before reconciliation begins. Debugging requires checking logs on both the management and hosted cluster sides.
