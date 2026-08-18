---
name: go-fintech-security-compliance
description: "Use for Go payment-data scope, tokenization, authorization, audit, retention, and PCI controls. Do not use as certification."
license: Apache-2.0
compatibility: "Go 1.24 or newer; standards and regulatory obligations require current qualified interpretation for the actual system."
---

# Go fintech security and compliance

Minimize the systems and people that can store, process, transmit, or affect sensitive payment data. Compliance evidence must follow actual controls, not documentation theater.

## Scope the data

Classify account data, sensitive authentication data, PII, secrets, tokens, derived identifiers, and audit records. Prefer hosted collection or tokenization that prevents raw data from entering the service. Map every store, log, queue, trace, backup, analytics stream, and support tool the data can reach.

## Enforce authority

Authenticate users and services at an assurance level matching risk. Authorize by subject, tenant, action, resource, amount, and workflow state. Separate maker/checker or privileged approval where required. Rotate and revoke credentials; isolate cryptographic keys and record key version without logging secret material.

## Preserve auditability

Record actor, authority, operation, target, before/after state reference, reason, time, correlation, and approval in tamper-evident durable storage. Keep logs and traces free of PAN, authentication data, tokens, secrets, and unnecessary PII. Restrict access and test evidence retrieval.

## Engineer change controls

Threat-model sensitive changes, review generated artifacts and dependencies, pin CI inputs, separate duties for releases, prove rollback/migration behavior, and keep retention/deletion schedules enforceable. A checklist does not certify PCI DSS or any regulation; involve qualified security, compliance, and legal owners.

Read [references/control-boundaries.md](references/control-boundaries.md) for scope and audit failure cases.

## Output contract

State data classification, trust boundary, authority decision, control owner, evidence, and residual compliance uncertainty. Never claim certification from code review.
