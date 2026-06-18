# ADR-0004: Scanner registry: bespoke SDK scanners primary, Cloud Control experimental

- **Status:** Accepted
- **Date:** 2026-06-18

## Context
Fetching actual cloud state must avoid driftctl's fate: it hand-wrote a lister and deserializer per type, hit a ceiling around 120 types, and was archived under that maintenance burden. AWS Cloud Control API (CCAPI) offers a generic List/Get over many CloudFormation types, but its breadth is oversold: many types are not listable without a parent identifier, ListResources returns only partial properties (forcing N+1 GetResource), and its PascalCase CloudFormation schema does not match Terraform's snake_case attributes. We assume a small set of trustworthy bespoke scanners beats broad-but-noisy generic coverage for credibility.

## Options considered
- **CCAPI as the default generic scanner** — broad coverage, but low fidelity with an identity and field-mapping tax, and throttle-prone.
- **Bespoke aws-sdk-go-v2 scanners for a curated high-value set** — high fidelity per type, with CCAPI behind an explicit experimental `--generic` flag for breadth insurance.
- **Pure per-type hand-coding for everything** — full control, but the driftctl maintenance treadmill.

## Decision
We adopt a Scanner registry keyed by Terraform type. Bespoke SDK scanners (S3, security group, EC2 instance, IAM role to start) carry v1 and the demo, while CCAPI is experimental, opt-in, and set-membership-only.

## Consequences
- High-fidelity, trustworthy coverage for the curated set, avoiding CCAPI's partial-property and schema-mismatch problems.
- Breadth is limited to hand-built scanners; generic coverage is deferred to an experimental opt-in path.
- **Revisit if:** a measured, verified set of parent-free, identity-mappable CCAPI types proves large and reliable.
