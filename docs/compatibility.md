# Compatibility

Verified structurally on 2026-08-18 against current official contracts.

| Client | Generated surface | Behavioral evidence |
|---|---|---|
| Codex | Three plugins, repository marketplace, `agents/openai.yaml` | Primary runner; representative and release harness supported |
| Claude Code | Three plugin manifests and marketplace | Structurally compatible; behavioral evidence pending |
| Cursor | Portable Agent Plugins plus `.cursor/skills` adapters | Structurally compatible; behavioral evidence pending |
| OpenCode | `.agents/skills` adapters | Structurally compatible; behavioral evidence pending |
| Open Agent Skills installers | Canonical repository-root `skills/` | Portable discovery surface; installer behavior is tool-specific |

Canonical frontmatter uses `name`, `description`, `license`, and `compatibility`. Client-specific metadata is generated. No behavioral superiority claim transfers from Codex to another client.

The current tree passes the pinned official Agent Skills validator for all 20 canonical skills, the Codex plugin validator for all three packages, and Claude Code manifest validation for all three packages plus its marketplace. Cursor and OpenCode have structural schema/layout evidence; behavioral validation depends on authenticated runners.

Run `skillctl eval preflight` for the active runner environment. See [install.md](install.md) for layouts.
