// Package evaluation runs isolated agent-skill benchmark cells and scores their
// deterministic criteria before any semantic model judgment.
package evaluation

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
)

// ClientStatus is a read-only runner availability and authentication result.
type ClientStatus struct {
	Client        string `json:"client"`
	Executable    string `json:"executable,omitempty"`
	Available     bool   `json:"available"`
	Authenticated bool   `json:"authenticated"`
	Version       string `json:"version,omitempty"`
	Evidence      string `json:"evidence"`
}

// RunOptions selects one isolated benchmark arm.
type RunOptions struct {
	Runner        string
	Model         string
	Arm           string
	SkillsRoot    string
	Kind          string
	Case          string
	ExplicitSkill bool
	RoutingMap    map[string][]string
	SkillMap      map[string]string
	Limit         int
	Seed          int64
	Timeout       time.Duration
	OutputPath    string
}

// Result is one resumable JSONL benchmark cell.
type Result struct {
	SchemaVersion  int                  `json:"schema_version"`
	RunID          string               `json:"run_id"`
	OpaqueArm      string               `json:"opaque_arm"`
	Arm            string               `json:"arm"`
	Runner         string               `json:"runner"`
	Model          string               `json:"model"`
	ClientVersion  string               `json:"client_version"`
	Skill          string               `json:"skill"`
	CaseID         string               `json:"case_id"`
	Kind           string               `json:"kind"`
	Prompt         string               `json:"prompt"`
	ExpectedRoute  string               `json:"expected_route,omitempty"`
	ExpectedRoutes []string             `json:"expected_routes,omitempty"`
	Response       string               `json:"response"`
	ExitCode       int                  `json:"exit_code"`
	DurationMS     int64                `json:"duration_ms"`
	RunnerEvents   string               `json:"runner_events"`
	Error          string               `json:"error,omitempty"`
	Graders        []corpus.Grader      `json:"graders,omitempty"`
	Invariants     []string             `json:"expected_invariants,omitempty"`
	Forbidden      []string             `json:"forbidden_outcomes,omitempty"`
	GraderRuns     map[string]GraderRun `json:"grader_runs,omitempty"`
	FixtureFiles   map[string]string    `json:"fixture_files,omitempty"`
	Metadata       map[string]string    `json:"metadata"`
}

// GraderRun preserves deterministic fixture evidence from the isolated cell.
type GraderRun struct {
	Passed   bool   `json:"passed"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
}

// Score is a deterministic score for one result.
type Score struct {
	Result
	Passed      bool            `json:"passed"`
	Score       float64         `json:"score"`
	GraderScore map[string]bool `json:"grader_score"`
	Failures    []string        `json:"failures,omitempty"`
}

// Report summarizes deterministic pass rates and a Wilson confidence interval.
type Report struct {
	SchemaVersion int     `json:"schema_version"`
	Arm           string  `json:"arm"`
	Cases         int     `json:"cases"`
	Passed        int     `json:"passed"`
	PassRate      float64 `json:"pass_rate"`
	CILower95     float64 `json:"ci_lower_95"`
	CIUpper95     float64 `json:"ci_upper_95"`
	CriticalFails int     `json:"critical_failures"`
	Input         string  `json:"input"`
}

// Preflight inspects supported local runners without changing authentication.
func Preflight(ctx context.Context) []ClientStatus {
	checks := []struct {
		client  string
		command string
		auth    []string
	}{
		{client: "codex", command: "codex", auth: []string{"login", "status"}},
		{client: "claude-code", command: "claude", auth: []string{"auth", "status"}},
		{client: "cursor", command: "cursor-agent"},
		{client: "opencode", command: "opencode"},
	}
	statuses := make([]ClientStatus, 0, len(checks))
	for _, check := range checks {
		status := ClientStatus{Client: check.client, Evidence: "executable not found"}
		path, err := exec.LookPath(check.command)
		if err != nil {
			statuses = append(statuses, status)
			continue
		}
		status.Available = true
		status.Executable = path
		status.Version = commandOutput(ctx, path, "--version")
		if len(check.auth) == 0 {
			status.Evidence = "runner executable found; authentication cannot be established non-interactively"
		} else {
			authOutput, authErr := commandOutputWithError(ctx, path, check.auth...)
			status.Authenticated = authErr == nil && (strings.Contains(strings.ToLower(authOutput), "logged in") || strings.Contains(authOutput, `"loggedIn": true`))
			status.Evidence = strings.TrimSpace(authOutput)
		}
		statuses = append(statuses, status)
	}
	return statuses
}

// Run executes selected cases in randomized order with a new temp directory and
// fresh ephemeral agent session for every case.
func Run(ctx context.Context, collection corpus.Collection, options RunOptions) (int, error) {
	if options.Runner != "codex" {
		return 0, fmt.Errorf("runner %q is not behaviorally enabled; use eval preflight for structural clients", options.Runner)
	}
	if options.OutputPath == "" {
		return 0, errors.New("output path is required")
	}
	if options.Timeout <= 0 {
		options.Timeout = 5 * time.Minute
	}
	if options.Seed == 0 {
		options.Seed = 1
	}
	cases := flattenCases(collection, options.Kind, options.Case)
	if err := validateArmMaps(options, cases); err != nil {
		return 0, err
	}
	rand.New(rand.NewSource(options.Seed)).Shuffle(len(cases), func(i, j int) { cases[i], cases[j] = cases[j], cases[i] })
	if options.Limit > 0 && options.Limit < len(cases) {
		cases = cases[:options.Limit]
	}
	if len(cases) == 0 {
		return 0, errors.New("no evaluation cases selected")
	}
	if err := os.MkdirAll(filepath.Dir(options.OutputPath), 0o755); err != nil {
		return 0, err
	}
	completed, err := completedCases(options.OutputPath)
	if err != nil {
		return 0, err
	}
	file, err := os.OpenFile(options.OutputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	runID := fmt.Sprintf("%s-%d", time.Now().UTC().Format("20060102T150405Z"), options.Seed)
	opaqueArm := fmt.Sprintf("arm-%08x", rand.New(rand.NewSource(options.Seed)).Uint32())
	clientVersion := commandOutput(ctx, "codex", "--version")
	written := 0
	for _, item := range cases {
		key := options.Arm + "/" + item.skill.Name + "/" + item.eval.ID
		if completed[key] {
			continue
		}
		result := runCase(ctx, collection.RepoRoot, options, runID, opaqueArm, clientVersion, item)
		if err := encoder.Encode(result); err != nil {
			return written, err
		}
		if err := file.Sync(); err != nil {
			return written, err
		}
		written++
	}
	return written, nil
}

func validateArmMaps(options RunOptions, cases []caseItem) error {
	if options.Arm == "ours" || options.Arm == "baseline" {
		return nil
	}
	for _, item := range cases {
		if item.eval.Kind == "routing" && len(options.RoutingMap[item.skill.Name+"/"+item.eval.ID]) == 0 {
			for _, canonical := range canonicalExpectedRoutes(item) {
				if canonical != "NONE" && options.SkillMap[canonical] == "" {
					return fmt.Errorf("competitor routing arm %q is missing a case override or -skill-map entry %s", options.Arm, canonical)
				}
			}
		}
		if options.ExplicitSkill && options.SkillMap[item.skill.Name] == "" {
			return fmt.Errorf("competitor explicit-skill arm %q is missing -skill-map entry %s", options.Arm, item.skill.Name)
		}
	}
	return nil
}

type caseItem struct {
	skill corpus.Skill
	eval  corpus.EvalCase
}

func flattenCases(collection corpus.Collection, kind, selectedCase string) []caseItem {
	var result []caseItem
	for _, skill := range collection.Skills {
		for _, eval := range skill.Evaluations.Cases {
			caseKey := skill.Name + "/" + eval.ID
			if (selectedCase == "" || selectedCase == caseKey) && (kind == "" || kind == "all" || eval.Kind == kind) {
				result = append(result, caseItem{skill: skill, eval: eval})
			}
		}
	}
	return result
}

func runCase(parent context.Context, repoRoot string, options RunOptions, runID, opaqueArm, clientVersion string, item caseItem) Result {
	started := time.Now()
	result := Result{
		SchemaVersion: 1, RunID: runID, OpaqueArm: opaqueArm, Arm: options.Arm,
		Runner: options.Runner, Model: options.Model, ClientVersion: clientVersion,
		Skill: item.skill.Name, CaseID: item.eval.ID, Kind: item.eval.Kind, Prompt: item.eval.Prompt,
		Graders: item.eval.Graders, Invariants: item.eval.ExpectedInvariants, Forbidden: item.eval.ForbiddenOutcomes,
		Metadata: map[string]string{"fixture_commit": gitHead(repoRoot), "grader_version": "skillctl-eval-v1"},
	}
	if item.eval.Kind == "routing" {
		result.ExpectedRoutes = expectedRoutes(options, item)
	}
	worktree, err := os.MkdirTemp("", "golangskills-eval-*")
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer os.RemoveAll(worktree)
	if item.eval.Fixture != "" {
		fixtureRoot := filepath.Join(repoRoot, filepath.FromSlash(item.eval.Fixture))
		if err := copyTree(fixtureRoot, worktree); err != nil {
			result.Error = "copy fixture: " + err.Error()
			return result
		}
	}
	selectedSkill := item.skill.Name
	if mapped := options.SkillMap[item.skill.Name]; mapped != "" {
		selectedSkill = mapped
	}
	if options.Arm != "baseline" && selectedSkill != "NONE" {
		source := options.SkillsRoot
		if source == "" {
			source = filepath.Join(repoRoot, "skills")
		}
		destination := filepath.Join(worktree, ".agents", "skills")
		if err := copySkills(source, destination, options.ExplicitSkill, selectedSkill); err != nil {
			result.Error = err.Error()
			return result
		}
	}
	prompt := item.eval.Prompt
	if item.eval.Kind == "routing" {
		prompt = "Choose whether one installed skill should handle the following request. Reply with exactly its skill ID, or NONE if no installed skill should activate. Do not solve the request.\n\n" + prompt
	} else if item.eval.Fixture != "" {
		prompt = "Modify the fixture in the current directory to solve the task. Keep the change scoped and leave the repository ready for its deterministic grader.\n\n" + prompt
		if options.ExplicitSkill && selectedSkill != "NONE" {
			prompt = "Use $" + selectedSkill + ".\n\n" + prompt
		}
	} else if options.ExplicitSkill && selectedSkill != "NONE" {
		prompt = "Use $" + selectedSkill + " for this task.\n\n" + prompt
	}
	caseContext, cancel := context.WithTimeout(parent, options.Timeout)
	defer cancel()
	lastMessage := filepath.Join(worktree, "last-message.txt")
	sandbox := "read-only"
	if item.eval.Fixture != "" {
		sandbox = "workspace-write"
	}
	arguments := []string{"exec", "--json", "--ephemeral", "--ignore-user-config", "--ignore-rules", "--skip-git-repo-check", "--sandbox", sandbox, "--cd", worktree, "--output-last-message", lastMessage}
	if options.Model != "" {
		arguments = append(arguments, "--model", options.Model)
	}
	arguments = append(arguments, prompt)
	command := exec.CommandContext(caseContext, "codex", arguments...)
	events, runErr := command.CombinedOutput()
	result.RunnerEvents = string(events)
	result.ExitCode = 0
	if runErr != nil {
		result.Error = runErr.Error()
		result.ExitCode = exitCode(runErr)
	}
	response, readErr := os.ReadFile(lastMessage)
	if readErr == nil {
		result.Response = strings.TrimSpace(string(response))
	} else if runErr == nil {
		result.Error = readErr.Error()
	}
	if item.eval.Fixture != "" {
		result.GraderRuns = runFixtureGraders(parent, worktree, item.eval.Graders)
		result.FixtureFiles = snapshotFixture(worktree)
	}
	result.DurationMS = time.Since(started).Milliseconds()
	return result
}

func expectedRoutes(options RunOptions, item caseItem) []string {
	if options.Arm == "baseline" {
		return []string{"NONE"}
	}
	if routes := options.RoutingMap[item.skill.Name+"/"+item.eval.ID]; len(routes) > 0 {
		return append([]string(nil), routes...)
	}
	canonical := canonicalExpectedRoutes(item)
	if options.Arm == "ours" {
		return canonical
	}
	mapped := make([]string, 0, len(canonical))
	for _, route := range canonical {
		if armRoute := options.SkillMap[route]; armRoute != "" {
			mapped = append(mapped, armRoute)
		}
	}
	if len(mapped) == 0 {
		return []string{"NONE"}
	}
	return mapped
}

func canonicalExpectedRoutes(item caseItem) []string {
	if item.eval.ShouldActivate != nil && *item.eval.ShouldActivate {
		return []string{item.skill.Name}
	}
	if len(item.eval.ConfusesWith) > 0 {
		return append([]string(nil), item.eval.ConfusesWith...)
	}
	return []string{"NONE"}
}

// ScoreFile applies only deterministic graders and writes JSONL scores.
func ScoreFile(inputPath, outputPath string) (int, error) {
	input, err := os.Open(inputPath)
	if err != nil {
		return 0, err
	}
	defer input.Close()
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return 0, err
	}
	output, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer output.Close()
	encoder := json.NewEncoder(output)
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	count := 0
	for scanner.Scan() {
		var result Result
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return count, err
		}
		score := scoreResult(result)
		if err := encoder.Encode(score); err != nil {
			return count, err
		}
		count++
	}
	return count, scanner.Err()
}

func scoreResult(result Result) Score {
	score := Score{Result: result, GraderScore: make(map[string]bool)}
	if result.Error != "" || result.ExitCode != 0 {
		score.Failures = append(score.Failures, "runner failed")
		return score
	}
	if result.Kind == "routing" {
		expected := result.ExpectedRoutes
		if len(expected) == 0 && result.ExpectedRoute != "" {
			expected = []string{result.ExpectedRoute}
		}
		response := strings.TrimSpace(result.Response)
		for _, route := range expected {
			if strings.EqualFold(response, route) {
				score.Passed = true
				break
			}
		}
		if score.Passed {
			score.Score = 1
		} else {
			score.Failures = append(score.Failures, "route mismatch")
		}
		return score
	}
	totalWeight, passedWeight := 0.0, 0.0
	lower := strings.ToLower(result.Response)
	for _, grader := range result.Graders {
		passed := false
		switch grader.Kind {
		case "go-test":
			passed = result.GraderRuns[grader.ID].Passed
		default:
			passed = true
			for _, required := range grader.Required {
				if !strings.Contains(lower, strings.ToLower(required)) {
					passed = false
				}
			}
			for _, forbidden := range grader.Forbidden {
				if strings.Contains(lower, strings.ToLower(forbidden)) {
					passed = false
				}
			}
		}
		score.GraderScore[grader.ID] = passed
		totalWeight += grader.Weight
		if passed {
			passedWeight += grader.Weight
		} else {
			score.Failures = append(score.Failures, "grader "+grader.ID+" failed")
		}
	}
	if totalWeight > 0 {
		score.Score = passedWeight / totalWeight
	}
	score.Passed = score.Score == 1
	return score
}

// ReportFile summarizes a scored JSONL artifact.
func ReportFile(inputPath string) (Report, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return Report{}, err
	}
	defer file.Close()
	report := Report{SchemaVersion: 1, Input: inputPath}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		var score Score
		if err := json.Unmarshal(scanner.Bytes(), &score); err != nil {
			return Report{}, err
		}
		if report.Arm == "" {
			report.Arm = score.Arm
		}
		report.Cases++
		if score.Passed {
			report.Passed++
		}
		if strings.HasPrefix(score.Skill, "go-fin") || strings.Contains(score.Skill, "money") || strings.Contains(score.Skill, "payment") || strings.Contains(score.Skill, "ledger") {
			if !score.Passed {
				report.CriticalFails++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return Report{}, err
	}
	if report.Cases > 0 {
		report.PassRate = float64(report.Passed) / float64(report.Cases)
		report.CILower95, report.CIUpper95 = wilson(report.Passed, report.Cases)
	}
	return report, nil
}

func wilson(successes, trials int) (float64, float64) {
	if trials == 0 {
		return 0, 0
	}
	z := 1.959963984540054
	n := float64(trials)
	p := float64(successes) / n
	denominator := 1 + z*z/n
	center := (p + z*z/(2*n)) / denominator
	margin := z * math.Sqrt((p*(1-p)+z*z/(4*n))/n) / denominator
	return math.Max(0, center-margin), math.Min(1, center+margin)
}

func completedCases(path string) (map[string]bool, error) {
	completed := make(map[string]bool)
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return completed, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		var result Result
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return nil, err
		}
		completed[result.Arm+"/"+result.Skill+"/"+result.CaseID] = true
	}
	return completed, scanner.Err()
}

func copySkills(source, destination string, explicit bool, skillName string) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	copied := false
	for _, entry := range entries {
		if !entry.IsDir() || explicit && entry.Name() != skillName {
			continue
		}
		if err := copyTree(filepath.Join(source, entry.Name()), filepath.Join(destination, entry.Name())); err != nil {
			return err
		}
		copied = true
	}
	if explicit && !copied {
		return fmt.Errorf("mapped skill %q not found under %s", skillName, source)
	}
	return nil
}

func runFixtureGraders(parent context.Context, worktree string, graders []corpus.Grader) map[string]GraderRun {
	runs := make(map[string]GraderRun)
	for _, grader := range graders {
		if grader.Kind != "go-test" {
			continue
		}
		args := []string{"test", "-mod=readonly", "./..."}
		if grader.Target != "" {
			args = append([]string{"test", "-mod=readonly"}, strings.Fields(grader.Target)...)
		}
		ctx, cancel := context.WithTimeout(parent, 2*time.Minute)
		command := exec.CommandContext(ctx, "go", args...)
		command.Dir = worktree
		command.Env = append(os.Environ(), "GOPROXY=off", "GOSUMDB=off", "GOTOOLCHAIN=local")
		output, err := command.CombinedOutput()
		cancel()
		run := GraderRun{Passed: err == nil, ExitCode: exitCode(err), Output: string(output)}
		if err == nil {
			run.ExitCode = 0
		}
		runs[grader.ID] = run
	}
	return runs
}

func snapshotFixture(root string) map[string]string {
	files := make(map[string]string)
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		if entry.IsDir() && (relative == ".agents" || relative == ".git") {
			return filepath.SkipDir
		}
		if entry.IsDir() || relative == "last-message.txt" {
			return nil
		}
		if filepath.Ext(relative) != ".go" && relative != "go.mod" && relative != "go.sum" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err == nil && len(content) <= 256*1024 {
			files[filepath.ToSlash(relative)] = string(content)
		}
		return nil
	})
	return files
}

func copyTree(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refuse skill symlink %s", path)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, 0o644)
	})
}

func commandOutput(ctx context.Context, command string, arguments ...string) string {
	output, _ := commandOutputWithError(ctx, command, arguments...)
	return strings.TrimSpace(output)
}

func commandOutputWithError(ctx context.Context, command string, arguments ...string) (string, error) {
	output, err := exec.CommandContext(ctx, command, arguments...).CombinedOutput()
	return string(output), err
}

func gitHead(root string) string {
	return commandOutput(context.Background(), "git", "-C", root, "rev-parse", "HEAD")
}

func exitCode(err error) int {
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}
	return -1
}

// WriteReport writes an indented report.
func WriteReport(output io.Writer, report Report) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// SortedStatuses keeps preflight output deterministic.
func SortedStatuses(statuses []ClientStatus) []ClientStatus {
	sort.Slice(statuses, func(i, j int) bool { return statuses[i].Client < statuses[j].Client })
	return statuses
}
