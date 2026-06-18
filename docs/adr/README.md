# Architecture Decision Records

This directory records significant architectural decisions for tf-drift-detector using the MADR-lite format. Decisions are append-only: they are superseded rather than deleted, and each new decision adds a new numbered file.

| ADR | Title | Status |
| --- | --- | --- |
| [0000](record-architecture-decisions.md) | Adopt ADRs | Accepted |
| [0001](0001-agentless-detection-model.md) | Agentless detection model (parse state + read cloud directly) | Accepted |
| [0002](0002-aws-only-v1-provider-interface.md) | AWS-only v1 behind a Provider interface | Accepted |
| [0003](0003-canonical-resource-model.md) | Two suppliers converge on one canonical resource model | Accepted |
| [0004](0004-scanner-registry-bespoke-primary.md) | Scanner registry: bespoke SDK scanners primary, Cloud Control experimental | Accepted |
| [0005](0005-per-type-identity-join.md) | The per-type identity join is the unit of extensibility | Accepted |
| [0006](0006-set-membership-default-deep-opt-in.md) | Set-membership by default; attribute diff (--deep) opt-in and bespoke-only | Accepted |
| [0007](0007-default-suppression-heuristic.md) | Default suppression via an AWS-owned heuristic, not a static skip-list | Accepted |
| [0008](0008-identity-scoped-partition-account-region.md) | Scope resource identity by (partition, account, region) from day one | Accepted |
| [0009](0009-partial-failure-scanstatus.md) | Partial-failure isolation with a per-scope ScanStatus | Accepted |
| [0010](0010-in-process-throttling.md) | Own throttling in-process; SDK adaptive retry only as a backstop | Accepted |

Further decisions (e.g. versioned JSON output contract, CI-gate exit codes, tfstate parsing library choice, persistence, scheduling, detection-only stance) will be recorded as ADRs 0011+ as those milestones land.
