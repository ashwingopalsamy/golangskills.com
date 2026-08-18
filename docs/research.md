# Research record

Research was performed on 2026-08-18 before implementation.

## Standards and product behavior

- Open Agent Skills specification: <https://agentskills.io/specification>
- OpenAI skill authoring: <https://learn.chatgpt.com/docs/build-skills>
- OpenAI plugin skills: <https://developers.openai.com/plugins/build/skills>
- Claude Code skills: <https://code.claude.com/docs/en/slash-commands>
- GitHub Copilot skills: <https://docs.github.com/en/copilot/how-tos/copilot-customization/custom-instructions/add-agent-skills>
- Gemini CLI skills: <https://geminicli.com/docs/cli/skills/>
- OpenCode skills: <https://opencode.ai/docs/skills/>
- Roo Code skills: <https://docs.roocode.com/features/skills>
- Windsurf Cascade skills: <https://docs.windsurf.com/windsurf/cascade/skills>
- Go brand and trademark guidelines: <https://go.dev/brand>

## Go and production sources

- Current Go downloads: <https://go.dev/dl/>
- Go memory model: <https://go.dev/ref/mem>
- Context package: <https://pkg.go.dev/context>
- Race detector: <https://go.dev/doc/articles/race_detector>
- Go security guidance: <https://go.dev/doc/security/best-practices>
- `net/http`: <https://pkg.go.dev/net/http>
- Database transactions: <https://go.dev/doc/database/execute-transactions>
- Database connection management: <https://go.dev/doc/database/manage-connections>
- Google SRE cascading failures: <https://sre.google/sre-book/addressing-cascading-failures/>
- AWS timeouts, retries, backoff, and jitter: <https://aws.amazon.com/builders-library/timeouts-retries-and-backoff-with-jitter/>
- Google Cloud Pub/Sub exactly-once scope: <https://cloud.google.com/pubsub/docs/exactly-once-delivery>
- Stripe idempotent requests: <https://docs.stripe.com/api/idempotent_requests>

Each skill’s `skill.json` narrows this list to the claims it uses and records authority kind and verification date.

## Empirical evidence

- SkillsBench found large but variable gains for curated skills, with negative deltas on some tasks and focused modules outperforming comprehensive documentation: <https://arxiv.org/abs/2602.12670>
- SWE-Skills-Bench found 39 of 49 software-engineering skills produced no pass-rate gain and version-mismatched guidance could degrade performance: <https://arxiv.org/abs/2603.15401>
- A 138,133-skill study found weak routing metadata, bloated bodies, and poor resource organization dominated reusability defects: <https://arxiv.org/abs/2608.08453>
- A security study found skill scripts correlated with higher vulnerability rates, supporting instruction-only defaults: <https://arxiv.org/abs/2601.10338>

These studies motivate the evaluation and security policy; their percentages are not treated as timeless product guarantees.
