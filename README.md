# Hookie CLI

Open-source CLI for [Hookie](https://hookie.sh): stream webhook events to your machine, inspect payloads locally, and forward to your dev server.

This repository contains the **client only** (CLI, embedded GUI, relay gRPC protocol). The hosted platform is not open source.

## Install

```bash
npm install -g @hookie-sh/cli
pnpm add -g @hookie-sh/cli
```

Requires Node.js 18+ for the npm wrapper (which downloads the native binary). Supported platforms: macOS (amd64, arm64), Linux (amd64, arm64), Windows (amd64).

## Quick start

```bash
hookie login
hookie apps
hookie listen --app-id billing-api-k7m2xp
```

Run `hookie listen` with no app or source for a quick anonymous ephemeral webhook URL (`/a/brave-falcon-k7m2xp`).

## Environment variables

| Variable              | Purpose                                                                                |
| --------------------- | -------------------------------------------------------------------------------------- |
| `HOOKIE_APP_URL`      | Base URL of the Hookie app for login (default `https://app.hookie.sh` in prod builds). |
| `HOOKIE_RELAY_URL`    | gRPC relay address (`host:port`).                                                      |
| `HOOKIE_CONFIG_DIR`   | Directory for user config (token, machine ID).                                         |
| `HOOKIE_INSECURE_TLS` | Non-empty value forces plaintext gRPC (local dev).                                     |
| `HOOKIE_UI_PORT`      | Port for the local event GUI (default `4840`).                                         |

## Commands

```bash
hookie login
hookie logout
hookie apps
hookie sources
hookie listen --app-id <app-public-id>
hookie listen --app-id <app-public-id> --source-id <source-slug>
hookie listen --forward-to <url>
hookie init
```

Shell completion: `hookie completion <bash|zsh|fish|powershell>`.

See `hookie --help` and subcommand `--help` for flags.

## Repository layout

```
├── cmd/             # CLI commands (Go)
├── internal/        # auth, relay client, embedded GUI server
├── proto/           # relay gRPC protocol (canonical)
├── packages/
│   ├── gui/         # Vite UI embedded in the CLI
│   └── cli/         # @hookie-sh/cli npm wrapper
├── .agents/skills/  # agent skills (tool-agnostic)
└── Makefile
```

## Releasing

Maintainers cut releases from `main` by pushing a `v*` tag (triggers [Release CLI](https://github.com/hookie-sh/cli/actions/workflows/release.yml)).

In Cursor, invoke **@create-release** to compute the next semver from conventional commits and push the tag after confirmation. The skill lives at `.agents/skills/create-release/SKILL.md` (symlinked from `.cursor/skills`).

After CI stages the npm package, test with [hookie-sh/release-tester](https://github.com/hookie-sh/release-tester) before `npm stage approve`.

## License

Apache-2.0. See [LICENSE](LICENSE).

## Security

See [SECURITY.md](SECURITY.md).

## Code of conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
