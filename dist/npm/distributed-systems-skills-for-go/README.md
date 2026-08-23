# Distributed Systems Skills for Go

`@golangskills/distributed-systems@0.3.0-rc.1` is a versioned Agent Skills collection for production Go work. It is authored by [Ashwin Gopalsamy](https://ashwingopalsamy.in) and distributed under Apache-2.0.

## Install

```sh
npm install @golangskills/distributed-systems@0.3.0-rc.1
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

- `go-concurrency-lifecycle`
- `go-data-consistency`
- `go-distributed-coordination`
- `go-message-processing`
- `go-service-resilience`
- `review-go-distributed-change`

Source repository: https://github.com/ashwingopalsamy/golangskills.com
