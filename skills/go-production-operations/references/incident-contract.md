# Incident contract

An operator should be able to distinguish:

- no admission versus slow processing;
- caller cancellation versus internal deadline;
- dependency rejection versus local overload;
- retry attempts versus original operations;
- known failure versus ambiguous outcome;
- unready drain versus crashed process;
- dropped telemetry versus no events.

During recovery, ramp traffic, retries, reconnects, and cache fill so the recovery path does not reproduce overload.
