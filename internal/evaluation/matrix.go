package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
)

// MatrixOptions selects the same canonical cases across all arms and then
// randomizes the complete arm/case cell order.
type MatrixOptions struct {
	Runner        string
	Model         string
	Kind          string
	Split         string
	Case          string
	FixturesOnly  bool
	ExplicitSkill bool
	Repetition    int
	Limit         int
	Seed          int64
	Timeout       time.Duration
	MaxCells      int
	MaxFailures   int
	StopOnTimeout bool
	OutputPath    string
	FreezeID      string
	FreezeDigest  string
}

type matrixCell struct {
	options RunOptions
	item    caseItem
}

// FrozenMatrixArms returns baseline, this project, and every committed
// competitor arm with its immutable skill map.
func FrozenMatrixArms(collection corpus.Collection) ([]RunOptions, error) {
	if err := ValidateArmFiles(collection); err != nil {
		return nil, err
	}
	manifest, _, err := loadRuntimeArmInputs(collection.RepoRoot)
	if err != nil {
		return nil, err
	}
	arms := []RunOptions{
		{Arm: "baseline"},
		{Arm: "ours", SkillsRoot: filepath.Join(collection.RepoRoot, "skills")},
	}
	for _, spec := range manifest.Arms {
		var skillMap map[string]string
		if err := decodeStrictFile(filepath.Join(collection.RepoRoot, filepath.FromSlash(spec.SkillMap)), &skillMap); err != nil {
			return nil, fmt.Errorf("load arm %s skill map: %w", spec.Name, err)
		}
		arms = append(arms, RunOptions{Arm: spec.Name, SkillsRoot: spec.SkillsRoot, SkillMap: skillMap})
	}
	return arms, nil
}

// RunMatrix executes globally randomized cells. Each cell still receives a new
// temporary workspace and ephemeral client session through runCase.
func RunMatrix(ctx context.Context, collection corpus.Collection, arms []RunOptions, options MatrixOptions) (int, error) {
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
	if options.Repetition < 0 {
		return 0, errors.New("repetition must be non-negative")
	}
	budget, err := newAttemptBudget(options.MaxCells, options.MaxFailures, options.StopOnTimeout)
	if err != nil {
		return 0, err
	}
	cells, err := planMatrixCells(collection, arms, options)
	if err != nil {
		return 0, err
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
	runID := fmt.Sprintf("%s-matrix-%d-r%d", time.Now().UTC().Format("20060102T150405Z"), options.Seed, options.Repetition)
	clientVersion := commandOutput(ctx, "codex", "--version")
	for _, cell := range cells {
		key := benchmarkKeyWithFreeze(cell.options.Arm, cell.item.skill.Name, cell.item.eval.ID, benchmarkMode(options.ExplicitSkill), options.Repetition, options.FreezeDigest)
		if completed[key] {
			continue
		}
		result := runCase(ctx, collection.RepoRoot, cell.options, runID, opaqueLabel(options.Seed, cell.options.Arm), clientVersion, cell.item)
		if err := encoder.Encode(result); err != nil {
			return budget.written, err
		}
		if err := file.Sync(); err != nil {
			return budget.written, err
		}
		failed := result.Error != "" || result.ExitCode != 0
		switch budget.record(failed, result.Metadata["timed_out"] == "true") {
		case budgetStopTimeout:
			return budget.written, fmt.Errorf("matrix stopped after runner timeout at %s/%s", result.Arm, result.CaseID)
		case budgetStopFailures:
			return budget.written, fmt.Errorf("matrix stopped after %d runner failure(s)", budget.failures)
		case budgetStopAttempts:
			break
		default:
			continue
		}
		break
	}
	return budget.written, nil
}

func planMatrixCells(collection corpus.Collection, arms []RunOptions, options MatrixOptions) ([]matrixCell, error) {
	if len(arms) < 2 {
		return nil, errors.New("matrix requires at least two arms")
	}
	cases := flattenCases(collection, options.Kind, options.Split, options.Case, options.FixturesOnly)
	if len(cases) == 0 {
		return nil, errors.New("no evaluation cases selected")
	}
	random := rand.New(rand.NewSource(options.Seed))
	random.Shuffle(len(cases), func(i, j int) { cases[i], cases[j] = cases[j], cases[i] })
	if options.Limit > 0 && options.Limit < len(cases) {
		cases = cases[:options.Limit]
	}
	var cells []matrixCell
	seenArms := make(map[string]struct{}, len(arms))
	for _, arm := range arms {
		if arm.Arm == "" {
			return nil, errors.New("matrix arm name is required")
		}
		if _, duplicate := seenArms[arm.Arm]; duplicate {
			return nil, fmt.Errorf("duplicate matrix arm %q", arm.Arm)
		}
		seenArms[arm.Arm] = struct{}{}
		arm.Runner = options.Runner
		arm.Model = options.Model
		arm.Kind = options.Kind
		arm.Split = options.Split
		arm.Case = options.Case
		arm.FixturesOnly = options.FixturesOnly
		arm.ExplicitSkill = options.ExplicitSkill
		arm.Repetition = options.Repetition
		arm.Seed = options.Seed
		arm.Timeout = options.Timeout
		arm.FreezeID = options.FreezeID
		arm.FreezeDigest = options.FreezeDigest
		if err := validateArmMaps(arm, cases); err != nil {
			return nil, err
		}
		for _, item := range cases {
			cells = append(cells, matrixCell{options: arm, item: item})
		}
	}
	random.Shuffle(len(cells), func(i, j int) { cells[i], cells[j] = cells[j], cells[i] })
	return cells, nil
}
