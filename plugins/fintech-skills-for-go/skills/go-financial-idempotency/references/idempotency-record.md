# Idempotency record

Minimum fields commonly include tenant/principal scope, operation kind, key hash, request fingerprint, status, result reference, provider identity, created/updated/expiry time, and version.

The unique index must cover the full semantic scope. Concurrent callers either observe the committed result or a defined in-progress response. A failed validation before effect initiation may be retryable; a provider timeout after transmission is ambiguous, not a clean failure.
