# ADR-0001: Agentless detection model (parse state + read cloud directly)

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
Terraform state drifts from reality through ClickOps, out-of-band scripts, and skipped applies, so we must compare desired state against actual cloud configuration. We need to do this without the risk and latency of running Terraform itself. The core assumption is that tfstate is a usable desired-state proxy and that read-only cloud APIs are sufficient to observe actual state without launching the provider plugin.

## Options considered
- **Agentless (parse state + read cloud)** — Treat tfstate as a passive artifact, enumerate live cloud via read-only SDK/APIs, and diff in-process with no terraform binary and no plan/apply (driftctl/Firefly model); trade-off is that we own coverage for each resource type rather than inheriting it from a provider.
- **Stateful-replay** — Run a real terraform refresh-only / plan -detailed-exitcode and read the provider's output (HCP/Spacelift/env0/GitHub-Action model); inherits full provider coverage but brings apply risk, slow plans, terraform binary and provider version skew, and exit-code-2 false positives from pending config and Checks.

## Decision
We choose the agentless model as the core of the product and its key differentiator: parse tfstate as a passive artifact and diff it against live cloud read via read-only APIs.

## Consequences
- No terraform binary, no provider plugin, and no apply risk; detection is fast and safe to run continuously.
- We accept responsibility for cloud-side coverage per resource type rather than inheriting it from a provider.
- **Revisit if:** tfstate proves an inadequate desired-state proxy or read-only APIs cannot cover long-tail types, in which case an opt-in tfexec refresh-plan hybrid escape hatch may be added.
