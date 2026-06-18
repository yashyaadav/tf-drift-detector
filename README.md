# tf-drift-detector

Detect and triage **Terraform configuration drift** — compare your Terraform state against
live infrastructure and surface drift as clear, actionable findings.

> **Status:** 🚧 Work in progress — early scaffold.

## What is this?

Infrastructure managed by Terraform inevitably drifts from its declared state: someone
toggles a setting in the console, an out-of-band script mutates a resource, or a `terraform
apply` is skipped. `tf-drift-detector` aims to catch that drift early by inspecting a
Terraform plan / state diff and reporting *what* changed, *where*, and *how much it matters*.

## Goals

- Run `terraform plan` (or read an existing plan/state) and extract the set of drifted resources.
- Classify each drift by resource type and severity.
- Emit results in human-readable and machine-readable (JSON) formats for CI gates.

## Roadmap (high level)

- [ ] Parse Terraform plan JSON (`terraform show -json`)
- [ ] Identify and group drifted resources
- [ ] Severity classification
- [ ] CLI + JSON output
- [ ] CI integration example

## License

[MIT](./LICENSE)
