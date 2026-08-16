package evaluation

import "testing"

func TestAttemptBudgetStopsWithSafetyPrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		maxAttempts int
		maxFailures int
		stopTimeout bool
		failed      bool
		timedOut    bool
		want        budgetDecision
	}{
		{name: "successful pilot cap", maxAttempts: 1, want: budgetStopAttempts},
		{name: "failure before attempt cap", maxAttempts: 1, maxFailures: 1, failed: true, want: budgetStopFailures},
		{name: "timeout before failure cap", maxAttempts: 1, maxFailures: 1, stopTimeout: true, failed: true, timedOut: true, want: budgetStopTimeout},
		{name: "timeout can be recorded without stopping", stopTimeout: false, failed: true, timedOut: true, want: budgetContinue},
		{name: "zero limits are unlimited", want: budgetContinue},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			budget, err := newAttemptBudget(test.maxAttempts, test.maxFailures, test.stopTimeout)
			if err != nil {
				t.Fatal(err)
			}
			if got := budget.record(test.failed, test.timedOut); got != test.want {
				t.Fatalf("record() = %v, want %v", got, test.want)
			}
			if budget.written != 1 {
				t.Fatalf("written = %d, want 1", budget.written)
			}
			wantFailures := 0
			if test.failed {
				wantFailures = 1
			}
			if budget.failures != wantFailures {
				t.Fatalf("failures = %d, want %d", budget.failures, wantFailures)
			}
		})
	}
}

func TestAttemptBudgetRejectsNegativeLimits(t *testing.T) {
	t.Parallel()
	if _, err := newAttemptBudget(-1, 0, true); err == nil {
		t.Fatal("newAttemptBudget() accepted negative attempt limit")
	}
	if _, err := newAttemptBudget(0, -1, true); err == nil {
		t.Fatal("newAttemptBudget() accepted negative failure limit")
	}
}
