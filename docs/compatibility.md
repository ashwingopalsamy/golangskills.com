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

The current `0.2.0` archives were rebuilt twice byte-identically from source commit `b5e49f3`: Engineering `da456e19f74103301a5b9f289a91fbd0c3b0b5119e2e5a0b60888346a0ecdf93`, Distributed Systems `6bf9c09e8a4a71eda90116caaef36e5482712b60af678b0dccdb1c25a2dd055d`, and Fintech `e77762256700b2ebbce3ebb76d4cec459df00a96de0ff0f070dc1ff55841f94c`. This refresh is reproducible packaging evidence, not a new clean-clone install or cross-client behavioral result.

Run `skillctl eval preflight` for the active runner environment. See [install.md](install.md) for layouts.
