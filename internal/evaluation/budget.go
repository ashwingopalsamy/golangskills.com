package evaluation

import "fmt"

type budgetDecision uint8

const (
	budgetContinue budgetDecision = iota
	budgetStopAttempts
	budgetStopFailures
	budgetStopTimeout
)

// attemptBudget bounds newly persisted attempts without interpreting a valid
// candidate score. A zero limit is intentionally unlimited.
type attemptBudget struct {
	maxAttempts int
	maxFailures int
	stopTimeout bool
	written     int
	failures    int
}

func newAttemptBudget(maxAttempts, maxFailures int, stopTimeout bool) (*attemptBudget, error) {
	if maxAttempts < 0 {
		return nil, fmt.Errorf("maximum attempts must be non-negative")
	}
	if maxFailures < 0 {
		return nil, fmt.Errorf("maximum failures must be non-negative")
	}
	return &attemptBudget{maxAttempts: maxAttempts, maxFailures: maxFailures, stopTimeout: stopTimeout}, nil
}

func (budget *attemptBudget) record(failed, timedOut bool) budgetDecision {
	budget.written++
	if failed {
		budget.failures++
		if timedOut && budget.stopTimeout {
			return budgetStopTimeout
		}
		if budget.maxFailures > 0 && budget.failures >= budget.maxFailures {
			return budgetStopFailures
		}
	}
	if budget.maxAttempts > 0 && budget.written >= budget.maxAttempts {
		return budgetStopAttempts
	}
	return budgetContinue
}
