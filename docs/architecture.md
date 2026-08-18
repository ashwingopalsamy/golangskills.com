# Architecture

## Canonical model

`skills/` is the authored corpus. Each skill contains portable `SKILL.md`, schema-v2 `skill.json`, schema-v2 `evals.json`, optional one-hop Markdown references, and generated `agents/openai.yaml`.

`skill.json` owns collection, maturity, risk domains, Go support, claim IDs, relations, compatibility evidence, primary sources, and source provenance. `evals.json` separates routing decisions from quality cases with expected invariants, forbidden outcomes, deterministic graders, optional fixtures, and semantic rubrics.

The knowledge plane is independent:

- `research/corpus-lock.json`: repository identity, licenses, every file hash, canonical skill inventory, and hashed material markers;
- `knowledge/claims/canonical.json`: adjudicated technical claims and primary evidence;
- `knowledge/claims/reference-dispositions.json`: one disposition for every material marker;
- `THIRD_PARTY_NOTICES.md`: reuse policy and attribution.

## Generation

`skillctl generate` produces three Codex/Claude/Agent Plugin packages, Cursor `.cursor/skills` adapters, OpenCode `.agents/skills` adapters, catalogs, search data, site data, and `llms.txt`. Generated content never forks canonical knowledge.

## Context contract

Every `SKILL.md` is below 250 lines and 1,800 words. Conditional depth is one reference hop. Conservative discovery accounting assumes a 160-character installation path, limits each collection to 4,000 characters, and all collections together to 7,800.

## Authority

Claims distinguish adopted, qualified, organizational, version-specific, rejected, and outside-scope guidance. The Go specification and official implemented contracts outrank community convention. Security, financial, normative, and version-sensitive claims require primary evidence and explicit counterexamples.

## Security

Published skills are instruction-only. The validator rejects executable skill resources, symlinks, unexpected files, escaping references, stale evidence, and generated drift. Packaging uses deterministic archives, checksums, and provenance.
