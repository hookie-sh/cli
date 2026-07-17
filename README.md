# Hookie CLI

Open-source CLI for [Hookie](https://hookie.sh): stream webhook events to your machine, inspect payloads locally, and forward to your dev server.

This repository contains the **client only** (CLI, embedded GUI, relay gRPC protocol). The hosted platform is not open source.

## Install

```bash
npm install -g @hookie-sh/cli
pnpm add -g @hookie-sh/cli
npx @hookie-sh/cli listen
```

Requires Node.js 18+ for the npm wrapper (which downloads the native binary). Supported platforms: macOS (amd64, arm64), Linux (amd64, arm64), Windows (amd64).

## Quick start

```bash
hookie login
hookie apps
hookie listen --app-id <app-id>
```

Run `hookie listen` with no app or source for a quick anonymous ephemeral webhook URL.

## Build from source

Prerequisites: Go 1.25+, Node.js 24+, pnpm, `protoc` with Go plugins.

```bash
pnpm install
make build-cli-dev
./bin/hookie --help
```

Copy `.env.example` to `.env` and set `CLERK_PUBLISHABLE_KEY` (and optionally `HOOKIE_WEB_APP_URL`, `HOOKIE_RELAY_URL`) for local development.

## Environment variables

| Variable | Purpose |
| -------- | ------- |
| `CLERK_PUBLISHABLE_KEY` | Required for `hookie login` when not embedded in the build. |
| `HOOKIE_WEB_APP_URL` | Base URL of the Hookie web app for login (default `https://hookie.sh` in prod builds). |
| `HOOKIE_RELAY_URL` | gRPC relay address (`host:port`). |
| `HOOKIE_CONFIG_DIR` | Directory for user config (token, machine ID). |
| `HOOKIE_INSECURE_TLS` | Non-empty value forces plaintext gRPC (local dev). |
| `HOOKIE_UI_PORT` | Port for the local event GUI (default `4840`). |

## Commands

```bash
hookie login
hookie logout
hookie apps
hookie sources
hookie listen --app-id <app-id>
hookie listen --source-id <source-id>
hookie listen --forward <url>
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
└── Makefile
```

## License

Apache-2.0. See [LICENSE](LICENSE).

## Security

See [SECURITY.md](SECURITY.md).

## Code of conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
