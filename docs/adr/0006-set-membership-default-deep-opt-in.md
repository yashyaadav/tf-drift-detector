# ADR-0006: Set-membership by default; attribute diff (--deep) opt-in and bespoke-only

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
Drift classification can run at two depths: set-membership (is the resource present on both sides?) and attribute-level diff (do fields differ?). Attribute diffing is a false-positive minefield — ARNs, timestamps, default tags, JSON policy whitespace and ordering, and computed fields all produce spurious diffs. For generic CCAPI types the field vocabulary also differs (CFN PascalCase vs TF snake_case) with no field map, yielding garbage diffs.

## Options considered
- **Always attribute-diff everything** — surfaces field-level changes universally, but is noisy and slow (N+1 Gets) and produces false positives, including unmappable diffs on generic CCAPI types.
- **Set-membership by default, attribute diff behind --deep (bespoke-only)** — cheap and robust, delivers the orphan/unmanaged headline immediately; "changed" detection is opt-in and limited to bespoke scanners where we own the field mapping, at the cost of no field-level detection by default.

## Decision
Default to set-membership classification. Gate attribute diffing behind an opt-in `--deep` flag, and support `--deep` only on bespoke types where we own the CFN<->TF field mapping.

## Consequences
- Cheap, robust default that delivers the orphan/unmanaged headline with few false positives.
- No field-level "changed" detection by default, and `--deep` is unavailable for generic CCAPI types.
- **Revisit if:** a reliable CFN<->TF field map exists for a given type, allowing `--deep` to be safely extended to generic types per-type.
