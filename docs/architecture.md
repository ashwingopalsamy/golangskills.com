# Architecture

## Design objective

The project optimizes a constrained objective rather than skill count:

```text
net value = task-quality delta
          - discovery cost
          - activation mistakes
          - instruction conflicts
          - maintenance drift
          - supply-chain risk
```

A skill is successful only if it improves a consequential task relative to the same agent without the skill. Structural validity is necessary but not sufficient.

## Canonical model

`skills/` is the only authored skill corpus. Open Agent Skills Markdown is already the portable intermediate representation; inventing another content DSL would create a lowest-common-denominator abstraction and a second source of truth.

Each skill separates four concerns:

| Artifact | Authority |
| --- | --- |
| `SKILL.md` | Activation description and procedural instructions |
| `skill.json` | Catalog identity, version, status, compatibility, relations, and provenance |
| `evals.json` | Routing and output behavior to measure |
| `references/` | Conditional depth for a specific decision branch |

The Go tool reads those inputs and generates:

- `agents/openai.yaml` inside each canonical skill;
- `catalog/catalog.json` for the future website and external tooling;
- a Codex plugin package under `plugins/engineering-skills-for-go/`;
- a Claude plugin manifest in that package, reusing the same skill tree.

Generated artifacts are checked for freshness. No client adapter is allowed to edit the skill body.

## Taxonomy

The v1 taxonomy follows failure boundaries rather than language syntax:

- **execution:** concurrency lifecycle and HTTP transport boundaries;
- **state:** SQL transaction boundaries and message-processing semantics;
- **reliability:** remote-call resilience;
- **workflow:** risk-driven production change review.

These skills compose only when the change crosses boundaries. They do not require an orchestrator and do not instruct an agent to load the whole collection.

Generic formatting, naming, declarations, and basic error wrapping are deliberately absent from v1. Current base models and repository-local instructions usually cover them, while conflicting generic guidance can reduce task performance.

## Authority model

Every source has one of four kinds:

- `normative`: a language specification, standard-library contract, or open skill specification;
- `primary`: official project documentation describing implemented behavior;
- `operational`: first-party production guidance whose causal model generalizes but is not a Go rule;
- `organization-specific`: a convention valid for its publisher unless repository evidence adopts it.

The validator rejects missing verification dates and unknown kinds. Skill prose must not elevate a lower-authority source into a universal requirement.

## Versioning

Skills use independent semantic versions because a routing-description change can affect one skill without changing the others. The collection and plugin use their own semantic version. `status` is one of `experimental`, `beta`, or `stable`; v1 skills begin at `beta` until paired evaluations have been repeated across at least two current agent/model combinations.

Go guidance targets the two stable Go release families recorded in each `skill.json`. A skill must isolate version-sensitive advice and retain a compatible path when the repository’s declared Go version is older.

## Security boundary

Instruction-only skills are the default. Scripts require a separate threat model because empirical ecosystem research finds executable skills carry materially higher vulnerability rates. Client-specific tool grants are not placed in portable frontmatter; the user’s agent and repository policy remain authoritative.
