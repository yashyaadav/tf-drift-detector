# ADR-0010: Own throttling in-process; SDK adaptive retry only as a backstop

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
Orphan detection requires a List call for every scanned type in every region, and `--deep` adds N+1 GetResource calls, creating a real throttling cliff against Cloud Control and per-service APIs that have low, per-service TPS. AWS SDK adaptive retry is a backstop, not a strategy: under sustained throttling it merely slows to a crawl and can still exhaust attempts. We therefore need to bound call volume proactively rather than react to throttle responses after the fact.

## Options considered
- **Rely solely on SDK retry** — use `NewStandard`/`NewAdaptiveMode` for pacing; simplest, but it is reactive and degrades to a crawl or attempt exhaustion under sustained throttling.
- **Own pacing in-process** — a client-side token-bucket rate limiter keyed per (service, region) plus bounded concurrency (errgroup + semaphore) sized per service, pre-budgeting call volume (types x regions x ids), with SDK adaptive retry kept only as a safety net; more moving parts, but call volume becomes predictable and bounded.

## Decision
Own the rate limiting and concurrency in-process via a per-(service, region) token-bucket limiter and per-service bounded concurrency, keeping SDK adaptive retry only as a backstop. The assumption is that predictable, bounded call volume beats reactive retry-only behavior.

## Consequences
- Call volume is pre-budgeted and bounded per service and region, avoiding the throttling cliff from List and `--deep` GetResource fan-out.
- We accept the cost of maintaining and tuning in-process limiter and concurrency configuration rather than deferring entirely to the SDK.
- **Revisit if:** real-world throttle data shows the per-service limits are mis-sized; tune the limits per service as that data arrives.
