package evaluation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
)

func TestSemanticJudgePromptBlindsArmAndSkillIdentity(t *testing.T) {
	t.Parallel()
	result := Result{
		Arm: "gophers", OpaqueArm: "arm-secret", Skill: "canonical-secret-skill",
		Prompt:       "Review this design for correctness.",
		Response:     "GOPHERS and MURATMIRGUN/GOPHERS used COMPETITOR-SECRET-SKILL and $another-secret-skill to preserve the invariant.",
		RunnerEvents: `.agents/skills/competitor-secret-skill/SKILL.md`,
		Invariants:   []string{"Gophers preserves the invariant"},
		Forbidden:    []string{"Muratmirgun allows duplicate effects"},
	}
	prompt, err := semanticJudgePrompt(judgmentCandidate{result: result, blindLabel: "candidate-123"})
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"gophers", "Gophers", "GOPHERS", "arm-secret", "canonical-secret-skill", "COMPETITOR-SECRET-SKILL", "another-secret-skill", "MURATMIRGUN/GOPHERS", "Muratmirgun", "MURATMIRGUN"} {
		if strings.Contains(prompt, secret) {
			t.Errorf("semanticJudgePrompt() contains identity %q", secret)
		}
	}
	if !strings.Contains(prompt, "candidate-123") || !strings.Contains(prompt, "preserves the invariant") {
		t.Errorf("semanticJudgePrompt() omitted blinded label or rubric: %s", prompt)
	}
	for _, required := range []string{"forbidden_endorsed", "quote, describe, identify, or condemn", "never the task's code"} {
		if !strings.Contains(prompt, required) {
			t.Errorf("semanticJudgePrompt() omitted forbidden-outcome polarity rule %q", required)
		}
	}
}

func TestDecodeRubricVerdictFailsClosed(t *testing.T) {
	t.Parallel()
	result := Result{Invariants: []string{"one", "two"}, Forbidden: []string{"bad"}}
	tests := []struct {
		name    string
		content string
	}{
		{name: "wrong invariant count", content: `{"invariants":[true],"forbidden_endorsed":[false],"evidence":["a","b"],"summary":"ok"}`},
		{name: "unknown field", content: `{"invariants":[true,true],"forbidden_endorsed":[false],"evidence":["a","b","c"],"summary":"ok","score":1}`},
		{name: "trailing value", content: `{"invariants":[true,true],"forbidden_endorsed":[false],"evidence":["a","b","c"],"summary":"ok"} {}`},
		{name: "empty evidence", content: `{"invariants":[true,true],"forbidden_endorsed":[false],"evidence":["a","","c"],"summary":"ok"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var verdict RubricVerdict
			if err := decodeRubricVerdict([]byte(test.content), result, &verdict); err == nil {
				t.Errorf("decodeRubricVerdict(%q) succeeded, want error", test.name)
			}
		})
	}
}

func TestComputeRubricScoreUsesForbiddenOutcomesAsCritical(t *testing.T) {
	t.Parallel()
	score, passed, critical := computeRubricScore(RubricVerdict{
		Invariants: []bool{true, false}, ForbiddenEndorsed: []bool{true},
	})
	if score != 1.0/3.0 || passed || !critical {
		t.Errorf("computeRubricScore() = (%v, %v, %v), want (%v, false, true)", score, passed, critical, 1.0/3.0)
	}
}

func TestApplyJudgmentRejectsSourceDigestMismatch(t *testing.T) {
	t.Parallel()
	result := Result{
		RunID: "run", Arm: "ours", Runner: "codex", Skill: "go-data-consistency",
		CaseID: "quality", Mode: "native", Kind: "quality", Prompt: "task", Response: "answer",
		Invariants: []string{"one"}, Forbidden: []string{"bad"},
		Graders: []corpus.Grader{{ID: "terms", Kind: "contains", Required: []string{"answer"}, Weight: 1}},
	}
	verdict := RubricVerdict{Invariants: []bool{true}, ForbiddenEndorsed: []bool{false}, Evidence: []string{"yes", "absent"}, Summary: "ok"}
	judgmentID := semanticJudgmentID(result, "codex", "judge-model")
	judgment := Judgment{
		SchemaVersion: judgmentSchemaVersion, JudgmentID: judgmentID, Arm: result.Arm, Skill: result.Skill,
		CaseID: result.CaseID, Mode: result.Mode, Runner: "codex", Model: "judge-model",
		RubricVersion: rubricVersion, SamePlatform: true, Verdict: verdict,
		Score: 1, Passed: true, Metadata: map[string]string{"source_digest": "wrong", "evaluator_commit": strings.Repeat("a", 40)},
	}
	score := scoreResult(result)
	applyJudgment(&score, judgment)
	if score.Passed || score.Semantic == nil || score.Semantic.Passed || !strings.Contains(strings.Join(score.Failures, " "), "source digest mismatch") {
		t.Errorf("applyJudgment() score = %#v, want fail-closed digest rejection", score)
	}
}

func TestApplyJudgmentRejectsMissingEvaluatorCommit(t *testing.T) {
	t.Parallel()
	result := Result{
		Arm: "ours", Runner: "codex", Skill: "go-data-consistency", CaseID: "quality", Kind: "quality",
		Prompt: "task", Response: "answer", Invariants: []string{"one"}, Forbidden: []string{"bad"},
	}
	verdict := RubricVerdict{
		Invariants: []bool{true}, ForbiddenEndorsed: []bool{false},
		Evidence: []string{"present", "absent"}, Summary: "all criteria pass",
	}
	judgment := validJudgment(result, "codex", "judge-model", verdict)
	delete(judgment.Metadata, "evaluator_commit")
	score := scoreResult(result)
	applyJudgment(&score, judgment)
	if score.Passed || !strings.Contains(strings.Join(score.Failures, " "), "evaluator commit missing or invalid") {
		t.Errorf("applyJudgment() score = %#v, want fail-closed evaluator commit rejection", score)
	}
}

func TestScoreFileWithJudgmentsRejectsMissingRequiredJudgment(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	inputPath := filepath.Join(root, "raw.jsonl")
	judgmentPath := filepath.Join(root, "judgments.jsonl")
	outputPath := filepath.Join(root, "scores.jsonl")
	result := Result{
		Arm: "ours", Runner: "codex", Skill: "go-data-consistency", CaseID: "quality", Kind: "quality",
		Prompt: "task", Response: "answer", Invariants: []string{"one"}, Forbidden: []string{"bad"},
		Graders: []corpus.Grader{{ID: "terms", Kind: "contains", Required: []string{"literal-not-present"}, Weight: 1}},
	}
	content, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputPath, append(content, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	placeholder := Judgment{SchemaVersion: judgmentSchemaVersion, JudgmentID: "unrelated", Runner: "codex", Model: "judge-model"}
	content, err = json.Marshal(placeholder)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(judgmentPath, append(content, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ScoreFileWithJudgments(inputPath, judgmentPath, outputPath); err != nil {
		t.Fatal(err)
	}
	scores, err := readScores(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(scores) != 1 || scores[0].Passed || !strings.Contains(strings.Join(scores[0].Failures, " "), "semantic judgment missing") {
		t.Errorf("ScoreFileWithJudgments() = %#v, want missing-judgment failure", scores)
	}
}

func TestScoreFileWithJudgmentsMergesValidSemanticScore(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	inputPath := filepath.Join(root, "raw.jsonl")
	judgmentPath := filepath.Join(root, "judgments.jsonl")
	outputPath := filepath.Join(root, "scores.jsonl")
	result := Result{
		Arm: "ours", Runner: "codex", Skill: "go-data-consistency", CaseID: "quality", Kind: "quality",
		Prompt: "task", Response: "answer", Invariants: []string{"one"}, Forbidden: []string{"bad"},
		Graders: []corpus.Grader{{ID: "terms", Kind: "contains", Required: []string{"literal-not-present"}, Weight: 1}},
	}
	writeJSONLine(t, inputPath, result)
	verdict := RubricVerdict{
		Invariants: []bool{true}, ForbiddenEndorsed: []bool{false},
		Evidence: []string{"present", "absent"}, Summary: "all criteria pass",
	}
	judgment := validJudgment(result, "codex", "judge-model", verdict)
	writeJSONLine(t, judgmentPath, judgment)

	if _, err := ScoreFileWithJudgments(inputPath, judgmentPath, outputPath); err != nil {
		t.Fatal(err)
	}
	scores, err := readScores(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(scores) != 1 || !scores[0].Passed || scores[0].Score != 1 || scores[0].GraderScore["terms"] || len(scores[0].Failures) != 0 || scores[0].ScorerVersion != "skillctl-eval-v3-semantic" {
		t.Errorf("ScoreFileWithJudgments() = %#v, want valid semantic score", scores)
	}
}

func TestRequiresSemanticJudgmentSkipsFailedRunnerAndExecutableOracle(t *testing.T) {
	t.Parallel()
	base := Result{Kind: "quality", Invariants: []string{"one"}}
	if !requiresSemanticJudgment(base) {
		t.Fatal("requiresSemanticJudgment() rejected eligible quality case")
	}
	failed := base
	failed.ExitCode = 1
	if requiresSemanticJudgment(failed) {
		t.Fatal("requiresSemanticJudgment() accepted failed runner result")
	}
	executable := base
	executable.Graders = []corpus.Grader{{Kind: "go-test"}}
	if requiresSemanticJudgment(executable) {
		t.Fatal("requiresSemanticJudgment() accepted executable oracle case")
	}
}

func TestJudgmentRetriesRemainResumableAndLatestSuccessWins(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "judgments.jsonl")
	result := Result{
		Arm: "ours", Runner: "codex", Skill: "go-data-consistency", CaseID: "quality", Kind: "quality",
		Prompt: "task", Response: "answer", Invariants: []string{"one"}, Forbidden: []string{"bad"},
	}
	verdict := RubricVerdict{
		Invariants: []bool{true}, ForbiddenEndorsed: []bool{false},
		Evidence: []string{"present", "absent"}, Summary: "all criteria pass",
	}
	success := validJudgment(result, "codex", "judge-model", verdict)
	failure := success
	failure.Error = "temporary evaluator failure"
	failure.Verdict = RubricVerdict{}
	failure.Score, failure.Passed = 0, false
	writeJSONLines(t, path, failure, success)

	completed, err := completedJudgments(path)
	if err != nil {
		t.Fatal(err)
	}
	if !completed[success.JudgmentID] {
		t.Fatal("completedJudgments() did not mark the successful retry complete")
	}
	index, err := readJudgments(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := index.byID[success.JudgmentID]; got.Error != "" || !got.Passed {
		t.Errorf("readJudgments() retained %#v, want latest successful retry", got)
	}
}

func TestSemanticSourceDigestBindsRubricArrayBoundaries(t *testing.T) {
	t.Parallel()
	left := Result{Prompt: "task", Response: "answer", Invariants: []string{"one"}, Forbidden: []string{"two"}}
	right := Result{Prompt: "task", Response: "answer", Invariants: []string{"one", "two"}}
	if semanticSourceDigest(left) == semanticSourceDigest(right) {
		t.Fatal("semanticSourceDigest() does not bind invariant and forbidden array boundaries")
	}
}

func validJudgment(result Result, runner, model string, verdict RubricVerdict) Judgment {
	score, passed, critical := computeRubricScore(verdict)
	return Judgment{
		SchemaVersion: judgmentSchemaVersion, JudgmentID: semanticJudgmentID(result, runner, model),
		Arm: result.Arm, Skill: result.Skill, CaseID: result.CaseID, Mode: result.Mode,
		Runner: runner, Model: model, RubricVersion: rubricVersion,
		SamePlatform: strings.EqualFold(result.Runner, runner), Verdict: verdict,
		Score: score, Passed: passed, Critical: critical,
		Metadata: map[string]string{"source_digest": semanticSourceDigest(result), "evaluator_commit": strings.Repeat("a", 40)},
	}
}

func writeJSONLine(t *testing.T, path string, value any) {
	t.Helper()
	writeJSONLines(t, path, value)
}

func writeJSONLines(t *testing.T, path string, values ...any) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(file)
	for _, value := range values {
		if err := encoder.Encode(value); err != nil {
			file.Close()
			t.Fatal(err)
		}
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
