# Security

## Reporting a vulnerability

If you believe you've found a security vulnerability in the Hookie CLI, please report it responsibly.

**Do not** open a public GitHub issue for security-sensitive bugs.

**Instead:**

1. Email the maintainers with a clear description of the issue, steps to reproduce, and impact. You can find maintainer contact in the repository.
2. Allow a reasonable time for a fix before any public disclosure (we aim to respond within 7 days and address critical issues as soon as possible).

We appreciate your efforts to disclose your findings in a responsible manner and will acknowledge your contribution when the issue is fixed (unless you prefer to remain anonymous).

## Scope

- The Hookie CLI: Go binary, embedded local GUI, gRPC client (`proto/`), and npm wrapper (`@hookie-sh/cli`).
- Out of scope: the hosted Hookie control plane (web app, ingest, relay server), third-party services (e.g. Clerk), and issues that require physical access or social engineering.

## Supported versions

We provide security updates for the current major release. When in doubt, use the latest release from [GitHub Releases](https://github.com/hookie-sh/cli/releases) or npm.
