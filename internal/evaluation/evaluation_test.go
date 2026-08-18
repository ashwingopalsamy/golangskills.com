package evaluation

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
)

func TestScoreRouting(t *testing.T) {
	t.Parallel()
	score := scoreResult(Result{Kind: "routing", ExpectedRoutes: []string{"go-data-consistency", "go-message-processing"}, Response: "go-message-processing"})
	if !score.Passed || score.Score != 1 {
		t.Fatalf("score = %#v", score)
	}
}

func TestRunFixtureGradersExecutesGoTest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module fixture/test\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "fixture_test.go"), []byte("package fixture_test\nimport \"testing\"\nfunc TestPass(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runs := runFixtureGraders(context.Background(), root, []corpus.Grader{{ID: "tests", Kind: "go-test", Target: "./..."}})
	if !runs["tests"].Passed {
		t.Fatalf("grader run = %#v", runs["tests"])
	}
}

func TestCopySkillsRejectsMissingExplicitMapping(t *testing.T) {
	t.Parallel()
	source := t.TempDir()
	err := copySkills(source, filepath.Join(t.TempDir(), "skills"), true, "missing-skill")
	if err == nil {
		t.Fatal("copySkills() succeeded for a missing explicit mapping")
	}
}

func TestValidateArmMapsRequiresEverySelectedCompetitorCase(t *testing.T) {
	t.Parallel()
	activate := true
	cases := []caseItem{{skill: corpus.Skill{Name: "go-data-consistency"}, eval: corpus.EvalCase{ID: "route-commit", Kind: "routing", ShouldActivate: &activate}}}
	err := validateArmMaps(RunOptions{Arm: "competitor", RoutingMap: map[string][]string{"another/case": {"skill"}}}, cases)
	if err == nil {
		t.Fatal("validateArmMaps() accepted an incomplete competitor routing map")
	}
}

func TestExpectedRoutesMapsCanonicalSkillForCompetitor(t *testing.T) {
	t.Parallel()
	activate := true
	item := caseItem{skill: corpus.Skill{Name: "go-data-consistency"}, eval: corpus.EvalCase{ID: "route-commit", ShouldActivate: &activate}}
	routes := expectedRoutes(RunOptions{Arm: "competitor", SkillMap: map[string]string{"go-data-consistency": "golang-database"}}, item)
	if len(routes) != 1 || routes[0] != "golang-database" {
		t.Fatalf("expectedRoutes() = %v", routes)
	}
}

func TestExpectedRoutesUsesDeclaredConfusionTarget(t *testing.T) {
	t.Parallel()
	activate := false
	item := caseItem{skill: corpus.Skill{Name: "go-message-processing"}, eval: corpus.EvalCase{
		ID: "avoid-channel", ShouldActivate: &activate, ConfusesWith: []string{"go-concurrency-lifecycle"},
	}}
	routes := expectedRoutes(RunOptions{Arm: "ours"}, item)
	if len(routes) != 1 || routes[0] != "go-concurrency-lifecycle" {
		t.Fatalf("expectedRoutes() = %v", routes)
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
