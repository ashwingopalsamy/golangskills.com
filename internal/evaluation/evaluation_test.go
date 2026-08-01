package evaluation

import (
	"testing"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
)

func TestScoreRouting(t *testing.T) {
	t.Parallel()
	score := scoreResult(Result{Kind: "routing", ExpectedRoute: "go-data-consistency", Response: "go-data-consistency"})
	if !score.Passed || score.Score != 1 {
		t.Fatalf("score = %#v", score)
	}
}

func TestScoreDeterministicGrader(t *testing.T) {
	t.Parallel()
	score := scoreResult(Result{
		Kind: "quality", Response: "Use a unique constraint and one transaction.",
		Graders: []corpus.Grader{{ID: "atomic", Kind: "contains", Required: []string{"unique", "transaction"}, Weight: 1}},
	})
	if !score.Passed || !score.GraderScore["atomic"] {
		t.Fatalf("score = %#v", score)
	}
}

func TestWilsonIntervalContainsObservedRate(t *testing.T) {
	t.Parallel()
	lower, upper := wilson(60, 100)
	if lower >= 0.6 || upper <= 0.6 {
		t.Fatalf("interval = [%f, %f], want it to contain 0.6", lower, upper)
	}
}
