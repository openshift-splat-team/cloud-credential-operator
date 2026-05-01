# cloud-credential-operator Domain Models

**Last Updated:** 2026-05-01

---

## Overview

This directory documents the domain models, custom resource definitions (CRDs), and core data structures used in cloud-credential-operator.

---

## Custom Resource Definitions (CRDs)

For Kubernetes operator projects, document each CRD here.

**Example structure for each CRD:**

### ResourceName

- **API Group:** `example.com/v1alpha1`
- **Kind:** `ResourceName`
- **Plural:** `resourcenames`
- **Scope:** Namespaced | Cluster

**Purpose:** What this resource represents

**Spec Fields:**
- `field1` (string, required) - Description
- `field2` (int, optional) - Description

**Status Fields:**
- `conditions` ([]Condition) - Resource conditions
- `phase` (string) - Current phase (Pending, Ready, Error)

**Example:**
```yaml
apiVersion: example.com/v1alpha1
kind: ResourceName
metadata:
  name: example
  namespace: default
spec:
  field1: "value"
  field2: 42
status:
  phase: Ready
  conditions:
    - type: Ready
      status: "True"
      reason: ReconciliationSucceeded
```

**Validation:**
- Field1 must match pattern `^[a-z0-9-]+$`
- Field2 must be between 1-100

**Related Documentation:**
- Controller reconciliation logic: [../architecture/components.md](../architecture/components.md)
- API reference: See generated API docs

---

## Core Data Structures

For non-operator projects, document key data structures.

### Structure 1

**Purpose:** Description

**Fields:**
```go
type MyStruct struct {
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
}
```

**Validation Rules:**
- Field1: required, non-empty
- Field2: must be positive

---

## API Versioning

**Current Version:** v1alpha1

**Versioning Policy:**
- `v1alpha1` - Initial experimental API
- `v1beta1` - API stabilizing, may have breaking changes
- `v1` - Stable API, backward compatibility guaranteed

**Deprecated Fields:**
- (None currently)

**Migration Guides:**
- [v1alpha1 → v1beta1](migrations/v1alpha1-to-v1beta1.md) (if applicable)

---

## Related Documentation

- [Components Overview](../architecture/components.md) - How these models are used
- [Development Guide](../cloud-credential-operator_DEVELOPMENT.md) - Adding new fields
- Generated API docs - Full API reference

---

**Note:** For each major CRD or domain model, create a dedicated file (e.g., `credentialsrequest.md`) with detailed specification.
