# ADR-0001: Actuator Pattern for Multi-Cloud Credential Management

**Date**: 2018-01-01
**Status**: Accepted

## Context

CCO must support credentials provisioning across AWS, Azure, GCP, OpenStack, vSphere, and other platforms with a single controller codebase. Each cloud has a completely different API and credential model (IAM policies, RBAC roles, Service Account keys, application credentials).

The operator also needs to work in disconnected/manual environments where it does not interact with cloud APIs at all.

## Decision

Define an `Actuator` interface that encapsulates all cloud-specific operations (Create, Update, Delete, Exists). The `credentialsrequest` controller is cloud-agnostic; the actuator is injected at startup based on detected infrastructure platform type.

A `DummyActuator` provides the no-op implementation for unsupported or Manual-mode deployments.

## Rationale

The actuator pattern isolates cloud API changes to individual packages (pkg/aws, pkg/azure, etc.) without touching the reconciliation logic. Adding a new cloud provider requires only implementing the Actuator interface and registering it in `pkg/operator/controller.go`.

### Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| Actuator interface (chosen) | Clean separation, testable, extensible | Interface must stay stable as new methods added |
| Switch-in-controller | Simple for few clouds | Controller becomes unmanageable at 8+ providers |
| Separate operators per cloud | Maximum isolation | Maintenance overhead, duplicated logic |

## Consequences

**Positive**: New cloud providers can be added without changing the controller. Unit tests can use mock actuators. Manual mode is trivially a no-op actuator.

**Negative**: Adding capabilities to the interface (e.g., `IsTimedTokenCluster`, `Upgradeable`) requires all actuators to implement the method, including those that don't use the feature (returning sensible defaults).
