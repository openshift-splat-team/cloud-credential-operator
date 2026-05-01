# cloud-credential-operator Ecosystem and References

**Last Updated:** 2026-05-01

---

## Purpose

This document provides links to related projects, upstream dependencies, documentation, and external resources relevant to cloud-credential-operator.

---

## Upstream Projects

### Kubernetes

**Relationship:** cloud-credential-operator runs on Kubernetes

**Resources:**
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [API Reference](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Controller Runtime](https://github.com/kubernetes-sigs/controller-runtime) - Framework for building operators

**Version Compatibility:**
- Supported Kubernetes versions: 1.24+
- Controller Runtime version: v0.15.x

---

### OpenShift (if applicable)

**Relationship:** cloud-credential-operator is part of OpenShift platform

**Resources:**
- [OpenShift Documentation](https://docs.openshift.com/)
- [OpenShift Enhancement Proposals](https://github.com/openshift/enhancements)
- [OpenShift CI (Prow)](https://docs.ci.openshift.org/)

**Version Compatibility:**
- Supported OpenShift versions: 4.12+

---

## Related Platform Projects

### Cloud Provider Integrations

**vSphere (VMware):**
- [govmomi](https://github.com/vmware/govmomi) - vSphere API client
- [vSphere CSI Driver](https://github.com/kubernetes-sigs/vsphere-csi-driver)
- [vSphere Cloud Provider](https://github.com/kubernetes/cloud-provider-vsphere)

**AWS:**
- [AWS SDK for Go](https://github.com/aws/aws-sdk-go)
- [AWS Cloud Provider](https://github.com/kubernetes/cloud-provider-aws)

**GCP:**
- [GCP SDK](https://cloud.google.com/go)
- [GCP Cloud Provider](https://github.com/kubernetes/cloud-provider-gcp)

**Azure:**
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)
- [Azure Cloud Provider](https://github.com/kubernetes-sigs/cloud-provider-azure)

---

## Sister Projects

Projects in the same team or ecosystem:

- **[Project 1](https://github.com/org/project1)** - Description
- **[Project 2](https://github.com/org/project2)** - Description
- **[Project 3](https://github.com/org/project3)** - Description

See team repository for full project list: `../../team/ai-docs/architecture/projects.md`

---

## Dependencies

### Direct Dependencies

Key libraries and frameworks used by cloud-credential-operator:

**Go Modules:**
- `k8s.io/client-go` - Kubernetes client
- `sigs.k8s.io/controller-runtime` - Controller framework
- `github.com/spf13/cobra` - CLI framework (if applicable)
- `github.com/prometheus/client_golang` - Metrics

**Python Packages (if applicable):**
- `kubernetes` - Kubernetes client
- `pytest` - Testing framework

**JavaScript/TypeScript (if applicable):**
- `@kubernetes/client-node` - Kubernetes client
- `react` - UI framework

See `go.mod`, `requirements.txt`, or `package.json` for complete dependency list.

### Indirect Dependencies

- Authentication/authorization libraries
- Logging frameworks
- Testing utilities

---

## Standards and Specifications

### Kubernetes Standards

- [Custom Resource Definition (CRD)](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
- [Controller Pattern](https://kubernetes.io/docs/concepts/architecture/controller/)
- [Admission Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)

### Cloud Provider Standards

- **AWS:** [AWS Well-Architected Framework](https://aws.amazon.com/architecture/well-architected/)
- **GCP:** [GCP Architecture Framework](https://cloud.google.com/architecture/framework)
- **Azure:** [Azure Well-Architected Framework](https://docs.microsoft.com/azure/architecture/framework/)
- **vSphere:** [vSphere API Reference](https://developer.vmware.com/apis/vsphere-automation/latest/)

---

## Documentation Resources

### Team-Level Documentation

See team repository for:
- **Workflows:** Sprint process, epic breakdown, triage
- **Practices:** Coding standards, testing guidelines
- **Roles:** Hat-switching, responsibilities

Location: `../../team/ai-docs/`

### Platform Documentation

**Operator Development:**
- [Operator SDK](https://sdk.operatorframework.io/)
- [Operator Best Practices](https://sdk.operatorframework.io/docs/best-practices/)
- [Kubebuilder Book](https://book.kubebuilder.io/)

**Testing:**
- [Kubernetes Testing Guide](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-testing/testing.md)
- [E2E Testing Framework](https://github.com/kubernetes-sigs/e2e-framework)

**CI/CD:**
- [Prow Documentation](https://docs.prow.k8s.io/)
- [OpenShift CI](https://docs.ci.openshift.org/)

---

## Community and Support

### Communication Channels

**Team Channels:**
- Slack: `#team-channel` (internal)
- GitHub Discussions: Project discussions tab
- Mailing List: team-list@example.com (if applicable)

**Upstream Communities:**
- Kubernetes Slack: `#sig-cloud-provider`, `#kubebuilder`, etc.
- OpenShift Slack: `#forum-openshift`, `#forum-<component>`

### Meetings

**Team Meetings:**
- Sprint planning: Bi-weekly (see team calendar)
- Sprint review: Bi-weekly
- Standup: Daily (async)

**Upstream Meetings:**
- SIG meetings: Check [Kubernetes calendar](https://calendar.google.com/calendar/embed?src=calendar%40kubernetes.io)
- OpenShift meetings: Check [OpenShift calendar](https://calendar.google.com/calendar/embed?src=openshift.io_5s2lnu98o7vjhm8hs5q4vkp7s0%40group.calendar.google.com)

---

## Learning Resources

### Getting Started

**Kubernetes:**
- [Kubernetes Basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/)
- [Kubernetes the Hard Way](https://github.com/kelseyhightower/kubernetes-the-hard-way)

**Operator Development:**
- [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- [Operator SDK Tutorial](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/)

**Cloud Providers:**
- [vSphere Docs](https://docs.vmware.com/en/VMware-vSphere/index.html)
- [AWS Documentation](https://docs.aws.amazon.com/)
- [GCP Documentation](https://cloud.google.com/docs)
- [Azure Documentation](https://docs.microsoft.com/azure/)

### Advanced Topics

- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Controller Runtime Deep Dive](https://engineering.bitnami.com/articles/kubebuilder-deep-dive.html)
- [Writing Controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md)

---

## Related Documentation

- [Development Guide](../cloud-credential-operator_DEVELOPMENT.md) - Build and develop
- [Testing Guide](../cloud-credential-operator_TESTING.md) - Test suites
- [Components](../architecture/components.md) - Architecture
- [ADRs](../decisions/) - Architectural decisions

---

**Note:** Update this document as the ecosystem evolves, dependencies change, or new resources become available.
