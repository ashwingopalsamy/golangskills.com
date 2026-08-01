# Installation layouts

All distributed files are generated from canonical `skills/`. Choose one, two, or all collections; skill IDs never overlap.

## Codex

Each collection under `plugins/` has `.codex-plugin/plugin.json`, `skills/`, and generated `agents/openai.yaml`. The repository marketplace at `.agents/plugins/marketplace.json` lists all three local plugins.

## Claude Code

Each collection includes `.claude-plugin/plugin.json`; `.claude-plugin/marketplace.json` lists the three sources. The current release validates this structure but does not claim behavioral performance because behavioral validation depends on an authenticated runner.

## Cursor

Each collection is a portable Agent Plugin with root `plugin.json`. Project-local copies are also available under `adapters/cursor/<collection>/.cursor/skills/`. The current release validates structure but cannot run Cursor behavior because behavioral validation depends on an authenticated runner.

## OpenCode

Copy the selected `adapters/opencode/<collection>/.agents/skills/` tree into a project or configure it as a skills source. The current release validates the documented `.agents/skills/<name>/SKILL.md` contract but cannot run behavior because the OpenCode adapter has structural compatibility evidence; behavioral validation requires an authenticated runner.

## Open Agent Skills discovery

The repository-root `skills/` directory is canonical and compatible with tools that install selected skills from a Git repository, including the open `npx skills` discovery workflow. Collection membership is published in `catalog/collections.json` for installers that want collection-level selection.

Generated adapters are disposable. Run `go run ./cmd/skillctl generate` after changing a canonical skill, then `go run ./cmd/skillctl check`.
