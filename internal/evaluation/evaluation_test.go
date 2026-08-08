package evaluation

import (
	"context"
	"encoding/json"
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

func TestScoreRoutingAcceptsIsolatedPluginNamespace(t *testing.T) {
	t.Parallel()
	score := scoreResult(Result{Kind: "routing", ExpectedRoutes: []string{"go"}, Response: "go:go"})
	if !score.Passed {
		t.Fatalf("score = %#v", score)
	}
}

func TestRunFixtureGradersExecutesGoTest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module fixture/test\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "fixture.go"), []byte("package fixture\n\nfunc Value() int { return 1 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repository := t.TempDir()
	oracleDir := filepath.Join(repository, "evaluations", "oracles", "fixture")
	if err := os.MkdirAll(oracleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oracleDir, "hidden_test.go"), []byte("package fixture\nimport \"testing\"\nfunc TestHidden(t *testing.T) { if Value() != 1 { t.Fatal(Value()) } }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "hidden_test.go")); !os.IsNotExist(err) {
		t.Fatalf("oracle visible before grading: %v", err)
	}
	runs := runFixtureGraders(context.Background(), repository, root, []corpus.Grader{{ID: "tests", Kind: "go-test", Target: "./...", Oracle: "evaluations/oracles/fixture/hidden_test.go"}})
	if !runs["tests"].Passed {
		t.Fatalf("grader run = %#v", runs["tests"])
	}
	info, err := os.Stat(filepath.Join(root, "hidden_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o444 {
		t.Fatalf("oracle mode = %o, want 444", info.Mode().Perm())
	}
}

func TestNormalizedGradersInfersLockedOracle(t *testing.T) {
	t.Parallel()
	graders := normalizedGraders(corpus.EvalCase{
		Fixture: "evaluations/fixtures/race-counter",
		Graders: []corpus.Grader{{ID: "race", Kind: "go-test", Target: "-race ./...", Weight: 1}},
	})
	if len(graders) != 1 || graders[0].Oracle != "evaluations/oracles/race-counter/hidden_test.go" {
		t.Fatalf("normalizedGraders() = %#v", graders)
	}
}

func TestInstallOracleRejectsNonCanonicalPathAndCollision(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	oracleDir := filepath.Join(repository, "evaluations", "oracles", "fixture")
	if err := os.MkdirAll(oracleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oracleDir, "hidden_test.go"), []byte("package fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	worktree := t.TempDir()
	if err := installOracle(repository, worktree, "evaluations/oracles/fixture/../fixture/hidden_test.go"); err == nil {
		t.Fatal("installOracle() accepted a non-canonical path")
	}
	if err := os.WriteFile(filepath.Join(worktree, "hidden_test.go"), []byte("agent-controlled\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installOracle(repository, worktree, "evaluations/oracles/fixture/hidden_test.go"); err == nil {
		t.Fatal("installOracle() overwrote an existing destination")
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

func TestParseUsage(t *testing.T) {
	t.Parallel()
	events := "warning\n{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":120,\"cached_input_tokens\":80,\"output_tokens\":7,\"reasoning_output_tokens\":3}}\n"
	usage := parseUsage(events)
	if usage.InputTokens != 120 || usage.CachedInputTokens != 80 || usage.OutputTokens != 7 || usage.ReasoningOutputTokens != 3 {
		t.Fatalf("usage = %#v", usage)
	}
}

func TestReportIncludesRoutingCollectionsAndTokens(t *testing.T) {
	t.Parallel()
	path := writeScoreFile(t, []Score{
		{Result: Result{Arm: "ours", Skill: "go-language-engineering", Collection: "engineering-skills-for-go", Kind: "routing", ExpectedRoutes: []string{"go-language-engineering"}, Response: "go-language-engineering", Usage: Usage{InputTokens: 100}}, Passed: true, Score: 1},
		{Result: Result{Arm: "ours", Skill: "go-language-engineering", Collection: "engineering-skills-for-go", Kind: "routing", ExpectedRoutes: []string{"NONE"}, Response: "go-language-engineering", Usage: Usage{InputTokens: 100}}, Score: 0},
		{Result: Result{Arm: "ours", Skill: "go-money-and-ledgers", Collection: "fintech-skills-for-go", Kind: "quality", Usage: Usage{InputTokens: 100}}, Score: 0},
	})
	report, err := ReportFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != 2 || report.Cases != 3 || report.Passed != 1 || report.CriticalFails != 1 {
		t.Fatalf("report = %#v", report)
	}
	if report.Routing.UnrelatedCases != 1 || report.Routing.FalseActivations != 1 || report.Routing.FalseActivationRate != 1 {
		t.Fatalf("routing = %#v", report.Routing)
	}
	if report.Tokens.InputTokens != 300 || report.Tokens.ScorePerKInputTokens == 0 {
		t.Fatalf("tokens = %#v", report.Tokens)
	}
}

func TestCompareFilesUsesCompletePairsAndHalfCreditTies(t *testing.T) {
	t.Parallel()
	candidate := writeScoreFile(t, []Score{
		{Result: Result{Arm: "ours", Skill: "go-language-engineering", CaseID: "one", Usage: Usage{InputTokens: 100}}, Passed: true, Score: 1},
		{Result: Result{Arm: "ours", Skill: "go-language-engineering", CaseID: "two", Usage: Usage{InputTokens: 100}}, Passed: true, Score: 1},
	})
	competitor := writeScoreFile(t, []Score{
		{Result: Result{Arm: "competitor", Skill: "go-language-engineering", CaseID: "one", Usage: Usage{InputTokens: 200}}, Score: 0},
		{Result: Result{Arm: "competitor", Skill: "go-language-engineering", CaseID: "two", Usage: Usage{InputTokens: 200}}, Passed: true, Score: 1},
	})
	report, err := CompareFiles(candidate, competitor)
	if err != nil {
		t.Fatal(err)
	}
	if !report.CompletePairs || report.PairedCases != 2 || report.CandidateWins != 1 || report.Ties != 1 || report.PairedWinRate != 0.75 {
		t.Fatalf("comparison = %#v", report)
	}
	if report.ParetoRelation != "candidate-dominates" {
		t.Fatalf("pareto relation = %q", report.ParetoRelation)
	}
}

func TestReportsSelectAndCompareArmsFromMixedArtifact(t *testing.T) {
	t.Parallel()
	path := writeScoreFile(t, []Score{
		{Result: Result{Arm: "ours", Skill: "go-language-engineering", CaseID: "one", Usage: Usage{InputTokens: 100}}, Passed: true, Score: 1},
		{Result: Result{Arm: "competitor", Skill: "go-language-engineering", CaseID: "one", Usage: Usage{InputTokens: 100}}, Score: 0},
	})
	if _, err := ReportFile(path); err == nil {
		t.Fatal("ReportFile() accepted a mixed artifact without an arm selector")
	}
	report, err := ReportFileForArm(path, "ours")
	if err != nil || report.Cases != 1 || report.Passed != 1 {
		t.Fatalf("ReportFileForArm() report = %#v, err = %v", report, err)
	}
	comparison, err := CompareArms(path, "ours", path, "competitor")
	if err != nil || !comparison.CompletePairs || comparison.CandidateWins != 1 {
		t.Fatalf("CompareArms() report = %#v, err = %v", comparison, err)
	}
}

func writeScoreFile(t *testing.T, scores []Score) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "scores.jsonl")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(file)
	for _, score := range scores {
		if err := encoder.Encode(score); err != nil {
			file.Close()
			t.Fatal(err)
		}
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
