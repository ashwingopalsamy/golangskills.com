# Go: Fintech

`@golangskills/fintech@0.4.0` is a versioned Agent Skills collection for production Go work. It is authored by [Ashwin Gopalsamy](https://ashwingopalsamy.in) and distributed under Apache-2.0.

## Install

```sh
npm install @golangskills/fintech@0.4.0
```

The package is data-only. npm installation does not run a lifecycle script or change an agent configuration.

## Client layouts

- Codex: use the package root as a plugin source; it contains `.codex-plugin/plugin.json` and `skills/`.
- Claude Code: use the package root as a plugin source; it contains `.claude-plugin/plugin.json` and `skills/`.
- Cursor: copy `cursor/.cursor/skills/` into the project’s `.cursor/skills/` directory.
- OpenCode: copy `opencode/.agents/skills/` into the project’s `.agents/skills/` directory.

## GitHub discovery

The open Skills CLI installs directly from the canonical repository:

```sh
npx skills add ashwingopalsamy/golangskills.com
```

## Contents

- `go-clearing-settlement-reconciliation`
- `go-financial-idempotency`
- `go-fintech-security-compliance`
- `go-money-and-ledgers`
- `go-payment-lifecycles`
- `review-go-fintech-change`

Source repository: https://github.com/ashwingopalsamy/golangskills.com
