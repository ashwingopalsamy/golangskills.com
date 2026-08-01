# Payment machine

For every command record:

- local operation ID and provider idempotency key;
- requested amount/currency and canonical fingerprint;
- provider request ID and returned object/version;
- known, failed, or ambiguous outcome;
- events that can resolve ambiguity;
- next legal actions and amount limits.

Late success after local timeout must advance from persisted evidence, not be rejected because a synchronous request already returned. Cancellation and reversal are not interchangeable; follow the rail's actual semantics.
