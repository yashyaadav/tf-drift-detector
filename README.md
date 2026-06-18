# tf-drift-detector

**Agentless Terraform drift detection.** Compare a Terraform **state file** against your
**live cloud** and surface drift as clear, actionable findings — **without running
`terraform plan` or `apply`**, and without a Terraform binary in the hot path.

> **Status:** 🚧 Early development — building the M0 vertical slice (see [Roadmap](#roadmap)).

```text
$ tfdrift scan --types aws_security_group --region us-east-1

  82% of live resources are managed by Terraform  ·  3 unmanaged  ·  1 missing

  CLASS       TYPE                 ID                     REGION
  unmanaged   aws_security_group   sg-0abc123 (ClickOps)  us-east-1
  missing     aws_security_group   sg-0def456             us-east-1
```

## Why

Infrastructure managed by Terraform inevitably drifts from reality: someone toggles a setting
in the console, an out-of-band script mutates a resource, or an `apply` gets skipped. The most
dangerous drift is **invisible to `terraform plan`** — resources created entirely outside of
Terraform (ClickOps "orphans") never appear in a plan because Terraform doesn't know they exist.

`tf-drift-detector` reads your Terraform **state** as the source of *desired* truth, independently
**enumerates live cloud resources** via read-only provider APIs, and reconciles the two. It
classifies every resource as:

- **managed** — in state *and* in the cloud (optionally attribute-diffed with `--deep`)
- **unmanaged** — in the cloud but **not** in any state (the headline: ClickOps / orphans)
- **missing** — in state but **deleted** out-of-band

…and reports a single **IaC coverage %** — *what fraction of your live infrastructure Terraform
actually manages.*

## How it works

```
Terraform state ─▶ StateLoader ─▶ StateReader ─┐
                                                ├─▶ Normalizer ─▶ DriftEngine ─▶ Reporter ─▶ console / JSON
Cloud APIs ─▶ Provider.Scan ─▶ Scanner registry ┘                (set-membership +        (+ history later)
                                                                  coverage %)
```

- **Agentless** — no `terraform plan`/`apply`, no apply risk, fast read-only scans.
- **Extensible** — a `Provider` interface (AWS first; Azure/GCP later) and a per-resource
  `Scanner` registry. The real unit of work per type is the **identity join** (Terraform `id` ↔
  cloud primary identifier), enforced by conformance tests.
- **Trustworthy by default** — a built-in suppression heuristic hides AWS-managed/default
  resources (service-linked roles, default VPC/SG, `aws:`-prefixed tags, AWS-managed KMS) so a
  clean account reports ~zero false orphans. Override with `.driftignore`.

The architectural reasoning behind these choices is recorded as [ADRs](docs/adr/).

## Install

```bash
go install github.com/yashyaadav/tf-drift-detector/cmd/tfdrift@latest
```

> The repository is `tf-drift-detector`; the binary it produces is **`tfdrift`**.

## Usage

```bash
# Scan everything in the default region against a local state file
tfdrift scan --state ./terraform.tfstate

# Scope a scan (fast, deterministic, throttle-safe)
tfdrift scan --state ./terraform.tfstate --types aws_security_group,aws_s3_bucket --region us-east-1

# Machine-readable output for CI / jq (logs go to stderr; stdout stays pure JSON)
tfdrift scan --state ./terraform.tfstate --output json | jq '.summary'

# CI gate: fail the build if any unmanaged resource is found
tfdrift scan --state ./terraform.tfstate --fail-on unmanaged
```

**Exit codes:** `0` = clean · `2` = drift detected · `1` = operational error.

Required AWS permissions are **read-only** (`Describe*`/`List*`/`Get*` for the scanned services,
plus read access to the state backend). See [SECURITY.md](SECURITY.md).

## Roadmap

- **M0** — local state, single account/region, bespoke scanners (S3, security group, EC2
  instance, IAM role), set-membership classification, console + JSON, default suppression,
  coverage %, scope flags. *(in progress)*
- **M1** — more resource types, S3 remote state, `--deep` attribute diff, `.driftignore`,
  `--fail-on` CI gate, multi-region fan-out with rate limiting + per-scope scan status.
- **M2** — experimental `--generic` (AWS Cloud Control) breadth scanner, SQLite scan history.
- **M3** — scheduler (one-shot + cron) and an embedded web dashboard.
- **M4** — Azure & GCP providers, multi-account scanning, opt-in remediation *suggestions*.

## Development

```bash
make build   # compile ./bin/tfdrift
make test    # run tests (add UPDATE=1 to refresh golden files)
make check   # fmt + vet + lint + test
```

## License

[MIT](./LICENSE) © 2026 Yash Yadav
