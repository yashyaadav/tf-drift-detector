# ADR-0000: Record architecture decisions

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
tf-drift-detector is an agentless Terraform drift detector written in Go, and its significant architectural decisions need to be documented alongside their context and consequences. Without a durable record, the reasoning behind decisions made in code review or scattered planning docs gets lost, leaving future contributors and reviewers unable to audit why a choice was made. We assume that a lightweight, low-friction format kept next to the code is more likely to be maintained than a separate heavyweight process.

## Options considered
- **Keep decisions only in code review / a planning doc** — zero new tooling, but rationale is ephemeral and gets lost as reviews scroll off and planning docs go stale.
- **Adopt lightweight Markdown ADRs in docs/adr** — numbered, append-only files where superseded decisions are never deleted; small per-decision writing overhead in exchange for a durable, auditable history.

## Decision
We adopt MADR-lite ADRs stored as Markdown files in `docs/adr/NNNN-title.md`, numbered sequentially and append-only.

## Consequences
- Durable rationale lives next to the code and can be audited by future contributors and reviewers.
- A small per-decision authoring overhead is accepted for each significant architectural choice.
- Reversals are recorded as new superseding ADRs rather than edits or deletions, preserving history.
- **Revisit if:** the overhead discourages contributors from recording decisions, or the ADR set grows large and unwieldy enough that the flat numbered-file format no longer scales.
