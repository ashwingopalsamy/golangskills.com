package commitambiguity

import "testing"

func TestDecisionDistinguishesKnownAndAmbiguousOutcomes(t *testing.T) {
	tests := []struct {
		name         string
		responseLost bool
		found        bool
		want         Decision
	}{
		{name: "confirmed after lost response", responseLost: true, found: true, want: Committed},
		{name: "unknown after lost response", responseLost: true, found: false, want: Reconcile},
		{name: "confirmed normally", found: true, want: Committed},
		{name: "known absent", want: Retry},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Decide(test.responseLost, test.found); got != test.want {
				t.Fatalf("Decide(%v, %v) = %q, want %q", test.responseLost, test.found, got, test.want)
			}
		})
	}
}
