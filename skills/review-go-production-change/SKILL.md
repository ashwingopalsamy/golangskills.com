---
name: review-go-production-change
description: "Review a Go diff or change plan for production correctness using risk-driven causal analysis across concurrency, persistence, remote calls, compatibility, security, and operability. Use for PRs that change stateful or distributed behavior. Do not use for formatting-only review, greenfield implementation, or exhaustive whole-repo audits."
license: Apache-2.0
compatibility: "Go 1.24 or newer. Repository contracts and changed behavior outrank generic conventions."
---

# Review a Go production change

Review the behavior the change creates, not its resemblance to a generic checklist. Find issues that can change correctness, safety, durability, availability, compatibility, or incident response. Repository evidence outranks external style preferences.

## Establish review scope

Inspect the smallest path needed to understand the diff:

1. changed code and configuration;
2. callers and callees that define the changed contract;
3. schema, protocol, or migration artifacts touched by the behavior;
4. tests only as evidence of intended behavior, not proof of correctness;
5. history only when it resolves an otherwise material ambiguity.

State the externally observable behavior and the invariants the change must preserve. Do not review unrelated files merely for completeness.

## Build a change graph

Map the changed input to side effects and outputs:

```text
entry -> validation/admission -> state read -> decision
      -> local write / remote effect -> commit or ack -> response/telemetry
```

Mark new or changed boundaries:

- goroutine or queue;
- transaction or migration;
- network attempt or timeout;
- broker publish/acknowledgment;
- serialization or public API;
- authentication/authorization/trust;
- shutdown/startup or operational signal.

Review only the lanes the graph touches.

## Analyze failure schedules

For each changed side effect, consider the smallest relevant counterexamples:

- concurrent execution with the same or conflicting identity;
- cancellation before work, during a blocking call, and after a durable effect;
- timeout with unknown remote outcome;
- crash immediately before and after commit, publish, or acknowledgment;
- retry or redelivery after partial success;
- overload with queues and dependencies saturated;
- mixed old/new versions during rollout or rollback;
- malformed, unauthorized, or adversarial input at a trust boundary.

Do not mechanically enumerate schedules that cannot affect the change.

## Select the applicable risk lane

### Concurrency

Check owner, lifetime, termination, capacity, and ordering for every changed goroutine or shared value. Use the reasoning in `go-concurrency-lifecycle` when the diff adds real concurrent paths.

### Persistence

Check the durable invariant, transaction owner, isolation, retry scope, and commit ambiguity. Ensure external effects are not incorrectly assumed atomic with database state. Use `go-sql-transactions` for SQL paths.

### Messaging

Check stable identity, deduplication durability, ordering scope, effect-before-ack sequence, poison handling, and outbox/inbox gaps. Use `go-message-processing` when redelivery or publication semantics matter.

### Remote calls

Check propagated deadlines, replay safety, retry ownership, maximum amplification, admission, and ambiguous results. Use `go-service-resilience` and, for `net/http` resource contracts, `go-http-boundaries`.

### Compatibility

Check public Go APIs, serialized fields, error identities, database schemas, message versions, configuration defaults, and rolling-version coexistence. A source-compatible Go change can still break runtime data or operations.

### Security and trust

Trace untrusted data through parsing, authorization, query/command construction, filesystem/network destinations, logs, and returned errors. Check authorization at the resource/action boundary, not merely authentication middleware. Do not assert a vulnerability without an executable trust path.

### Operability

Check whether on-call can distinguish success, rejection, overload, dependency failure, retry, and ambiguous outcome. Labels must be bounded; logs must avoid secrets and sensitive payloads. Ensure shutdown and recovery preserve the same invariants as steady state.

Read [references/finding-quality.md](references/finding-quality.md) before writing findings.

## Respect evidence and authority

Classify a concern before reporting it:

- **defect:** a concrete schedule violates a required contract;
- **risk:** the diff makes a consequential assumption that repository evidence does not establish;
- **preference:** an alternative may be clearer but current behavior remains correct.

Report defects and material risks. Omit preferences unless the user explicitly asks for style feedback. Do not turn Google, Uber, or another organization’s guide into a language rule unless this repository adopts it.

## Avoid common review failures

- Do not demand tests, comments, interfaces, layers, or abstractions by default.
- Do not report an issue already prevented by a constraint visible in the same execution path.
- Do not flag theoretical races without overlapping access and a missing ordering edge.
- Do not recommend retries without replay semantics and a budget.
- Do not treat a transaction as protection for effects outside its database.
- Do not infer authorization from route reachability.
- Do not claim a check passed unless it was actually run.
- Do not hide a high-impact failure among style nits.

## Write actionable findings

Each finding contains:

1. **Title:** imperative and specific to the consequence.
2. **Location:** the smallest changed line range that creates the issue.
3. **Trigger:** the concrete input, interleaving, crash point, or rollout state.
4. **Impact:** corrupted state, duplicate effect, outage, leak, security boundary break, or compatibility failure.
5. **Mechanism:** the causal chain from code to impact.
6. **Correction:** the smallest design change that restores the invariant, when clear.

Use severity from impact and likelihood, not complexity:

- **P0:** immediate catastrophic or actively exploitable impact;
- **P1:** likely severe production failure or data/security damage;
- **P2:** material failure under plausible conditions;
- **P3:** bounded correctness or maintenance risk worth fixing.

If there are no actionable findings, say so and name only significant residual uncertainty. Do not invent a finding to make the review look complete.

## Verification boundary

Review is analysis, not authorization to modify files or run broad validation. Use read-only inspection by default. If the user requested checks, run the smallest checks that exercise the suspected behavior and report their actual scope. A passing test is evidence about paths executed under its environment; it is not a regression-free guarantee.

## Output contract

Lead with findings in descending severity, each tied to a changed line. Keep the summary short. Separate unverified residual risks from findings. Do not provide a full checklist, praise, or generic best-practice lecture.
