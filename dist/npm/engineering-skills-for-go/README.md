# Engineering Skills for Go

`@golangskills/engineering@0.3.0-rc.1` is a versioned Agent Skills collection for production Go work. It is authored by [Ashwin Gopalsamy](https://ashwingopalsamy.in) and distributed under Apache-2.0.

## Install

```sh
npm install @golangskills/engineering@0.3.0-rc.1
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

- `go-language-engineering`
- `go-performance-and-diagnostics`
- `go-production-operations`
- `go-project-and-api-design`
- `go-security-hardening`
- `go-service-boundaries`
- `go-testing-and-verification`
- `review-go-engineering-change`

Source repository: https://github.com/ashwingopalsamy/golangskills.com
