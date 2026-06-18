# ADR-0005: The per-type identity join is the unit of extensibility

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
The headline feature is set-membership classification (managed/unmanaged/missing), which depends entirely on correctly matching a state resource to its corresponding cloud resource. The Terraform `id` is frequently not the cloud's primary identifier: `aws_route53_record` id is `ZONEID_name_type`, `aws_iam_role_policy` id is `role:policy`, `aws_s3_bucket_policy` id is the bucket name, and many other types use composite ids. A naive id intersection produces both false `unmanaged` results (a managed resource looks orphaned) and false `missing` results (the same resource counted twice under different keys), which is the worst possible failure for the lead feature.

## Options considered
- **Name-mapping registry line** — Treat "adding a type" as a single registry entry; looks easy, but is silently wrong wherever the Terraform id diverges from the cloud primary identifier.
- **Explicit IdentityFn + conformance test** — Each type ships an explicit IdentityFn (Terraform id <-> cloud primary identifier) plus a conformance golden test asserting the id round-trips between a tfstate fixture and a cloud fixture; more work per type, but the join is verified.

## Decision
The IdentityFn plus its conformance test is the real unit of work and the gate for promoting any type. We assume identity mapping is mechanical-but-mandatory per type, and we explicitly reject "one registry line" as a goal wherever it would hide the join.

## Consequences
- The set-membership classification is protected at the source, eliminating the false `unmanaged`/`missing` failures that a naive id intersection would produce.
- Adding a type costs more up front: every type must ship a hand-written IdentityFn and a round-tripping conformance test before it can be promoted.
- **Revisit if:** the identity mapping turns out to be more than mechanical for some type (e.g. it cannot be expressed as a deterministic round-trip between tfstate and cloud fixtures), breaking the per-type assumption.
