# ADR-0003: Two suppliers converge on one canonical resource model

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
The desired side (tfstate JSON) and the actual side (cloud SDK / Cloud Control structs) have completely different shapes, yet one generic diff engine must compare them. We need a representation that lets a single engine reason about both without special-casing every provider type. We assume a single canonical container can be expressive enough for set-membership and, for bespoke types, attribute diffing.

## Options considered
- **Single canonical model** — Reduce both sides to one `resource.Resource{Type, ID, Address, ARN, Partition, Account, Region, Source, Attributes}` and run one engine over it (driftctl's two-supplier model); trade-off is the upfront normalization layer each supplier must implement.
- **Per-type bespoke comparison** — Compare raw structs with hand-written code per type, no shared model; trade-off is far more code and no single engine.

## Decision
Both suppliers normalize into the canonical `resource.Resource`, and the engine only ever sees that type.

## Consequences
- One generic diff engine works across both desired and actual sides, regardless of their native shapes.
- Each supplier accepts the cost of a normalization layer that maps its raw structs into the canonical container.
- The container is shared, but per-type field contents still differ (see ADR-0005, see ADR-0006).
- **Revisit if:** the canonical container stops being expressive enough for set-membership or per-type attribute diffing.
