# Security

## Reporting a vulnerability

Please report security issues privately via GitHub's **"Report a vulnerability"** (Security
Advisories) on this repository, rather than opening a public issue. You'll get an acknowledgement
within a few days.

## Security model

`tf-drift-detector` is designed to be **read-only**:

- It never runs `terraform plan`, `apply`, or any mutating Terraform command.
- It never writes to or modifies cloud resources — it only calls `Describe*` / `List*` / `Get*`
  APIs to enumerate live state.
- It reads Terraform state and cloud inventory; it does not perform remediation.

### Least-privilege AWS access

Grant the scanning principal **read-only** access to the services you scan, plus read access to the
state backend. AWS-managed `ReadOnlyAccess` is a convenient (broad) starting point; prefer a scoped
policy limited to the `Describe*`/`List*`/`Get*` actions of the resource types you actually scan,
plus `s3:GetObject` (and `kms:Decrypt` if applicable) for a remote state backend.

`AccessDenied` is surfaced as a distinct per-scope **scan status** — denied scopes are reported, not
silently treated as "no resources," so a permissions gap can never masquerade as a clean account.

### Handling sensitive data

- Terraform state can contain secrets in **plaintext**. This tool redacts values marked sensitive in
  state before they reach reports or logs, and never persists raw state.
- Do not commit real `*.tfstate` / `*.tfvars` files; they are git-ignored by default.
- With `--output json`, only stdout carries the report; diagnostic logs go to stderr.
