# Contributing

Thanks for your interest in `tf-drift-detector`! It's an early-stage, agentless Terraform drift
detector. Contributions of all sizes are welcome.

## Getting started

```bash
git clone git@github.com:yashyaadav/tf-drift-detector.git
cd tf-drift-detector
make build      # compile ./bin/tfdrift
make check      # fmt + vet + lint + test
```

Requires **Go 1.24+**.

## Project shape

- `cmd/tfdrift` — binary entrypoint (wiring only).
- `internal/` — application code (CLI, state reading, cloud scanners, engine, reporters).
- `pkg/` — the stable, importable core (`resource`, `schema`, `provider` interfaces).
- `docs/adr/` — Architecture Decision Records. **Read these before proposing structural changes** —
  most "why is it done this way?" questions are answered there.

## Adding a new AWS resource type

A resource type is not "done" when it lists. The unit of work is the **identity join**: mapping a
Terraform `id` to the cloud's primary identifier so set-membership classification is correct. A new
type must ship:

1. A `Scanner` implementation (or a registry entry for the generic path).
2. An `IdentityFn` mapping Terraform `id` ↔ cloud identifier.
3. A **conformance test** asserting the identity round-trips between a `testdata` tfstate fixture and
   a cloud fixture.
4. Any necessary default-suppression entries for AWS-managed/default instances of that type.

See [ADR-0005](docs/adr/0005-per-type-identity-join.md).

## Conventions

- Run `make check` before opening a PR; CI runs `go vet`, `go test -race`, and `golangci-lint`.
- Keep `stdout` pure for `--output json`; all logs go to `stderr`.
- The JSON output in `pkg/schema` is a **versioned contract** — additive changes only within a major.
- Record any new architectural decision as an ADR.

## Reporting bugs

Open an issue with the command you ran, the (redacted) output, and your Go/OS versions. Never paste
real Terraform state or credentials.
