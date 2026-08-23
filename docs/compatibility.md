# Compatibility

Verified structurally on 2026-08-24 against current official contracts.

| Client | Generated surface | Behavioral evidence |
|---|---|---|
| Codex | Three plugins, repository marketplace, `agents/openai.yaml` | Primary runner; representative and release harness supported |
| Claude Code | Three plugin manifests and marketplace | Structurally compatible; behavioral evidence pending |
| Cursor | Portable Agent Plugins plus `.cursor/skills` adapters | Structurally compatible; behavioral evidence pending |
| OpenCode | `.agents/skills` adapters | Structurally compatible; behavioral evidence pending |
| Open Agent Skills installers | Canonical repository-root `skills/` | Portable discovery surface; installer behavior is tool-specific |

Canonical frontmatter uses `name`, `description`, `license`, and `compatibility`. Client-specific metadata is generated. No behavioral superiority claim transfers from Codex to another client.

The current tree passes the pinned official Agent Skills validator for all 20 canonical skills, the Codex plugin validator for all three packages, and Claude Code manifest validation for all three packages plus its marketplace. Cursor and OpenCode have structural schema/layout evidence; behavioral validation depends on authenticated runners.

A clean local clone at commit `bb68087` reproduced all three archive checksums, registered the Codex marketplace, installed and enabled all three plugins, and routed a future payment-event prompt to `fintech-skills-for-go:go-payment-lifecycles`. The temporary installations and marketplace registration were removed after the check.

The current `0.2.0` archives were rebuilt twice byte-identically from source commit `b942392`: Engineering `e29e341af71c494b4dc97782fb2d25f4ac16875fe8eff35578f96eaa41fce17c`, Distributed Systems `96cddad77147298ef7f6167911f7cb044e9767ce738fa25c8b318a1dd115b7d4`, and Fintech `5e30a93d65a360c6edba202ee543770301c014cc15ddfc409104d002fce80478`. This refresh is reproducible packaging evidence, not a new clean-clone install or cross-client behavioral result.

Run `skillctl eval preflight` for the active runner environment. See [install.md](install.md) for layouts.
