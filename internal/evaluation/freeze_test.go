package evaluation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
)

func TestHoldoutCommitmentConcealsCasesAndDetectsMutation(t *testing.T) {
	t.Parallel()
	activate := true
	holdout := PrivateHoldout{SchemaVersion: 1, ID: "rc1-hidden", Cases: []PrivateHoldoutCase{{
		Skill: "go-service-resilience", TaskType: "diagnosis",
		Case: corpus.EvalCase{ID: "secret-recovery-herd", Kind: "routing", Split: "holdout", Prompt: "secret prompt words", ShouldActivate: &activate, Reason: "secret reason"},
	}}}
	collection := corpus.Collection{Skills: []corpus.Skill{{Name: "go-service-resilience", Metadata: corpus.Metadata{Collection: "distributed-systems-skills-for-go"}}}}
	key := []byte("01234567890123456789012345678901")
	commitment, err := commitHoldout(holdout, key, collection)
	if err != nil {
		t.Fatal(err)
	}
	public, err := json.Marshal(commitment)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"secret-recovery-herd", "secret prompt words", "secret reason"} {
		if strings.Contains(string(public), secret) {
			t.Fatalf("public commitment exposed %q: %s", secret, public)
		}
	}
	if commitment.Method != holdoutCommitmentMethod || commitment.CaseCount != 1 || len(commitment.Strata) != 1 {
		t.Fatalf("commitment = %#v", commitment)
	}

	mutated := holdout
	mutated.Cases = append([]PrivateHoldoutCase(nil), holdout.Cases...)
	mutated.Cases[0].Case.Prompt = "changed after treatment"
	changed, err := commitHoldout(mutated, key, collection)
	if err != nil {
		t.Fatal(err)
	}
	if changed.Digest == commitment.Digest {
		t.Fatal("holdout mutation did not change the commitment")
	}
}

func TestBindMatrixFreezeRejectsProtocolDrift(t *testing.T) {
	t.Parallel()
	lock := testFreeze()
	digest, err := BenchmarkFreezeDigest(lock)
	if err != nil {
		t.Fatal(err)
	}
	valid := MatrixOptions{
		Runner: "codex", Model: "gpt-test", Split: "development", Kind: "all",
		Repetition: 1, Seed: 102, Timeout: 2 * time.Minute,
	}
	if err := BindMatrixFreeze(lock, digest, &valid); err != nil {
		t.Fatal(err)
	}
	if valid.FreezeID != lock.ID || valid.FreezeDigest != digest {
		t.Fatalf("freeze identity was not bound: %#v", valid)
	}

	tests := map[string]func(*MatrixOptions){
		"model":      func(options *MatrixOptions) { options.Model = "changed" },
		"seed":       func(options *MatrixOptions) { options.Seed++ },
		"timeout":    func(options *MatrixOptions) { options.Timeout += time.Millisecond },
		"repetition": func(options *MatrixOptions) { options.Repetition = 2 },
		"split":      func(options *MatrixOptions) { options.Split = "all" },
		"limit":      func(options *MatrixOptions) { options.Limit = 1 },
	}
	for name, mutate := range tests {
		name, mutate := name, mutate
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options := MatrixOptions{
				Runner: "codex", Model: "gpt-test", Split: "development", Kind: "all",
				Repetition: 1, Seed: 102, Timeout: 2 * time.Minute,
			}
			mutate(&options)
			if err := BindMatrixFreeze(lock, digest, &options); err == nil {
				t.Fatal("BindMatrixFreeze() accepted protocol drift")
			}
		})
	}
}

func TestValidateResultArtifactFreezeRejectsForeignCells(t *testing.T) {
	t.Parallel()
	lock := testFreeze()
	digest, err := BenchmarkFreezeDigest(lock)
	if err != nil {
		t.Fatal(err)
	}
	result := Result{
		SchemaVersion: 1, OpaqueArm: opaqueLabel(101, "ours"), Arm: "ours", Runner: "codex",
		Model: "gpt-test", ClientVersion: "codex-test", Skill: "go-service-resilience",
		CaseID: "quality-herd", Mode: "native", Split: "development", Repetition: 0, Kind: "quality",
		Metadata: map[string]string{
			"freeze_id": lock.ID, "freeze_digest": digest, "treatment_seed": "101",
			"grader_version": runnerGraderVersion, "fixture_commit": strings.Repeat("a", 40),
		},
	}
	path := writeResultArtifact(t, result)
	if err := ValidateResultArtifactFreeze(path, lock, digest); err != nil {
		t.Fatal(err)
	}

	result.Metadata["freeze_digest"] = strings.Repeat("0", 64)
	path = writeResultArtifact(t, result)
	if err := ValidateResultArtifactFreeze(path, lock, digest); err == nil {
		t.Fatal("ValidateResultArtifactFreeze() accepted a foreign freeze digest")
	}
}

func TestCompletedCasesRetriesFailedAttemptAndSeparatesFreeze(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "run.jsonl")
	failed := Result{
		Arm: "ours", Skill: "skill", CaseID: "case", Mode: "native", Error: "timeout",
		Metadata: map[string]string{"freeze_digest": "freeze-a"},
	}
	writeFreezeJSONLines(t, path, failed)
	completed, err := completedCases(path)
	if err != nil {
		t.Fatal(err)
	}
	if completed[benchmarkKeyWithFreeze("ours", "skill", "case", "native", 0, "freeze-a")] {
		t.Fatal("failed attempt suppressed its retry")
	}

	failed.Error = ""
	writeFreezeJSONLines(t, path, failed)
	completed, err = completedCases(path)
	if err != nil {
		t.Fatal(err)
	}
	if !completed[benchmarkKeyWithFreeze("ours", "skill", "case", "native", 0, "freeze-a")] {
		t.Fatal("successful retry was not resumable")
	}
	if completed[benchmarkKeyWithFreeze("ours", "skill", "case", "native", 0, "freeze-b")] {
		t.Fatal("one freeze suppressed a cell from another freeze")
	}
}

func TestDigestPathIsStableAndDetectsContentAndMode(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "case.txt")
	if err := os.WriteFile(path, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	first, err := digestPath(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := digestPath(root)
	if err != nil || first != second {
		t.Fatalf("stable digest = %q, %q, %v", first, second, err)
	}
	if err := os.WriteFile(path, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	contentDigest, err := digestPath(root)
	if err != nil || contentDigest == first {
		t.Fatalf("content digest = %q, want change from %q, err = %v", contentDigest, first, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	modeDigest, err := digestPath(root)
	if err != nil || modeDigest == contentDigest {
		t.Fatalf("mode digest = %q, want change from %q, err = %v", modeDigest, contentDigest, err)
	}
}

func TestReadHoldoutKeyRequiresOwnerOnlyLowercaseHex(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "holdout.key")
	if err := os.WriteFile(path, []byte(strings.Repeat("ab", 32)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	key, err := readHoldoutKey(path, true)
	if err != nil || len(key) != 32 {
		t.Fatalf("readHoldoutKey() length = %d, err = %v", len(key), err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readHoldoutKey(path, true); err == nil {
		t.Fatal("readHoldoutKey() accepted group/world-readable key")
	}
	if key, err := readHoldoutKey(path, false); err != nil || len(key) != 32 {
		t.Fatalf("published key verification length = %d, err = %v", len(key), err)
	}
}

func testFreeze() BenchmarkFreeze {
	return BenchmarkFreeze{
		SchemaVersion: 1, ID: "rc1", CreatedOn: "2026-08-24", CreatedFromCommit: strings.Repeat("a", 40),
		Protocol: FreezeProtocol{
			Runner: "codex", Model: "gpt-test", ClientVersion: "codex-test", ToolchainVersion: "go-test",
			JudgeRunner: "codex", JudgeModel: "gpt-judge", Modes: []string{"explicit", "native"},
			TreatmentSeeds: []int64{101, 102}, JudgeSeeds: []int64{201, 202},
			TreatmentTimeoutMS: int64((2 * time.Minute).Milliseconds()), JudgmentTimeoutMS: int64((3 * time.Minute).Milliseconds()),
			RunnerGraderVersion: runnerGraderVersion, RubricVersion: rubricVersion,
			DeterministicScorer: "skillctl-eval-v2", SemanticScorer: "skillctl-eval-v3-semantic",
		},
		Inputs: []FrozenInput{{Path: "skills", SHA256: strings.Repeat("b", 64)}},
		Arms:   []FrozenArm{{Name: "baseline"}, {Name: "ours", SkillsSHA256: strings.Repeat("c", 64)}},
		PublicCases: []FrozenCase{{
			Key: "go-service-resilience/quality-herd", Collection: "distributed-systems-skills-for-go",
			Kind: "quality", Split: "development", SHA256: strings.Repeat("d", 64),
		}},
	}
}

func writeResultArtifact(t *testing.T, result Result) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "result.jsonl")
	writeFreezeJSONLines(t, path, result)
	return path
}

func writeFreezeJSONLines(t *testing.T, path string, values ...any) {
	t.Helper()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
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

func TestBenchmarkFreezeDigestIsDeterministic(t *testing.T) {
	t.Parallel()
	lock := testFreeze()
	first, err := BenchmarkFreezeDigest(lock)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BenchmarkFreezeDigest(lock)
	if err != nil || first != second {
		t.Fatalf("digests = %q and %q, err = %v", first, second, err)
	}
	mutated := lock
	mutated.Protocol.Model = "another-model"
	third, err := BenchmarkFreezeDigest(mutated)
	if err != nil || reflect.DeepEqual(first, third) {
		t.Fatalf("mutated digest = %q, original = %q, err = %v", third, first, err)
	}
}
