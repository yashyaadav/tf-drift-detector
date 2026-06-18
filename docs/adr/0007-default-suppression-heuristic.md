# ADR-0007: Default suppression via an AWS-owned heuristic, not a static skip-list

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
Pure set-membership comparison floods the report with false "unmanaged" findings from AWS-managed, default, and ephemeral resources: the default VPC/SG/route table/NACL, service-linked roles, AWS-managed KMS keys, auto-created buckets, and similar. A clean account must scan to roughly zero unmanaged resources or the tool looks naive. A hardcoded list of specific default ids rots immediately, since default SG ids are per-VPC and AWS keeps adding new managed defaults.

## Options considered
- **Static skip-list of known default resource ids** — simple to ship, but wrong almost immediately because ids are per-VPC and the list cannot keep pace with new managed defaults.
- **Maintained heuristic predicate** — matches ownership signals (service-linked role path `/aws-service-role/`, owner alias `aws`, `aws:` tag prefix, default-VPC/SG detection, AWS-managed key detection); requires upkeep as patterns evolve but generalizes across accounts.

## Decision
Ship a first-class, tested suppression heuristic, asserted by a "clean account -> ~0 unmanaged" test, with a user-overridable `.driftignore` layered on top.

## Consequences
- A clean account scans to roughly zero unmanaged, so findings stay credible and signal-rich.
- We accept the ongoing maintenance cost of curating the heuristic as AWS evolves its managed-default patterns.
- **Revisit if:** AWS introduces new managed-default patterns that the ownership signals (paths, aliases, tag prefixes) no longer capture, breaking the assumption that ownership signals generalize better than enumerated ids.
