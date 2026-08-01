# Security

Agent Skills are instructions injected into agents that may already have filesystem, shell, network, or account access. Treat skill changes like executable supply-chain changes even when they contain only Markdown.

This repository’s published skills contain no executable scripts, remote fetch instructions, credential requirements, or broad tool grants. References are documentation links, never instructions to download and execute content. Generated plugin packages must be byte-equivalent to canonical skill source.

Review contributions for prompt injection, hidden Unicode, data-exfiltration instructions, privilege expansion, unsafe interpolation, unverifiable remote dependencies, and claims that bypass repository or user authority.

Report a vulnerability privately to the repository owner through GitHub Security Advisories. Do not include secrets, credentials, or unrelated personal data in a report.
