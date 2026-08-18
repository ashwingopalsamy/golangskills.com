# Agent evaluation

The repository distinguishes structural eval coverage from observed agent improvement. `evals.json` defines checked fixtures for every skill; paired model runs determine whether those fixtures produce marginal value.

## Protocol

1. Hold the agent, model, scenario, and scoring rubric constant.
2. Run a baseline arm instructed not to read repository skills.
3. Run a treatment arm instructed to read only the selected skill and relevant references.
4. Score externally observable engineering decisions as binary criteria. Do not score tone or wording.
5. Test discovery separately using only the available-skill names and descriptions.
6. Preserve conditions, prompts, raw outputs, scores, and limitations.

This is not a statistically useful benchmark until scenarios have repeated trials across at least two current agent/model combinations. A same-model judge is a consistency check, not an independent oracle.

## Initial exploratory forward test

Run: [2026-08-18 Codex GPT-5.6 Sol](../evaluations/runs/2026-08-18-codex-gpt-5.6-sol.md)

| Scenario | Baseline | With skill | Delta | Observation |
| --- | ---: | ---: | ---: | --- |
| Payment consumer with Kafka, PostgreSQL, and remote rewards | 9/10 | 10/10 | +1 | Skill arm enumerated every crash boundary; both arms found the outbox/idempotency design. |
| Bounded image workers with cancellation | 10/10 | 10/10 | 0 | Both arms replaced spawn-then-semaphore with a fixed, owned worker lifecycle. |
| Three-hop retry amplification and overload | 10/10 | 10/10 | 0 | Both arms correctly bounded deadlines, retries, replay, capacity, and recovery. |
| Ten discovery-only routing prompts | n/a | 10/10 | n/a | All prompts selected the intended skill or `NONE`, including adjacent HTTP/resilience and SQL/message boundaries. |

The observed quality delta is 1/30 criteria over one trial. The run file preserves normalized output rather than byte-exact harness artifacts, so it is not eligible as a published benchmark under the policy below. It is directionally useful but too small and under-sampled to establish broad superiority over the base agent. All v1 skills therefore remain `beta`.

## Running future evaluations

Select quality and routing cases from each `skills/<name>/evals.json`. For code-producing tasks, use a pinned repository fixture and an executable verifier in addition to semantic scoring. Record model and agent versions, repository commit, trial count, token use when available, failures, and full outputs. Reject changes that improve the preferred wording without changing engineering behavior.
