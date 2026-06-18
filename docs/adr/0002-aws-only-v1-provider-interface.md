# ADR-0002: AWS-only v1 behind a Provider interface

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
Resource-type breadth is inherently per-cloud, so spreading effort across AWS, Azure, and GCP early dilutes the quality of each. driftctl's Azure and GCP support lagged badly behind its AWS support, illustrating the cost of going wide too soon. We are working under the assumption that nailing one provider end-to-end is worth more than shallow multi-cloud coverage.

## Options considered
- **AWS-only v1 with a clean Provider interface** — focus on one cloud end-to-end while keeping a seam for other clouds to slot in later; trade-off is no day-one multi-cloud story.
- **Multi-cloud from the start** — broad appeal up front, but a much larger surface area and slower path to a credible demo.

## Decision
Ship AWS-only behind a Provider abstraction, with Azure (Resource Graph) and GCP (Cloud Asset Inventory) planned as future implementations of the same interface.

## Consequences
- Concentrated effort yields deep, credible AWS coverage and a faster path to a working demo.
- No multi-cloud support at launch, and we accept the design cost of maintaining the Provider abstraction before a second cloud exists.
- **Revisit if:** AWS coverage is solid and a concrete second-cloud need emerges.
