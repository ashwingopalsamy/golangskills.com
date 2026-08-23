# Repository instructions

Treat `skills/` and each `skill.json` as canonical source. Do not edit `catalog/catalog.json`, generated `agents/openai.yaml` files, files below `plugins/engineering-skills-for-go/skills/`, or files below `dist/npm/` by hand; update canonical input and run `go run ./cmd/skillctl generate` followed by `go run ./cmd/skillctl npm package -version <version>`.

Keep skills focused on one engineering job. Prefer causal invariants and decision procedures over style checklists. Classify source authority honestly: Go specifications and package documentation can be normative; company guides are organization-specific; operational literature is evidence, not a language rule.

Do not add executable scripts to a skill without a documented threat model and a demonstrated need that instructions cannot meet. Do not add a dependency to repository tooling when the Go standard library is sufficient.

Every skill change must preserve the repository discovery and body budgets and must update its routing and quality eval cases. Run the requested checks only when the user authorizes validation.
