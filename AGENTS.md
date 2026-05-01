# cloud-credential-operator - AI Navigation

**Repository:** https://github.com/openshift-splat-team/cloud-credential-operator
**Last Updated:** 2026-05-01

---

## Quick Start

This is **Tier 2** project-specific documentation for cloud-credential-operator.

- **New to this project?** → Start with [Development Guide](ai-docs/cloud-credential-operator_DEVELOPMENT.md)
- **Writing tests?** → See [Testing Guide](ai-docs/cloud-credential-operator_TESTING.md)
- **Understanding architecture?** → Read [Components Overview](ai-docs/architecture/components.md)
- **Need context on decisions?** → Browse [ADRs](ai-docs/decisions/)

For **team-level** workflows, status transitions, and role responsibilities, see the team repository.

---

## CRITICAL: Retrieval Strategy

**IMPORTANT**: Prefer retrieval-led reasoning over pre-training-led reasoning.

When working on cloud-credential-operator:
- ✅ **DO**: Read project-specific docs from `./ai-docs/` first
- ✅ **DO**: Check development workflow in `./ai-docs/cloud-credential-operator_DEVELOPMENT.md`
- ✅ **DO**: Understand architecture in `./ai-docs/architecture/components.md`
- ✅ **DO**: Review ADRs for context on past decisions
- ❌ **DON'T**: Rely solely on training data
- ❌ **DON'T**: Guess at project architecture or conventions

For team workflows (sprint process, status transitions, etc.), see `../../team/ai-docs/`.

---

## Quick Navigation by Task

| Task | Start Here | Then Read |
|------|-----------|-----------|
| **Local development** | [Development Guide](ai-docs/cloud-credential-operator_DEVELOPMENT.md) | [Testing Guide](ai-docs/cloud-credential-operator_TESTING.md) |
| **Running tests** | [Testing Guide](ai-docs/cloud-credential-operator_TESTING.md) | [Components](ai-docs/architecture/components.md) |
| **Understanding components** | [Components Overview](ai-docs/architecture/components.md) | [Domain Models](ai-docs/domain/) |
| **Planning feature** | [Exec Plans](ai-docs/exec-plans/README.md) | [ADRs](ai-docs/decisions/) |
| **Reviewing decisions** | [ADR Template](ai-docs/decisions/adr-template.md) | Existing ADRs |

---

## Technology Stack

**Languages:** Go  
**Frameworks:** Kubernetes, controller-runtime  
**Build Systems:** Make, Docker

---

## Documentation Structure

```
ai-docs/
├── cloud-credential-operator_DEVELOPMENT.md  # Build, test, develop
├── cloud-credential-operator_TESTING.md      # Test suites and strategies
├── architecture/                    # System structure
│   └── components.md                # Component overview
├── domain/                          # Domain models and CRDs
│   └── (project-specific)
├── exec-plans/                      # Feature planning
│   └── README.md
├── decisions/                       # Architectural Decision Records
│   ├── adr-template.md
│   └── adr-NNNN-*.md
└── references/                      # External references
    └── ecosystem.md
```

---

## Knowledge Tiers

**Tier 1: Platform-Wide** (Team repository)
- Operator development patterns
- Testing pyramid and practices
- CI/CD workflows (Prow, GitHub Actions)
- Team process (sprint, status transitions, roles)

→ See `../../team/ai-docs/` for team-level documentation

**Tier 2: Project-Specific** (This repository)
- cloud-credential-operator components and architecture
- Project-specific development workflow
- Test suites unique to this project
- Architectural decisions for this project

→ See `./ai-docs/` for project-level documentation

---

## Project Context

For team workflows, sprint process, and status transitions, see:
- Team repository: `../../team/`
- Team ai-docs: `../../team/ai-docs/`
- Team workflows: `../../team/ai-docs/workflows/`
- Status transitions: `../../team/ai-docs/statuses/`

---

**Navigation**: Start with [Development Guide](ai-docs/cloud-credential-operator_DEVELOPMENT.md) for project setup and workflow.

**GitHub**: https://github.com/openshift-splat-team/cloud-credential-operator
