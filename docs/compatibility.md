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

The current `0.2.0` archives were rebuilt twice byte-identically from source commit `3953d40`: Engineering `91463b4d6e8a1a4bfa9478cf4718bdbfad87199fde22191af9d2b7c4b16d4149`, Distributed Systems `ce189c5a249f43f546bb0890a7774fb0a9737e7a950b5580e233126597840b79`, and Fintech `3951610b9bc1d2ceb0fa63ab28e8c17000deea1f42e921ea03e852b95ace0c84`. This refresh is reproducible packaging evidence, not a new clean-clone install or cross-client behavioral result.

Run `skillctl eval preflight` for the active runner environment. See [install.md](install.md) for layouts.
