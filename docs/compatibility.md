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

The current `0.2.0` archives were rebuilt twice byte-identically from source commit `a9ca133`: Engineering `0dd6cca2a2adae5824b23463a3be9fec94e33f9c63742f496c55d52b963db814`, Distributed Systems `203bf0005c5b846c9565ef4313038c8384eddf450c34e01b2a0744a312d510c0`, and Fintech `cdfb49abc03bd09e6be923270a266d141c5492ab70c0eeb6687480c454f346e5`. This refresh is reproducible packaging evidence, not a new clean-clone install or cross-client behavioral result.

Run `skillctl eval preflight` for the active runner environment. See [install.md](install.md) for layouts.
