# cloud-credential-operator Components

**Last Updated:** 2026-05-01

---

## Overview

This document describes the major components and architecture of cloud-credential-operator.

**Tech Stack:** **Languages:** Go  
**Frameworks:** Kubernetes, controller-runtime  
**Build Systems:** Make, Docker

---

## High-Level Architecture

```
┌─────────────────────────────────────────────┐
│           cloud-credential-operator                   │
│                                             │
│  ┌──────────────┐      ┌─────────────────┐ │
│  │              │      │                 │ │
│  │  Component A │─────▶│   Component B   │ │
│  │              │      │                 │ │
│  └──────────────┘      └─────────────────┘ │
│                                             │
└─────────────────────────────────────────────┘
```

*(Replace with project-specific architecture diagram)*

---

## Core Components

### Component 1: [Name]

**Purpose:** Brief description of what this component does

**Location:** `pkg/component1/`

**Responsibilities:**
- Responsibility 1
- Responsibility 2
- Responsibility 3

**Key Types:**
- `Type1` - Description
- `Type2` - Description

**Interactions:**
- Calls Component 2 for X
- Listens to events from Y
- Stores data in Z

**Example Usage:**
```go
// Code example showing how this component is used
```

---

### Component 2: [Name]

**Purpose:** Brief description

**Location:** `pkg/component2/`

**Responsibilities:**
- Responsibility 1
- Responsibility 2

**Key Types:**
- `Type1` - Description

**Interactions:**
- Interacts with Component 1
- Calls external service X

---

## For Operator Projects

### Controllers

**Purpose:** Reconcile Kubernetes resources

**Location:** `pkg/controllers/`

- **aws** - `pkg/operator/secretannotator/aws`
- **azure** - `pkg/operator/secretannotator/azure`
- **gcp** - `pkg/operator/secretannotator/gcp`
- **openstack** - `pkg/operator/secretannotator/openstack`
- **vsphere** - `pkg/operator/secretannotator/vsphere`
- **operator** - `pkg/operator`
- **loglevel** - `pkg/operator/loglevel`

**Reconciliation Pattern:**
1. Fetch resource from Kubernetes API
2. Validate resource spec
3. Create/update dependent resources
4. Update resource status
5. Requeue if needed

**Example Reconciliation:**
```go
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Fetch the resource
    obj := &v1alpha1.MyResource{}
    if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Reconciliation logic here

    // Update status
    if err := r.Status().Update(ctx, obj); err != nil {
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}
```

---

### Custom Resource Definitions (CRDs)

See [Domain Models](../domain/) for detailed CRD specifications.

**Defined CRDs:**

*(No CRDs detected)*

---

## API Layer

**Purpose:** Define interfaces and types

**Location:** `pkg/api/` or `api/`

**Key Types:**
- Request/Response structures
- Configuration types
- Status types

---

## Data Flow

```
User/Client
    ↓
API Server
    ↓
Controller/Handler
    ↓
Business Logic
    ↓
External Systems
```

**Example Flow:**
1. User creates CustomResource
2. Controller watches for changes
3. Controller validates resource
4. Controller calls cloud provider API
5. Controller updates resource status

---

## External Dependencies

### Kubernetes API

**Usage:** CRUD operations on Kubernetes resources

**Authentication:** Service account with appropriate RBAC

### Cloud Provider APIs (if applicable)

**AWS:**
- SDK: `aws-sdk-go`
- Services: EC2, IAM, S3, etc.

**GCP:**
- SDK: `cloud.google.com/go`
- Services: Compute, IAM, Storage, etc.

**Azure:**
- SDK: `github.com/Azure/azure-sdk-for-go`
- Services: Compute, Network, Storage, etc.

**vSphere:**
- SDK: `github.com/vmware/govmomi`
- APIs: vCenter, ESXi

### Other Dependencies

- **Database:** PostgreSQL, Redis, etc.
- **Message Queue:** RabbitMQ, Kafka, etc.
- **Cache:** Redis, Memcached, etc.

---

## Configuration

### Config Locations

- **In-cluster:** ConfigMaps, Secrets
- **Command-line:** Flags passed to binary
- **Environment:** Environment variables
- **Files:** Config files mounted to container

### Config Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. ConfigMap/Secret values
4. Default values (lowest priority)

---

## Observability

### Logging

**Framework:** klog, logrus, or standard log

**Log Levels:**
- `ERROR` - Errors that need attention
- `WARN` - Warnings that may need attention
- `INFO` - Informational messages
- `DEBUG` - Verbose debugging

**Structured Logging:**
```go
log.Info("resource reconciled",
    "name", resource.Name,
    "namespace", resource.Namespace,
    "generation", resource.Generation)
```

### Metrics

**Framework:** Prometheus client

**Key Metrics:**
- `reconcile_duration_seconds` - Time to reconcile resources
- `reconcile_errors_total` - Count of reconciliation errors
- `resource_count` - Number of managed resources

**Metrics Endpoint:** `/metrics`

### Tracing (if applicable)

**Framework:** OpenTelemetry

**Traced Operations:**
- API calls
- Controller reconciliation
- External service calls

---

## Error Handling

### Error Types

```go
type CustomError struct {
    Code    string
    Message string
    Cause   error
}
```

### Retry Logic

- **Transient errors:** Retry with exponential backoff
- **Permanent errors:** Don't retry, update status with error
- **Rate limits:** Respect retry-after headers

### Error Propagation

- Wrap errors with context
- Preserve original error for debugging
- Log errors at appropriate level

---

## Security Considerations

### Authentication

- Service account tokens for in-cluster communication
- API keys for external services
- Certificate-based auth where applicable

### Authorization

- RBAC for Kubernetes resources
- Principle of least privilege
- Separate service accounts per component

### Secrets Management

- Store secrets in Kubernetes Secrets
- Never log secret values
- Rotate credentials regularly

---

## Performance Considerations

### Caching

- Cache frequently accessed data
- Invalidate cache on updates
- Use TTL for time-sensitive data

### Rate Limiting

- Respect API rate limits
- Implement client-side rate limiting
- Use backoff for retries

### Resource Limits

- Set appropriate CPU/memory limits
- Monitor resource usage
- Scale based on load

---

## Related Documentation

- [Development Guide](../cloud-credential-operator_DEVELOPMENT.md) - How to build and run
- [Testing Guide](../cloud-credential-operator_TESTING.md) - How to test components
- [Domain Models](../domain/) - CRD specifications
- [ADRs](../decisions/) - Architectural decisions

---

**Note:** This is a template. Update with project-specific component details, architecture diagrams, and actual code examples.
