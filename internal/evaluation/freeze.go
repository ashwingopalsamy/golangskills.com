package evaluation

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
)

const (
	freezeSchemaVersion     = 1
	holdoutSchemaVersion    = 1
	holdoutCommitmentMethod = "hmac-sha256-v1"
)

var freezeIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:[.-][a-z0-9]+)*$`)

// BenchmarkFreeze is the public, commit-safe release-candidate input lock. It
// contains no private holdout prompt, case identifier, or commitment key.
type BenchmarkFreeze struct {
	SchemaVersion     int                `json:"schema_version"`
	ID                string             `json:"id"`
	CreatedOn         string             `json:"created_on"`
	CreatedFromCommit string             `json:"created_from_commit"`
	Protocol          FreezeProtocol     `json:"protocol"`
	Inputs            []FrozenInput      `json:"inputs"`
	Arms              []FrozenArm        `json:"arms"`
	PublicCases       []FrozenCase       `json:"public_cases"`
	PrivateHoldout    *HoldoutCommitment `json:"private_holdout,omitempty"`
}

// FreezeProtocol captures settings that can change benchmark behavior or
// judgment reuse. Every repetition has a precommitted treatment and judge seed.
type FreezeProtocol struct {
	Runner              string   `json:"runner"`
	Model               string   `json:"model"`
	ClientVersion       string   `json:"client_version"`
	ToolchainVersion    string   `json:"toolchain_version"`
	JudgeRunner         string   `json:"judge_runner"`
	JudgeModel          string   `json:"judge_model"`
	Modes               []string `json:"modes"`
	TreatmentSeeds      []int64  `json:"treatment_seeds"`
	JudgeSeeds          []int64  `json:"judge_seeds"`
	TreatmentTimeoutMS  int64    `json:"treatment_timeout_ms"`
	JudgmentTimeoutMS   int64    `json:"judgment_timeout_ms"`
	RunnerGraderVersion string   `json:"runner_grader_version"`
	RubricVersion       string   `json:"rubric_version"`
	DeterministicScorer string   `json:"deterministic_scorer_version"`
	SemanticScorer      string   `json:"semantic_scorer_version"`
}

// FrozenInput is a digest of one repository input tree or file.
type FrozenInput struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// FrozenArm binds an opaque treatment to the exact installed skill bytes. The
// skills root is repository-relative so another checkout can reproduce it.
type FrozenArm struct {
	Name             string `json:"name"`
	RepositoryCommit string `json:"repository_commit,omitempty"`
	SkillsRoot       string `json:"skills_root,omitempty"`
	SkillsSHA256     string `json:"skills_sha256,omitempty"`
	SkillMap         string `json:"skill_map,omitempty"`
	SkillMapSHA256   string `json:"skill_map_sha256,omitempty"`
}

// FrozenCase makes the public case inventory reviewable without relying only
// on a whole-tree digest.
type FrozenCase struct {
	Key        string `json:"key"`
	Collection string `json:"collection"`
	Kind       string `json:"kind"`
	Split      string `json:"split"`
	SHA256     string `json:"sha256"`
}

// HoldoutCommitment proves that a private case set existed before treatment.
// The key and holdout are published only after release scoring.
type HoldoutCommitment struct {
	ID        string           `json:"id"`
	Method    string           `json:"method"`
	Digest    string           `json:"digest"`
	CaseCount int              `json:"case_count"`
	Strata    []HoldoutStratum `json:"strata"`
}

// HoldoutStratum exposes coverage counts without exposing prompts or case IDs.
type HoldoutStratum struct {
	Collection string `json:"collection"`
	Kind       string `json:"kind"`
	TaskType   string `json:"task_type"`
	Count      int    `json:"count"`
}

// PrivateHoldout is stored outside Git until release scoring completes.
type PrivateHoldout struct {
	SchemaVersion int                  `json:"schema_version"`
	ID            string               `json:"id"`
	Cases         []PrivateHoldoutCase `json:"cases"`
}

// PrivateHoldoutCase attaches a secret case to a canonical owner for routing,
// collection metrics, explicit invocation, and risk attribution.
type PrivateHoldoutCase struct {
	Skill    string          `json:"skill"`
	TaskType string          `json:"task_type"`
	Case     corpus.EvalCase `json:"case"`
}

// FreezeOptions contains the values that cannot be inferred from repository
// inputs. Client and toolchain versions are passed explicitly for testability.
type FreezeOptions struct {
	ID                 string
	CreatedOn          string
	Runner             string
	Model              string
	ClientVersion      string
	ToolchainVersion   string
	JudgeRunner        string
	JudgeModel         string
	Modes              []string
	TreatmentSeeds     []int64
	JudgeSeeds         []int64
	TreatmentTimeout   time.Duration
	JudgmentTimeout    time.Duration
	PrivateHoldoutPath string
	HoldoutKeyPath     string
}

// FreezeVerifyOptions controls whether a verifier has access to the still
// private holdout. Public-only verification still checks every disclosed input.
type FreezeVerifyOptions struct {
	ClientVersion      string
	ToolchainVersion   string
	PrivateHoldoutPath string
	HoldoutKeyPath     string
	PublicOnly         bool
}

// CreateBenchmarkFreeze builds a deterministic lock from committed inputs.
func CreateBenchmarkFreeze(collection corpus.Collection, options FreezeOptions) (BenchmarkFreeze, error) {
	if err := validateFreezeOptions(options); err != nil {
		return BenchmarkFreeze{}, err
	}
	if err := ValidateArmFiles(collection); err != nil {
		return BenchmarkFreeze{}, err
	}
	if err := requireCommittedFreezeInputs(collection.RepoRoot); err != nil {
		return BenchmarkFreeze{}, err
	}
	commit := gitHead(collection.RepoRoot)
	if !gitCommitPattern.MatchString(commit) {
		return BenchmarkFreeze{}, errors.New("benchmark freeze requires a committed Git revision")
	}
	inputs, err := freezeInputs(collection.RepoRoot)
	if err != nil {
		return BenchmarkFreeze{}, err
	}
	arms, err := freezeArms(collection)
	if err != nil {
		return BenchmarkFreeze{}, err
	}
	publicCases, err := freezePublicCases(collection)
	if err != nil {
		return BenchmarkFreeze{}, err
	}
	lock := BenchmarkFreeze{
		SchemaVersion:     freezeSchemaVersion,
		ID:                options.ID,
		CreatedOn:         options.CreatedOn,
		CreatedFromCommit: commit,
		Protocol: FreezeProtocol{
			Runner: options.Runner, Model: options.Model, ClientVersion: options.ClientVersion,
			ToolchainVersion: options.ToolchainVersion, JudgeRunner: options.JudgeRunner,
			JudgeModel: options.JudgeModel, Modes: normalizedModes(options.Modes),
			TreatmentSeeds: append([]int64(nil), options.TreatmentSeeds...), JudgeSeeds: append([]int64(nil), options.JudgeSeeds...),
			TreatmentTimeoutMS: options.TreatmentTimeout.Milliseconds(), JudgmentTimeoutMS: options.JudgmentTimeout.Milliseconds(),
			RunnerGraderVersion: runnerGraderVersion, RubricVersion: rubricVersion,
			DeterministicScorer: "skillctl-eval-v2", SemanticScorer: "skillctl-eval-v3-semantic",
		},
		Inputs: inputs, Arms: arms, PublicCases: publicCases,
	}
	if options.PrivateHoldoutPath != "" {
		if err := requirePrivateFreezeFile(collection.RepoRoot, options.PrivateHoldoutPath, "private holdout"); err != nil {
			return BenchmarkFreeze{}, err
		}
		if err := requirePrivateFreezeFile(collection.RepoRoot, options.HoldoutKeyPath, "holdout key"); err != nil {
			return BenchmarkFreeze{}, err
		}
		holdout, merged, err := LoadPrivateHoldout(options.PrivateHoldoutPath, collection)
		if err != nil {
			return BenchmarkFreeze{}, err
		}
		_ = merged
		key, err := readHoldoutKey(options.HoldoutKeyPath, true)
		if err != nil {
			return BenchmarkFreeze{}, err
		}
		commitment, err := commitHoldout(holdout, key, collection)
		if err != nil {
			return BenchmarkFreeze{}, err
		}
		lock.PrivateHoldout = &commitment
	}
	return lock, nil
}

// WriteBenchmarkFreeze writes a stable, reviewable JSON lock.
func WriteBenchmarkFreeze(path string, lock BenchmarkFreeze) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(lock); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

// VerifyBenchmarkFreeze detects drift and optionally opens and merges the
// committed private holdout. The returned digest binds run artifacts to the lock.
func VerifyBenchmarkFreeze(collection corpus.Collection, path string, options FreezeVerifyOptions) (BenchmarkFreeze, corpus.Collection, string, error) {
	var lock BenchmarkFreeze
	if err := decodeStrictFile(path, &lock); err != nil {
		return BenchmarkFreeze{}, corpus.Collection{}, "", err
	}
	if err := validateFreezeLock(lock); err != nil {
		return BenchmarkFreeze{}, corpus.Collection{}, "", err
	}
	if options.ClientVersion != "" && lock.Protocol.ClientVersion != options.ClientVersion {
		return BenchmarkFreeze{}, corpus.Collection{}, "", fmt.Errorf("client version drift: lock has %q, current is %q", lock.Protocol.ClientVersion, options.ClientVersion)
	}
	if options.ToolchainVersion != "" && lock.Protocol.ToolchainVersion != options.ToolchainVersion {
		return BenchmarkFreeze{}, corpus.Collection{}, "", fmt.Errorf("toolchain version drift: lock has %q, current is %q", lock.Protocol.ToolchainVersion, options.ToolchainVersion)
	}
	inputs, err := freezeInputs(collection.RepoRoot)
	if err != nil {
		return BenchmarkFreeze{}, corpus.Collection{}, "", err
	}
	if !reflect.DeepEqual(inputs, lock.Inputs) {
		return BenchmarkFreeze{}, corpus.Collection{}, "", errors.New("benchmark input drift: repository digests do not match the freeze")
	}
	arms, err := freezeArms(collection)
	if err != nil {
		return BenchmarkFreeze{}, corpus.Collection{}, "", err
	}
	if !reflect.DeepEqual(arms, lock.Arms) {
		return BenchmarkFreeze{}, corpus.Collection{}, "", errors.New("benchmark arm drift: installed bytes, maps, or commits do not match the freeze")
	}
	publicCases, err := freezePublicCases(collection)
	if err != nil {
		return BenchmarkFreeze{}, corpus.Collection{}, "", err
	}
	if !reflect.DeepEqual(publicCases, lock.PublicCases) {
		return BenchmarkFreeze{}, corpus.Collection{}, "", errors.New("benchmark case drift: public case inventory does not match the freeze")
	}
	merged := collection
	if lock.PrivateHoldout != nil && !options.PublicOnly {
		if options.PrivateHoldoutPath == "" || options.HoldoutKeyPath == "" {
			return BenchmarkFreeze{}, corpus.Collection{}, "", errors.New("full freeze verification requires -private-holdout and -holdout-key")
		}
		holdout, withHoldout, err := LoadPrivateHoldout(options.PrivateHoldoutPath, collection)
		if err != nil {
			return BenchmarkFreeze{}, corpus.Collection{}, "", err
		}
		key, err := readHoldoutKey(options.HoldoutKeyPath, false)
		if err != nil {
			return BenchmarkFreeze{}, corpus.Collection{}, "", err
		}
		commitment, err := commitHoldout(holdout, key, collection)
		if err != nil {
			return BenchmarkFreeze{}, corpus.Collection{}, "", err
		}
		if !reflect.DeepEqual(commitment, *lock.PrivateHoldout) {
			return BenchmarkFreeze{}, corpus.Collection{}, "", errors.New("private holdout commitment mismatch")
		}
		merged = withHoldout
	} else if lock.PrivateHoldout == nil && (options.PrivateHoldoutPath != "" || options.HoldoutKeyPath != "") {
		return BenchmarkFreeze{}, corpus.Collection{}, "", errors.New("freeze has no private holdout commitment")
	}
	digest, err := BenchmarkFreezeDigest(lock)
	if err != nil {
		return BenchmarkFreeze{}, corpus.Collection{}, "", err
	}
	return lock, merged, digest, nil
}

// BenchmarkFreezeDigest is stored in each treatment cell to prevent a resume
// from silently reusing results from another frozen protocol.
func BenchmarkFreezeDigest(lock BenchmarkFreeze) (string, error) {
	content, err := json.Marshal(lock)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:]), nil
}

// BindMatrixFreeze rejects settings that differ from the precommitted protocol
// and attaches the lock identity to every emitted cell.
func BindMatrixFreeze(lock BenchmarkFreeze, digest string, options *MatrixOptions) error {
	if options == nil {
		return errors.New("matrix options are required")
	}
	if err := validateTreatmentProtocol(lock, options.Runner, options.Model, benchmarkMode(options.ExplicitSkill), options.Split, options.Repetition, options.Seed, options.Timeout); err != nil {
		return err
	}
	if options.Limit != 0 || options.FixturesOnly {
		return errors.New("frozen matrix runs cannot use -limit or -fixtures-only; use -kind or -case for resumable subsets")
	}
	if options.Split == "holdout" && lock.PrivateHoldout == nil {
		return errors.New("frozen matrix has no private holdout")
	}
	if options.Split == "development" && options.Case != "" && !containsFrozenCase(lock.PublicCases, options.Case) {
		return fmt.Errorf("case %q is not in the frozen public inventory", options.Case)
	}
	options.FreezeID = lock.ID
	options.FreezeDigest = digest
	return nil
}

// BindRunFreeze is the single-arm equivalent used for targeted recovery cells.
func BindRunFreeze(lock BenchmarkFreeze, digest string, options *RunOptions) error {
	if options == nil {
		return errors.New("run options are required")
	}
	if err := validateTreatmentProtocol(lock, options.Runner, options.Model, benchmarkMode(options.ExplicitSkill), options.Split, options.Repetition, options.Seed, options.Timeout); err != nil {
		return err
	}
	if options.Limit != 0 || options.FixturesOnly {
		return errors.New("frozen runs cannot use -limit or -fixtures-only; use -kind or -case for resumable subsets")
	}
	if !containsFrozenArm(lock.Arms, options.Arm) {
		return fmt.Errorf("arm %q is not in the benchmark freeze", options.Arm)
	}
	if options.Split == "holdout" && lock.PrivateHoldout == nil {
		return errors.New("frozen run has no private holdout")
	}
	if options.Split == "development" && options.Case != "" && !containsFrozenCase(lock.PublicCases, options.Case) {
		return fmt.Errorf("case %q is not in the frozen public inventory", options.Case)
	}
	options.FreezeID = lock.ID
	options.FreezeDigest = digest
	return nil
}

// ValidateJudgmentFreeze binds semantic evaluation settings to the same lock.
func ValidateJudgmentFreeze(lock BenchmarkFreeze, runner, model string, seed int64, timeout time.Duration) error {
	if runner != lock.Protocol.JudgeRunner || model != lock.Protocol.JudgeModel {
		return fmt.Errorf("judge protocol drift: got %s/%s, want %s/%s", runner, model, lock.Protocol.JudgeRunner, lock.Protocol.JudgeModel)
	}
	if !containsSeed(lock.Protocol.JudgeSeeds, seed) {
		return fmt.Errorf("judge seed %d is not frozen", seed)
	}
	if timeout.Milliseconds() != lock.Protocol.JudgmentTimeoutMS {
		return fmt.Errorf("judge timeout drift: got %dms, want %dms", timeout.Milliseconds(), lock.Protocol.JudgmentTimeoutMS)
	}
	return nil
}

// ValidateResultArtifactFreeze ensures every raw cell belongs to this exact
// lock before scoring. It deliberately accepts incomplete/error cells so an
// interrupted append-only artifact remains auditable.
func ValidateResultArtifactFreeze(path string, lock BenchmarkFreeze, digest string) error {
	results, err := readFreezeResults(path)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return errors.New("frozen result artifact is empty")
	}
	for index, result := range results {
		prefix := fmt.Sprintf("result %d: ", index+1)
		if result.Metadata["freeze_id"] != lock.ID || result.Metadata["freeze_digest"] != digest {
			return errors.New(prefix + "freeze identity mismatch")
		}
		if !containsFrozenArm(lock.Arms, result.Arm) {
			return fmt.Errorf("%sarm %q is not frozen", prefix, result.Arm)
		}
		if result.Runner != lock.Protocol.Runner || result.Model != lock.Protocol.Model || result.ClientVersion != lock.Protocol.ClientVersion {
			return errors.New(prefix + "runner, model, or client version drift")
		}
		if result.Repetition < 0 || result.Repetition >= len(lock.Protocol.TreatmentSeeds) {
			return errors.New(prefix + "repetition is outside the frozen range")
		}
		seed, err := strconv.ParseInt(result.Metadata["treatment_seed"], 10, 64)
		if err != nil || seed != lock.Protocol.TreatmentSeeds[result.Repetition] {
			return errors.New(prefix + "treatment seed mismatch")
		}
		if result.OpaqueArm != opaqueLabel(seed, result.Arm) {
			return errors.New(prefix + "opaque arm label mismatch")
		}
		if !containsString(lock.Protocol.Modes, result.Mode) || !oneOfString(result.Split, "development", "holdout") {
			return errors.New(prefix + "mode or split is not frozen")
		}
		if result.Split == "development" && !containsFrozenCase(lock.PublicCases, result.Skill+"/"+result.CaseID) {
			return errors.New(prefix + "public case is not frozen")
		}
		if result.Split == "holdout" && lock.PrivateHoldout == nil {
			return errors.New(prefix + "holdout cell has no commitment")
		}
		if result.Metadata["grader_version"] != lock.Protocol.RunnerGraderVersion || !gitCommitPattern.MatchString(result.Metadata["fixture_commit"]) {
			return errors.New(prefix + "grader version or fixture commit is invalid")
		}
	}
	return nil
}

// LoadPrivateHoldout strictly loads secret cases and validates them by merging
// them into a copy of the canonical corpus.
func LoadPrivateHoldout(path string, collection corpus.Collection) (PrivateHoldout, corpus.Collection, error) {
	var holdout PrivateHoldout
	if err := decodeStrictFile(path, &holdout); err != nil {
		return PrivateHoldout{}, corpus.Collection{}, err
	}
	if holdout.SchemaVersion != holdoutSchemaVersion || !freezeIDPattern.MatchString(holdout.ID) || len(holdout.Cases) == 0 {
		return PrivateHoldout{}, corpus.Collection{}, errors.New("private holdout requires schema_version 1, a valid id, and at least one case")
	}
	merged := collection
	merged.Skills = append([]corpus.Skill(nil), collection.Skills...)
	indexes := make(map[string]int, len(merged.Skills))
	for index := range merged.Skills {
		merged.Skills[index].Evaluations.Cases = append([]corpus.EvalCase(nil), merged.Skills[index].Evaluations.Cases...)
		indexes[merged.Skills[index].Name] = index
	}
	for _, item := range holdout.Cases {
		index, exists := indexes[item.Skill]
		if !exists {
			return PrivateHoldout{}, corpus.Collection{}, fmt.Errorf("private holdout case %q names unknown skill %q", item.Case.ID, item.Skill)
		}
		if item.Case.Split != "holdout" {
			return PrivateHoldout{}, corpus.Collection{}, fmt.Errorf("private holdout case %s/%s must use split holdout", item.Skill, item.Case.ID)
		}
		if !oneOfString(item.TaskType, "implementation", "diagnosis", "design", "review") {
			return PrivateHoldout{}, corpus.Collection{}, fmt.Errorf("private holdout case %s/%s has invalid task_type %q", item.Skill, item.Case.ID, item.TaskType)
		}
		merged.Skills[index].Evaluations.Cases = append(merged.Skills[index].Evaluations.Cases, item.Case)
	}
	if _, err := corpus.Validate(merged); err != nil {
		return PrivateHoldout{}, corpus.Collection{}, fmt.Errorf("validate private holdout: %w", err)
	}
	return holdout, merged, nil
}

func validateFreezeOptions(options FreezeOptions) error {
	if !freezeIDPattern.MatchString(options.ID) {
		return errors.New("freeze id must contain lowercase letters, digits, dots, or hyphens")
	}
	if _, err := time.Parse("2006-01-02", options.CreatedOn); err != nil {
		return fmt.Errorf("freeze created_on must use YYYY-MM-DD: %w", err)
	}
	if options.Runner != "codex" || options.JudgeRunner != "codex" || options.Model == "" || options.JudgeModel == "" {
		return errors.New("freeze requires Codex treatment and judge runners with explicit models")
	}
	if options.ClientVersion == "" || options.ToolchainVersion == "" {
		return errors.New("freeze requires client and toolchain versions")
	}
	modes := normalizedModes(options.Modes)
	if len(modes) == 0 || len(modes) != len(options.Modes) {
		return errors.New("freeze modes must be unique native and/or explicit values")
	}
	if len(options.TreatmentSeeds) == 0 || len(options.TreatmentSeeds) != len(options.JudgeSeeds) {
		return errors.New("freeze requires one treatment and judge seed per repetition")
	}
	if hasDuplicateSeed(options.TreatmentSeeds) || hasDuplicateSeed(options.JudgeSeeds) || containsSeed(options.TreatmentSeeds, 0) || containsSeed(options.JudgeSeeds, 0) {
		return errors.New("freeze seeds must be nonzero and unique within each phase")
	}
	if options.TreatmentTimeout < time.Millisecond || options.JudgmentTimeout < time.Millisecond {
		return errors.New("freeze timeouts must be at least one millisecond")
	}
	if (options.PrivateHoldoutPath == "") != (options.HoldoutKeyPath == "") {
		return errors.New("private holdout and holdout key must be supplied together")
	}
	return nil
}

func validateFreezeLock(lock BenchmarkFreeze) error {
	if lock.SchemaVersion != freezeSchemaVersion || !freezeIDPattern.MatchString(lock.ID) || !gitCommitPattern.MatchString(lock.CreatedFromCommit) {
		return errors.New("invalid benchmark freeze identity or schema")
	}
	if _, err := time.Parse("2006-01-02", lock.CreatedOn); err != nil {
		return fmt.Errorf("invalid benchmark freeze created_on: %w", err)
	}
	options := FreezeOptions{
		ID: lock.ID, CreatedOn: lock.CreatedOn, Runner: lock.Protocol.Runner, Model: lock.Protocol.Model,
		ClientVersion: lock.Protocol.ClientVersion, ToolchainVersion: lock.Protocol.ToolchainVersion,
		JudgeRunner: lock.Protocol.JudgeRunner, JudgeModel: lock.Protocol.JudgeModel,
		Modes: lock.Protocol.Modes, TreatmentSeeds: lock.Protocol.TreatmentSeeds, JudgeSeeds: lock.Protocol.JudgeSeeds,
		TreatmentTimeout: time.Duration(lock.Protocol.TreatmentTimeoutMS) * time.Millisecond,
		JudgmentTimeout:  time.Duration(lock.Protocol.JudgmentTimeoutMS) * time.Millisecond,
	}
	if err := validateFreezeOptions(options); err != nil {
		return fmt.Errorf("invalid benchmark freeze protocol: %w", err)
	}
	if lock.Protocol.RunnerGraderVersion != runnerGraderVersion || lock.Protocol.RubricVersion != rubricVersion ||
		lock.Protocol.DeterministicScorer != "skillctl-eval-v2" || lock.Protocol.SemanticScorer != "skillctl-eval-v3-semantic" {
		return errors.New("benchmark freeze scorer protocol is unsupported by this skillctl revision")
	}
	if len(lock.Inputs) == 0 || len(lock.Arms) < 2 || len(lock.PublicCases) == 0 {
		return errors.New("benchmark freeze is missing inputs, arms, or public cases")
	}
	if lock.PrivateHoldout != nil {
		if !freezeIDPattern.MatchString(lock.PrivateHoldout.ID) || lock.PrivateHoldout.Method != holdoutCommitmentMethod ||
			len(lock.PrivateHoldout.Digest) != sha256.Size*2 || lock.PrivateHoldout.CaseCount <= 0 || len(lock.PrivateHoldout.Strata) == 0 {
			return errors.New("benchmark freeze has an invalid private holdout commitment")
		}
		if _, err := hex.DecodeString(lock.PrivateHoldout.Digest); err != nil {
			return errors.New("benchmark freeze has a non-hexadecimal private holdout commitment")
		}
	}
	return nil
}

func validateTreatmentProtocol(lock BenchmarkFreeze, runner, model, mode, split string, repetition int, seed int64, timeout time.Duration) error {
	if runner != lock.Protocol.Runner || model != lock.Protocol.Model {
		return fmt.Errorf("treatment protocol drift: got %s/%s, want %s/%s", runner, model, lock.Protocol.Runner, lock.Protocol.Model)
	}
	if !containsString(lock.Protocol.Modes, mode) {
		return fmt.Errorf("mode %q is not frozen", mode)
	}
	if !oneOfString(split, "development", "holdout") {
		return errors.New("frozen runs require -split development or -split holdout")
	}
	if repetition < 0 || repetition >= len(lock.Protocol.TreatmentSeeds) {
		return fmt.Errorf("repetition %d is outside the frozen range", repetition)
	}
	if seed != lock.Protocol.TreatmentSeeds[repetition] {
		return fmt.Errorf("treatment seed drift: got %d, want %d", seed, lock.Protocol.TreatmentSeeds[repetition])
	}
	if timeout.Milliseconds() != lock.Protocol.TreatmentTimeoutMS {
		return fmt.Errorf("treatment timeout drift: got %dms, want %dms", timeout.Milliseconds(), lock.Protocol.TreatmentTimeoutMS)
	}
	return nil
}

func containsFrozenArm(arms []FrozenArm, name string) bool {
	for _, arm := range arms {
		if arm.Name == name {
			return true
		}
	}
	return false
}

func containsFrozenCase(cases []FrozenCase, key string) bool {
	for _, item := range cases {
		if item.Key == key {
			return true
		}
	}
	return false
}

func containsSeed(seeds []int64, seed int64) bool {
	for _, item := range seeds {
		if item == seed {
			return true
		}
	}
	return false
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func freezeInputs(repoRoot string) ([]FrozenInput, error) {
	paths := []string{
		"cmd/skillctl",
		"evaluations/arms",
		"evaluations/fixtures",
		"evaluations/oracles",
		"go.mod",
		"internal",
		"research/corpus-lock.json",
		"skills",
	}
	inputs := make([]FrozenInput, 0, len(paths))
	for _, relative := range paths {
		digest, err := digestPath(filepath.Join(repoRoot, filepath.FromSlash(relative)))
		if err != nil {
			return nil, fmt.Errorf("digest freeze input %s: %w", relative, err)
		}
		inputs = append(inputs, FrozenInput{Path: relative, SHA256: digest})
	}
	return inputs, nil
}

func freezePublicCases(collection corpus.Collection) ([]FrozenCase, error) {
	var result []FrozenCase
	for _, skill := range collection.Skills {
		for _, item := range skill.Evaluations.Cases {
			if item.Split != "development" {
				continue
			}
			content, err := json.Marshal(struct {
				Skill       string          `json:"skill"`
				Collection  string          `json:"collection"`
				RiskDomains []string        `json:"risk_domains"`
				Case        corpus.EvalCase `json:"case"`
			}{skill.Name, skill.Metadata.Collection, skill.Metadata.RiskDomains, item})
			if err != nil {
				return nil, err
			}
			digest := sha256.Sum256(content)
			result = append(result, FrozenCase{
				Key: skill.Name + "/" + item.ID, Collection: skill.Metadata.Collection,
				Kind: item.Kind, Split: item.Split, SHA256: hex.EncodeToString(digest[:]),
			})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
	return result, nil
}

func freezeArms(collection corpus.Collection) ([]FrozenArm, error) {
	manifestPath := filepath.Join(collection.RepoRoot, "evaluations", "arms", "manifest.json")
	var manifest armManifest
	if err := decodeStrictFile(manifestPath, &manifest); err != nil {
		return nil, err
	}
	oursDigest, err := digestPath(filepath.Join(collection.RepoRoot, "skills"))
	if err != nil {
		return nil, err
	}
	result := []FrozenArm{{Name: "baseline"}, {
		Name: "ours", SkillsRoot: "skills", SkillsSHA256: oursDigest,
	}}
	for _, spec := range manifest.Arms {
		actualCommit, relativeSkillsRoot, err := gitCheckoutIdentity(spec.SkillsRoot)
		if err != nil {
			return nil, fmt.Errorf("arm %s: %w", spec.Name, err)
		}
		if actualCommit != spec.RepositoryCommit {
			return nil, fmt.Errorf("arm %s checkout is %s; want %s", spec.Name, actualCommit, spec.RepositoryCommit)
		}
		if dirty, err := gitPathDirty(spec.SkillsRoot); err != nil {
			return nil, fmt.Errorf("arm %s: %w", spec.Name, err)
		} else if dirty {
			return nil, fmt.Errorf("arm %s skills root has tracked or untracked changes", spec.Name)
		}
		skillsDigest, err := digestPath(spec.SkillsRoot)
		if err != nil {
			return nil, fmt.Errorf("arm %s skill digest: %w", spec.Name, err)
		}
		mapPath := filepath.Join(collection.RepoRoot, filepath.FromSlash(spec.SkillMap))
		mapDigest, err := digestPath(mapPath)
		if err != nil {
			return nil, fmt.Errorf("arm %s map digest: %w", spec.Name, err)
		}
		result = append(result, FrozenArm{
			Name: spec.Name, RepositoryCommit: spec.RepositoryCommit, SkillsRoot: relativeSkillsRoot,
			SkillsSHA256: skillsDigest, SkillMap: spec.SkillMap, SkillMapSHA256: mapDigest,
		})
	}
	return result, nil
}

func commitHoldout(holdout PrivateHoldout, key []byte, collection corpus.Collection) (HoldoutCommitment, error) {
	content, err := json.Marshal(holdout)
	if err != nil {
		return HoldoutCommitment{}, err
	}
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write(content); err != nil {
		return HoldoutCommitment{}, err
	}
	collections := make(map[string]string, len(collection.Skills))
	for _, skill := range collection.Skills {
		collections[skill.Name] = skill.Metadata.Collection
	}
	counts := make(map[string]int)
	for _, item := range holdout.Cases {
		key := strings.Join([]string{collections[item.Skill], item.Case.Kind, item.TaskType}, "\x00")
		counts[key]++
	}
	strata := make([]HoldoutStratum, 0, len(counts))
	for key, count := range counts {
		parts := strings.Split(key, "\x00")
		strata = append(strata, HoldoutStratum{Collection: parts[0], Kind: parts[1], TaskType: parts[2], Count: count})
	}
	sort.Slice(strata, func(i, j int) bool {
		left := strata[i].Collection + "\x00" + strata[i].Kind + "\x00" + strata[i].TaskType
		right := strata[j].Collection + "\x00" + strata[j].Kind + "\x00" + strata[j].TaskType
		return left < right
	})
	return HoldoutCommitment{
		ID: holdout.ID, Method: holdoutCommitmentMethod, Digest: hex.EncodeToString(mac.Sum(nil)),
		CaseCount: len(holdout.Cases), Strata: strata,
	}, nil
}

func readHoldoutKey(path string, requireOwnerOnly bool) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("read holdout key metadata: %w", err)
	}
	if !info.Mode().IsRegular() || requireOwnerOnly && info.Mode().Perm()&0o077 != 0 {
		return nil, errors.New("holdout key must be a regular file readable only by its owner before scoring")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	encoded := strings.TrimSpace(string(content))
	if len(encoded) < sha256.Size*2 || len(encoded)%2 != 0 {
		return nil, errors.New("holdout key must contain at least 32 bytes encoded as lowercase hexadecimal")
	}
	key, err := hex.DecodeString(encoded)
	if err != nil || hex.EncodeToString(key) != encoded {
		return nil, errors.New("holdout key must contain lowercase hexadecimal only")
	}
	return key, nil
}

func requirePrivateFreezeFile(repoRoot, path, label string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect %s: %w", label, err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%s must be a regular file readable only by its owner", label)
	}
	absoluteRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return err
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(absoluteRoot, absolutePath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil
	}
	if err := exec.Command("git", "-C", repoRoot, "check-ignore", "--quiet", "--", relative).Run(); err != nil {
		return fmt.Errorf("%s inside the repository must be ignored by Git", label)
	}
	return nil
}

func requireCommittedFreezeInputs(repoRoot string) error {
	paths := []string{"skills", "evaluations/fixtures", "evaluations/oracles", "evaluations/arms", "cmd/skillctl", "internal", "go.mod", "research/corpus-lock.json"}
	arguments := append([]string{"-C", repoRoot, "status", "--porcelain=v1", "--untracked-files=all", "--"}, paths...)
	output, err := exec.Command("git", arguments...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("inspect committed freeze inputs: %w", err)
	}
	if strings.TrimSpace(string(output)) != "" {
		return fmt.Errorf("benchmark freeze inputs must be committed and clean: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func gitCheckoutIdentity(path string) (string, string, error) {
	rootOutput, err := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("resolve Git root: %w", err)
	}
	root := strings.TrimSpace(string(rootOutput))
	commitOutput, err := exec.Command("git", "-C", path, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("resolve Git revision: %w", err)
	}
	commit := strings.TrimSpace(string(commitOutput))
	if !gitCommitPattern.MatchString(commit) {
		return "", "", errors.New("resolved Git revision is not a full commit")
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", "", err
	}
	relative, err := filepath.Rel(root, absolutePath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", errors.New("skills root is outside its Git checkout")
	}
	if relative == "." {
		return commit, ".", nil
	}
	return commit, filepath.ToSlash(relative), nil
}

func gitPathDirty(path string) (bool, error) {
	output, err := exec.Command("git", "-C", path, "status", "--porcelain=v1", "--untracked-files=all", "--", ".").CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("inspect Git status: %w", err)
	}
	return strings.TrimSpace(string(output)) != "", nil
}

func digestPath(root string) (string, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("symbolic links are not valid freeze inputs")
	}
	type record struct {
		path string
		mode string
		data []byte
	}
	var records []record
	if info.Mode().IsRegular() {
		content, err := os.ReadFile(root)
		if err != nil {
			return "", err
		}
		records = append(records, record{path: ".", mode: info.Mode().Perm().String(), data: content})
	} else if info.IsDir() {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if path == root {
				return nil
			}
			entryInfo, err := entry.Info()
			if err != nil {
				return err
			}
			if entryInfo.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("symbolic link %s is not a valid freeze input", path)
			}
			if entry.IsDir() {
				return nil
			}
			if !entryInfo.Mode().IsRegular() {
				return fmt.Errorf("non-regular file %s is not a valid freeze input", path)
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			records = append(records, record{path: filepath.ToSlash(relative), mode: entryInfo.Mode().Perm().String(), data: content})
			return nil
		})
		if err != nil {
			return "", err
		}
	} else {
		return "", errors.New("freeze input must be a regular file or directory")
	}
	sort.Slice(records, func(i, j int) bool { return records[i].path < records[j].path })
	hash := sha256.New()
	for _, item := range records {
		writeHashField(hash, item.path)
		writeHashField(hash, item.mode)
		writeHashBytes(hash, item.data)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeHashField(writer io.Writer, value string) {
	_, _ = io.WriteString(writer, strconv.Itoa(len(value))+":")
	_, _ = io.WriteString(writer, value)
}

func writeHashBytes(writer io.Writer, value []byte) {
	_, _ = io.WriteString(writer, strconv.Itoa(len(value))+":")
	_, _ = writer.Write(value)
}

func normalizedModes(modes []string) []string {
	seen := make(map[string]struct{}, len(modes))
	var result []string
	for _, mode := range modes {
		if !oneOfString(mode, "native", "explicit") {
			continue
		}
		if _, exists := seen[mode]; exists {
			continue
		}
		seen[mode] = struct{}{}
		result = append(result, mode)
	}
	sort.Strings(result)
	return result
}

func hasDuplicateSeed(seeds []int64) bool {
	seen := make(map[int64]struct{}, len(seeds))
	for _, seed := range seeds {
		if _, exists := seen[seed]; exists {
			return true
		}
		seen[seed] = struct{}{}
	}
	return false
}

func oneOfString(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

// readFreezeResults is intentionally strict about line framing but permits the
// result schema to evolve through normal json.Unmarshal compatibility.
func readFreezeResults(path string) ([]Result, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var results []Result
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		var result Result
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, scanner.Err()
}
