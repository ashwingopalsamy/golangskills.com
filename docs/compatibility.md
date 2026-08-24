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

The current `0.2.0` archives were rebuilt twice byte-identically from source commit `fdfd794`: Engineering `794dab1b67e9d211b7884faa4a97148e25d7ff9d703edfdebc3ffd532ec4a3fc`, Distributed Systems `bb01cbf4274873972ddf0d534b94b346b82500c0a273c61835de03dbe36e857a`, and Fintech `c2609986e0d48f40d60a26c3d6a51624abdb5d37974a6972211957a9168bc65e`. This refresh is reproducible packaging evidence, not a new clean-clone install or cross-client behavioral result.

Run `skillctl eval preflight` for the active runner environment. See [install.md](install.md) for layouts.
