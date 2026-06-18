# ADR-0009: Partial-failure isolation with a per-scope ScanStatus

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
Enumerating a real account makes thousands of calls, and throttling and AccessDenied are normal conditions rather than exceptions. A plain errgroup cancels all goroutines on the first error, so a single throttled or denied (account, region, type) would abort the whole scan and discard partial results, turning a transient error into a total CI-gate failure. Worse, treating AccessDenied as "no resources" would mis-report a permissions gap as a clean account or as false missing/unmanaged findings, and inflate the coverage metric.

## Options considered
- **Cancel-on-first-error fan-out** — simple errgroup-style concurrency, but brittle: one throttled or denied scope aborts the entire scan and throws away everything already collected.
- **Per-scope failure isolation** — collect partial results plus a per-(partition, account, region, type) ScanStatus enum {Scanned, Throttled, Denied, Skipped, Unsupported} that flows into the report and is excluded from the coverage denominator; more bookkeeping in exchange for resilience and honesty.

## Decision
Isolate failures per scope and surface ScanStatus in pkg/schema, collecting partial results alongside a per-(partition, account, region, type) status. AccessDenied is never counted as zero resources or as fully covered, and unscanned scopes are excluded from the coverage denominator.

## Consequences
- Transient throttling or a missing permission degrades coverage gracefully instead of failing the entire scan, so partial, honestly-annotated results still reach the report and the CI gate.
- We accept added bookkeeping: every scope must carry and propagate a ScanStatus, and consumers must interpret it correctly when computing coverage.
- **Revisit if:** consumers come to require all-or-nothing completeness, such that partial results are no longer more useful than a clean failure — at which point the assumption behind per-scope isolation no longer holds.
