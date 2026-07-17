# @hookie-sh/cli

npm wrapper for the **Hookie CLI**. Postinstall downloads the correct platform binary from [GitHub Releases](https://github.com/hookie-sh/cli/releases).

## Install

```bash
npm install -g @hookie-sh/cli
pnpm add -g @hookie-sh/cli
npx @hookie-sh/cli listen
```

**Requirements:** Node.js 18+. macOS (amd64, arm64), Linux (amd64, arm64), Windows (amd64).

## Usage

```bash
hookie login
hookie apps
hookie listen --app-id <app-id>
```

See the [repository README](https://github.com/hookie-sh/cli/blob/main/README.md) for build-from-source and environment variables.

## License

Apache-2.0
