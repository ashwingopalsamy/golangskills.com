# Benchmark and leadership contract

## Locked comparison set

The local freeze includes `cc-skills-golang`, `cxuu-golang-skills`, `spf13/go-skills`, `gophers`, the specialist `golang-ddd-skills` collection, the Google style snapshot, and the Uber style guide. Installable benchmark arms also include the no-skill baseline and this project. Style-only repositories contribute claims but are not installable benchmark arms. The additional-arm selection rationale is recorded in `research/competitor-selection.md`.

## Protocol

Run both native full-collection discovery and explicitly matched skills. Use at least 60 committed scenarios across implementation, diagnosis, design, and review. Each cell gets a fresh directory, new ephemeral session, randomized opaque arm, randomized execution order, identical runner/model/settings, and no cross-arm state. Development cases are public; release holdouts remain private until scoring, then become public and rotate.

Score executable behavior first. The committed suite includes 12 initially failing Go modules; treatment arms edit an isolated copy without the decisive tests, then the harness injects a read-only post-edit oracle, runs an allowlisted offline Go-test grader, and retains the resulting source, oracle, and output. This prevents visible-test repair from masquerading as domain-skill improvement. Use blinded semantic evaluation only when a deterministic oracle cannot decide. Freeze the rubric and evaluator configuration before treatment. Globally randomize opaque candidates, use a fresh ephemeral session for each, require schema-constrained criterion-level verdicts, recompute scores outside the model, and bind every verdict to the exact response and rubric digest. A compound invariant requires explicit support for every clause; topic mentions and promises of later review are insufficient. Missing, malformed, mixed-model, stale, or identity-mismatched judgments fail closed. Lexical checks are diagnostic once a complete semantic verdict exists, so vocabulary selected by this project's prose cannot become a hidden gate. Label Codex-on-Codex judgment `same_platform`, never independent. Repeat hard or uncertain cells three to five times. Persist raw response and judgment JSONL checkpoints.

For paired reports, match arms by canonical skill and case ID. A strict win receives 1, a tie 0.5, and a loss 0; resample complete pairs for confidence intervals. Report collection-level score deltas and total client-reported input tokens alongside score per thousand input tokens. If either arm lacks a case or token accounting, the corresponding completeness or Pareto result remains ineligible rather than being silently imputed.

Freeze routing and skill maps before treatment runs. Canonical positive cases map to the arm's matching skill; confusion-negative cases map to the declared adjacent skill rather than `NONE`. Committed canonical-to-arm skill maps drive both routing and explicit cells, with per-case routing overrides where needed. Missing mappings fail closed, and arm manifests must match the audited competitor commits and real skill directories.

## Category-leadership gates

A release may claim leadership only when all gates pass:

- 100% reference inventory and disposition coverage;
- no unresolved normative, version-sensitive, security, or financial claim conflict;
- all structure, generation, licensing, install, and freshness checks pass;
- hidden routing macro-F1 at least 0.95, every collection at least 0.90, unrelated false activation at most 2%;
- no discovery omission caused by context pressure;
- zero unresolved critical correctness or financial-integrity failures;
- deterministic fixture pass rate at least five points above the strongest competitor;
- paired win rate at least 60% against every competitor with the 95% lower bound above 50%;
- no domain more than five points below the best competitor;
- positive skill-assisted improvement over no-skill baseline;
- quality per thousand input tokens Pareto-superior or best-in-class;
- every number links to raw artifacts, fixture commit, client/model, grader version, and scoring code.

Until then, publish evidence without “number one” claims. The generated benchmark status intentionally remains ineligible.

`skillctl release check` validates `evaluations/reports/release-gates.json` against the complete required gate inventory, rejects unknown or missing gates, verifies every repository-local evidence path, and checks that `leadership_claim_eligible` is exactly equivalent to every gate passing. A hand-edited boolean cannot bypass the release contract.
