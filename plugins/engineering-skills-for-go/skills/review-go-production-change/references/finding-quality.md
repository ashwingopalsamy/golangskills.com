# Finding quality

## A finding must be falsifiable

Another engineer should be able to disprove a finding by showing one missing premise: the call cannot occur, access cannot overlap, a constraint rejects the input, a transaction includes the effect, or rollout never mixes versions. Include enough of the causal path for that evaluation.

## Tight locations

Attach the finding to the changed line that introduces the behavior, even when impact occurs later. Use the body to name the downstream path. Avoid file-wide ranges and comments on unchanged code unless the change makes the old code newly unsafe.

## Severity calibration

Use worst credible impact with plausible triggering conditions. A panic in an offline developer tool is not automatically P1. A one-line retry bug that can duplicate financial effects may be P1. Uncertainty about reachability lowers confidence, not necessarily impact; investigate before reporting.

## Correction scope

Recommend the smallest invariant-restoring change. Do not smuggle a refactor into a bug fix. When several designs are valid, state the required property rather than mandating a framework.

## Examples

Weak:

> Add a timeout here as a best practice.

Actionable:

> This new background-context call can outlive the request and holds one of the eight database connections. When the dependency stalls, canceled requests accumulate until every handler waits on the pool. Pass the request context into `Lookup` or derive a bounded child deadline at this ownership layer.

Weak:

> This may have a race.

Actionable:

> `Stop` closes `jobs` while `Submit` can still pass the unlocked `stopped` check and send. That interleaving panics with “send on closed channel.” Serialize admission and close under one owner, or stop closing a channel with concurrent senders.
