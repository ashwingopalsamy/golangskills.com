# Compatibility

Verified on 2026-08-18 against current official product documentation. “Native” means the product documents Agent Skills discovery and activation. “Adapter” means the content is compatible but installation maps it into a product-specific directory or plugin package. It does not mean every product interprets optional metadata identically.

| Agent | Level | Project path or package | Notes |
| --- | --- | --- | --- |
| Codex | Native + plugin | `.agents/skills/`; Codex plugin | First-class `agents/openai.yaml`; Codex discovery metadata has a global character cap |
| Claude Code | Native + plugin | `.claude/skills/`; Claude plugin | Portable six-field frontmatter subset; Claude-specific fields are not used in canonical files |
| GitHub Copilot | Native | `.github/skills/` or `.agents/skills/` | Skill scripts inherit agent permissions; this collection ships none |
| Gemini CLI | Native + extension-compatible | `.agents/skills/` or extension `skills/` | Activation requires user consent in current Gemini CLI behavior |
| Cursor | Native | `.cursor/skills/` or `.agents/skills/` | Current editor and CLI support Agent Skills |
| OpenCode | Native | `.opencode/skills/` or `.agents/skills/` | Skills are loaded on demand through its skill tool |
| Roo Code | Native | `.roo/skills/` or `.agents/skills/` | Mode-specific directories are a Roo extension, not used here |
| Windsurf Cascade | Adapter | `.windsurf/skills/` | Cascade follows Agent Skills progressive disclosure; Devin Desktop additionally scans `.agents/skills/` |
| Cline | Adapter | `.cline/skills/` or installer-mapped path | Canonical content is portable; do not assume `.agents/skills/` discovery without the installer |

## Portable subset

Canonical `SKILL.md` frontmatter uses only `name`, `description`, `license`, and `compatibility`, all defined by the open specification. Experimental `allowed-tools` is omitted because tool syntax and enforcement differ by client. Product-specific invocation controls stay in `agents/openai.yaml` or the generated plugin manifest.

The entire `SKILL.md` body is assumed to enter context after activation. References are conditional and one level deep. No skill depends on client-only prompt substitutions, subagent syntax, or command names.

## Installation strategy

- Codex: generated repository marketplace and plugin package.
- Claude Code: generated plugin manifest using the same packaged skills.
- Other clients: `npx skills add` or manual copy of selected canonical skill directories.

The repository does not generate nine copied client trees. The ecosystem installer already owns path mapping, and checked-in copies would multiply drift without improving skill semantics.

## Native validation evidence

On 2026-08-18, Claude Code 2.1.212 loaded the generated plugin directly and reported all six skills. Its projected discovery cost was approximately 545 always-on tokens, with approximately 1.6k-1.9k tokens loaded per invoked core skill before optional references. Both the plugin and marketplace passed `claude plugin validate --strict`.

The Codex package passed the current local plugin-ingestion validator, including its manifest, skill paths, and generated `agents/openai.yaml` files. Codex CLI 0.148.0-alpha.9 also registered the checkout as the `golangskills` marketplace, installed and enabled `engineering-skills-for-go` 0.1.0 from the generated local source, and removed the temporary installation cleanly. Canonical skills passed the official Agent Skills reference validator at upstream commit `69ef37e9424c0a7ea9dd2293b559e43ec8176379`. These checks establish packaging and discovery compatibility, not equivalent model behavior across clients.
