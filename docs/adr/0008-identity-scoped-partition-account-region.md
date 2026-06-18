# ADR-0008: Scope resource identity by (partition, account, region) from day one

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
Set-membership joins must be scoped: the same resource type across multiple accounts or regions would otherwise produce false cross-account orphans and missing resources. Throttle limits are also per-account-per-region. Multi-account is future scope, but `resource.Resource` and the versioned JSON schema are designed now, and retrofitting account/partition later would be a breaking change to the versioned contract in `pkg/schema`.

## Options considered
- **Add account/partition later when multi-account lands** — defers the work, but breaks the versioned schema and forces a future breaking change.
- **Bake Partition + Account + Region into `resource.Resource` and `ScanScope` now** — small upfront cost even though v1 populates a single account, but keeps the versioned contract forward-compatible.

## Decision
Resource identity carries (partition, account, region) from day one, and all set-membership joins are scoped by (partition, account, region), even though v1 populates a single account.

## Consequences
- Forward-compatible versioned schema; multi-account can land without breaking the `pkg/schema` contract.
- Small upfront cost: identity fields are carried and populated before they are strictly needed.
- **Revisit if:** multi-account never materializes and the added identity dimensions impose ongoing cost without benefit, making the forward-compatibility assumption no longer worth it.
