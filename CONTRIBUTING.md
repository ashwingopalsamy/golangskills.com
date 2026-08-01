# Contributing

## Quality bar

A skill belongs here only when it changes an agent’s decisions on a consequential, recurring Go engineering task. Before authoring, state:

1. the failure mode and who bears its cost;
2. the invariant or decision the base model commonly misses;
3. why repository evidence alone is insufficient;
4. the nearest existing skill and the non-overlapping boundary;
5. how a paired with-skill/without-skill evaluation can detect improvement.

Do not add a skill to encode formatting, naming taste, library promotion, or one organization’s convention as universal Go guidance.

## Skill contract

Each `skills/<name>/` directory contains:

- `SKILL.md`: portable instructions with only open-standard frontmatter;
- `skill.json`: version, category, compatibility, relations, and source provenance;
- `evals.json`: positive routing, negative routing, and output assertions;
- `agents/openai.yaml`: generated Codex UI metadata;
- `references/`: optional depth loaded only when the workflow routes to it.

Budgets enforced by `skillctl`:

- name: at most 64 characters and equal to the directory name;
- description: at most 500 characters in this repository;
- conservative serialized discovery payload: at most 4,000 characters per collection and 7,800 for all collections, including a 160-character location allowance per skill;
- `SKILL.md`: at most 300 lines and 3,000 words;
- at least two positive routing cases, two negative routing cases, and two quality cases;
- every material technical claim traceable to a current primary source or explicitly scoped operational source.

## Workflow

1. Add or change canonical files under `skills/`.
2. Update `docs/compatibility.md` only when verified platform behavior changes.
3. Run `go run ./cmd/skillctl generate`.
4. Run `go run ./cmd/skillctl check`, `go test ./...`, and `go vet ./...`.
5. Inspect the diff for generated-only churn or new overlap.

Do not publish benchmark scores without the agent, model, version, date, task fixtures, verifier, and baseline condition.
