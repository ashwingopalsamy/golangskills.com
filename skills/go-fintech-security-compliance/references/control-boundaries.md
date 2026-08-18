# Control boundaries

- Tokenization reduces scope only when detokenization and routing paths remain isolated.
- Encryption does not remove scope when the same service controls ciphertext and keys.
- Authentication does not authorize a tenant, resource, amount, refund, or administrative action.
- Audit records are weak when privileged actors can silently edit or delete them.
- Data minimization includes logs, traces, retries, DLQs, backups, support exports, and model prompts.
- Retention needs both deletion behavior and evidence that deletion completed across replicas and backups under policy.
