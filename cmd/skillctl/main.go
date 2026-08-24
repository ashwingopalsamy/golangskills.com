// Command skillctl validates and generates the Engineering Skills for Go corpus.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ashwingopalsamy/golangskills.com/internal/artifact"
	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
	"github.com/ashwingopalsamy/golangskills.com/internal/evaluation"
	"github.com/ashwingopalsamy/golangskills.com/internal/research"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(arguments []string, output io.Writer) error {
	if len(arguments) == 0 {
		return errors.New("usage: skillctl <audit refs|check|generate|stats> [options]")
	}
	if arguments[0] == "audit" {
		return runAudit(arguments[1:], output)
	}
	if arguments[0] == "eval" {
		return runEval(arguments[1:], output)
	}
	if arguments[0] == "release" {
		return runRelease(arguments[1:], output)
	}
	if arguments[0] == "package" {
		return runPackage(arguments[1:], output)
	}
	command := arguments[0]
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", "", "repository root or a directory below it")
	if err := flags.Parse(arguments[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}

	collection, err := corpus.Load(*root)
	if err != nil {
		return err
	}
	metrics, err := corpus.Validate(collection)
	if err != nil {
		return err
	}

	switch command {
	case "check":
		if err := evaluation.ValidateArmFiles(collection); err != nil {
			return err
		}
		outputs, err := corpus.Render(collection)
		if err != nil {
			return err
		}
		if err := corpus.CheckGenerated(collection, outputs); err != nil {
			return err
		}
		fmt.Fprintln(output, "corpus is valid and generated files are current")
		writeMetrics(output, metrics)
		return nil
	case "generate":
		outputs, err := corpus.Render(collection)
		if err != nil {
			return err
		}
		if err := corpus.WriteGenerated(collection, outputs); err != nil {
			return err
		}
		refreshed, err := corpus.Load(collection.RepoRoot)
		if err != nil {
			return err
		}
		if _, err := corpus.Validate(refreshed); err != nil {
			return err
		}
		if err := corpus.CheckGenerated(refreshed, outputs); err != nil {
			return err
		}
		fmt.Fprintf(output, "generated %d files\n", len(outputs))
		writeMetrics(output, metrics)
		return nil
	case "stats":
		writeMetrics(output, metrics)
		return nil
	default:
		return fmt.Errorf("unknown command %q; expected check, generate, or stats", command)
	}
}

func runPackage(arguments []string, output io.Writer) error {
	flags := flag.NewFlagSet("package", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", "", "repository root")
	version := flags.String("version", "0.2.0", "archive version")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	repositoryRoot, err := corpus.FindRepoRoot(*root)
	if err != nil {
		return err
	}
	paths, err := artifact.Package(repositoryRoot, *version)
	if err != nil {
		return err
	}
	for _, path := range paths {
		fmt.Fprintln(output, path)
	}
	return nil
}

func runEval(arguments []string, output io.Writer) error {
	if len(arguments) == 0 {
		return errors.New("usage: skillctl eval <preflight|freeze|verify-freeze|run|matrix|score|report> [options]")
	}
	switch arguments[0] {
	case "preflight":
		if len(arguments) != 1 {
			return errors.New("usage: skillctl eval preflight")
		}
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")
		return encoder.Encode(evaluation.SortedStatuses(evaluation.Preflight(context.Background())))
	case "freeze":
		flags := flag.NewFlagSet("eval freeze", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		root := flags.String("root", "", "repository root")
		id := flags.String("id", "", "release-candidate freeze id")
		model := flags.String("model", "", "treatment model")
		judgeModel := flags.String("judge-model", "", "semantic evaluator model")
		modes := flags.String("modes", "native,explicit", "comma-separated native and explicit modes")
		repetitions := flags.Int("repetitions", 3, "number of precommitted repetitions")
		seedBase := flags.Int64("seed-base", 2026082401, "first treatment seed")
		judgeSeedBase := flags.Int64("judge-seed-base", 2026083401, "first semantic evaluator seed")
		timeout := flags.Duration("timeout", 5*time.Minute, "per-treatment timeout")
		judgeTimeout := flags.Duration("judge-timeout", 5*time.Minute, "per-judgment timeout")
		privateHoldout := flags.String("private-holdout", "", "untracked private holdout JSON")
		holdoutKey := flags.String("holdout-key", "", "owner-only hexadecimal HMAC key file")
		createdOn := flags.String("created-on", time.Now().UTC().Format("2006-01-02"), "freeze date")
		outputPath := flags.String("output", "", "public freeze lock path")
		if err := flags.Parse(arguments[1:]); err != nil {
			return err
		}
		if *id == "" || *model == "" || *judgeModel == "" || *repetitions <= 0 {
			return errors.New("eval freeze requires -id, -model, -judge-model, and positive -repetitions")
		}
		collection, err := corpus.Load(*root)
		if err != nil {
			return err
		}
		if _, err := corpus.Validate(collection); err != nil {
			return err
		}
		clientVersion, err := authenticatedCodexVersion()
		if err != nil {
			return err
		}
		treatmentSeeds := sequentialSeeds(*seedBase, *repetitions)
		judgeSeeds := sequentialSeeds(*judgeSeedBase, *repetitions)
		lock, err := evaluation.CreateBenchmarkFreeze(collection, evaluation.FreezeOptions{
			ID: *id, CreatedOn: *createdOn, Runner: "codex", Model: *model,
			ClientVersion: clientVersion, ToolchainVersion: runtime.Version(),
			JudgeRunner: "codex", JudgeModel: *judgeModel, Modes: splitList(*modes),
			TreatmentSeeds: treatmentSeeds, JudgeSeeds: judgeSeeds,
			TreatmentTimeout: *timeout, JudgmentTimeout: *judgeTimeout,
			PrivateHoldoutPath: repositoryPath(collection.RepoRoot, *privateHoldout),
			HoldoutKeyPath:     repositoryPath(collection.RepoRoot, *holdoutKey),
		})
		if err != nil {
			return err
		}
		destination := *outputPath
		if destination == "" {
			destination = filepath.Join("evaluations", "releases", *id+".lock.json")
		}
		destination = repositoryPath(collection.RepoRoot, destination)
		if err := evaluation.WriteBenchmarkFreeze(destination, lock); err != nil {
			return err
		}
		digest, err := evaluation.BenchmarkFreezeDigest(lock)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "froze %d public cases and %d arms as %s (%s)\n", len(lock.PublicCases), len(lock.Arms), lock.ID, digest)
		if lock.PrivateHoldout != nil {
			fmt.Fprintf(output, "committed %d private holdout cases without publishing prompts or key\n", lock.PrivateHoldout.CaseCount)
		}
		fmt.Fprintln(output, destination)
		return nil
	case "verify-freeze":
		flags := flag.NewFlagSet("eval verify-freeze", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		root := flags.String("root", "", "repository root")
		lockPath := flags.String("freeze", "", "public freeze lock path")
		privateHoldout := flags.String("private-holdout", "", "private holdout JSON")
		holdoutKey := flags.String("holdout-key", "", "holdout HMAC key file")
		publicOnly := flags.Bool("public-only", false, "verify disclosed inputs without opening a committed private holdout")
		checkEnvironment := flags.Bool("check-environment", false, "also require current Codex and Go versions")
		if err := flags.Parse(arguments[1:]); err != nil {
			return err
		}
		if *lockPath == "" {
			return errors.New("eval verify-freeze requires -freeze")
		}
		collection, err := corpus.Load(*root)
		if err != nil {
			return err
		}
		if _, err := corpus.Validate(collection); err != nil {
			return err
		}
		verify := evaluation.FreezeVerifyOptions{
			PrivateHoldoutPath: repositoryPath(collection.RepoRoot, *privateHoldout),
			HoldoutKeyPath:     repositoryPath(collection.RepoRoot, *holdoutKey), PublicOnly: *publicOnly,
		}
		if *checkEnvironment {
			verify.ClientVersion, err = authenticatedCodexVersion()
			if err != nil {
				return err
			}
			verify.ToolchainVersion = runtime.Version()
		}
		lock, _, digest, err := evaluation.VerifyBenchmarkFreeze(collection, repositoryPath(collection.RepoRoot, *lockPath), verify)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "verified benchmark freeze %s (%s)\n", lock.ID, digest)
		return nil
	case "run":
		flags := flag.NewFlagSet("eval run", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		root := flags.String("root", "", "repository root")
		runner := flags.String("runner", "codex", "agent runner")
		model := flags.String("model", "", "runner model")
		arm := flags.String("arm", "ours", "benchmark arm")
		skillsRoot := flags.String("skills-root", "", "skills directory for the arm")
		routingMapPath := flags.String("routing-map", "", "locked JSON map from canonical skill/case to accepted arm routes")
		skillMapPath := flags.String("skill-map", "", "locked JSON map from canonical skill IDs to arm skill IDs")
		kind := flags.String("kind", "all", "routing, quality, or all")
		split := flags.String("split", "all", "development, holdout, or all")
		selectedCase := flags.String("case", "", "one canonical skill/case key to execute")
		fixturesOnly := flags.Bool("fixtures-only", false, "select only quality cases with executable fixtures")
		explicit := flags.Bool("explicit", false, "install and invoke only the expected canonical skill")
		repetition := flags.Int("repetition", 0, "non-negative repeat index recorded in cell identity")
		limit := flags.Int("limit", 0, "maximum cases")
		seed := flags.Int64("seed", 1, "randomization seed")
		timeout := flags.Duration("timeout", 5*time.Minute, "per-case timeout")
		outputPath := flags.String("output", "evaluations/runs/manual.jsonl", "resumable JSONL artifact")
		freezePath := flags.String("freeze", "", "release-candidate freeze lock")
		privateHoldout := flags.String("private-holdout", "", "private holdout JSON for a frozen holdout run")
		holdoutKey := flags.String("holdout-key", "", "holdout HMAC key file")
		if err := flags.Parse(arguments[1:]); err != nil {
			return err
		}
		collection, err := corpus.Load(*root)
		if err != nil {
			return err
		}
		if _, err := corpus.Validate(collection); err != nil {
			return err
		}
		var lock evaluation.BenchmarkFreeze
		var freezeDigest string
		if *freezePath != "" {
			clientVersion, versionErr := authenticatedCodexVersion()
			if versionErr != nil {
				return versionErr
			}
			lock, collection, freezeDigest, err = evaluation.VerifyBenchmarkFreeze(collection, repositoryPath(collection.RepoRoot, *freezePath), evaluation.FreezeVerifyOptions{
				ClientVersion: clientVersion, ToolchainVersion: runtime.Version(),
				PrivateHoldoutPath: repositoryPath(collection.RepoRoot, *privateHoldout), HoldoutKeyPath: repositoryPath(collection.RepoRoot, *holdoutKey),
				PublicOnly: *split != "holdout",
			})
			if err != nil {
				return err
			}
		} else if *privateHoldout != "" || *holdoutKey != "" {
			return errors.New("private holdout execution requires -freeze")
		}
		routingMap, err := loadRoutingMap(*routingMapPath)
		if err != nil {
			return err
		}
		skillMap, err := loadSkillMap(*skillMapPath)
		if err != nil {
			return err
		}
		runOptions := evaluation.RunOptions{
			Runner: *runner, Model: *model, Arm: *arm, SkillsRoot: *skillsRoot, Kind: *kind, Case: *selectedCase,
			Split: *split, FixturesOnly: *fixturesOnly, ExplicitSkill: *explicit, Repetition: *repetition,
			RoutingMap: routingMap, SkillMap: skillMap, Limit: *limit, Seed: *seed, Timeout: *timeout,
			OutputPath: filepath.Join(collection.RepoRoot, filepath.FromSlash(*outputPath)),
		}
		if *freezePath != "" {
			if *routingMapPath != "" || *skillMapPath != "" || *skillsRoot != "" {
				return errors.New("frozen single-arm runs derive skills and maps from the arm lock; custom arm flags are not allowed")
			}
			frozenArms, armErr := evaluation.FrozenMatrixArms(collection)
			if armErr != nil {
				return armErr
			}
			found := false
			for _, frozenArm := range frozenArms {
				if frozenArm.Arm == *arm {
					runOptions.SkillsRoot, runOptions.SkillMap = frozenArm.SkillsRoot, frozenArm.SkillMap
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("arm %q is not frozen", *arm)
			}
			if err := evaluation.BindRunFreeze(lock, freezeDigest, &runOptions); err != nil {
				return err
			}
		}
		written, err := evaluation.Run(context.Background(), collection, runOptions)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "wrote %d isolated evaluation cells\n", written)
		return nil
	case "matrix":
		flags := flag.NewFlagSet("eval matrix", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		root := flags.String("root", "", "repository root")
		runner := flags.String("runner", "codex", "agent runner")
		model := flags.String("model", "", "runner model")
		kind := flags.String("kind", "all", "routing, quality, or all")
		split := flags.String("split", "all", "development, holdout, or all")
		selectedCase := flags.String("case", "", "one canonical skill/case key to execute")
		fixturesOnly := flags.Bool("fixtures-only", false, "select only quality cases with executable fixtures")
		explicit := flags.Bool("explicit", false, "install and invoke only the mapped skill")
		repetition := flags.Int("repetition", 0, "non-negative repeat index recorded in cell identity")
		limit := flags.Int("limit", 0, "maximum canonical cases selected for every arm")
		maxCells := flags.Int("max-cells", 0, "stop after this many newly written cells; zero means unlimited")
		maxFailures := flags.Int("max-failures", 1, "stop after this many runner failures; zero means unlimited")
		stopOnTimeout := flags.Bool("stop-on-timeout", true, "stop immediately when a cell reaches its timeout")
		seed := flags.Int64("seed", 1, "case and global cell randomization seed")
		timeout := flags.Duration("timeout", 5*time.Minute, "per-cell timeout")
		outputPath := flags.String("output", "evaluations/runs/matrix.jsonl", "resumable mixed-arm JSONL artifact")
		freezePath := flags.String("freeze", "", "release-candidate freeze lock")
		privateHoldout := flags.String("private-holdout", "", "private holdout JSON for a frozen holdout run")
		holdoutKey := flags.String("holdout-key", "", "holdout HMAC key file")
		if err := flags.Parse(arguments[1:]); err != nil {
			return err
		}
		collection, err := corpus.Load(*root)
		if err != nil {
			return err
		}
		if _, err := corpus.Validate(collection); err != nil {
			return err
		}
		var lock evaluation.BenchmarkFreeze
		var freezeDigest string
		if *freezePath != "" {
			clientVersion, versionErr := authenticatedCodexVersion()
			if versionErr != nil {
				return versionErr
			}
			lock, collection, freezeDigest, err = evaluation.VerifyBenchmarkFreeze(collection, repositoryPath(collection.RepoRoot, *freezePath), evaluation.FreezeVerifyOptions{
				ClientVersion: clientVersion, ToolchainVersion: runtime.Version(),
				PrivateHoldoutPath: repositoryPath(collection.RepoRoot, *privateHoldout), HoldoutKeyPath: repositoryPath(collection.RepoRoot, *holdoutKey),
				PublicOnly: *split != "holdout",
			})
			if err != nil {
				return err
			}
		} else if *privateHoldout != "" || *holdoutKey != "" {
			return errors.New("private holdout execution requires -freeze")
		}
		arms, err := evaluation.FrozenMatrixArms(collection)
		if err != nil {
			return err
		}
		matrixOptions := evaluation.MatrixOptions{
			Runner: *runner, Model: *model, Kind: *kind, Split: *split, Case: *selectedCase, FixturesOnly: *fixturesOnly,
			ExplicitSkill: *explicit, Repetition: *repetition, Limit: *limit, Seed: *seed, Timeout: *timeout,
			MaxCells: *maxCells, MaxFailures: *maxFailures, StopOnTimeout: *stopOnTimeout,
			OutputPath: filepath.Join(collection.RepoRoot, filepath.FromSlash(*outputPath)),
		}
		if *freezePath != "" {
			if err := evaluation.BindMatrixFreeze(lock, freezeDigest, &matrixOptions); err != nil {
				return err
			}
		}
		written, err := evaluation.RunMatrix(context.Background(), collection, arms, matrixOptions)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "wrote %d globally randomized isolated matrix cells\n", written)
		return nil
	case "score":
		flags := flag.NewFlagSet("eval score", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		input := flags.String("input", "", "raw JSONL artifact")
		outputPath := flags.String("output", "", "scored JSONL artifact")
		judgmentsPath := flags.String("judgments", "", "resumable blinded semantic judgment JSONL artifact")
		judgeRunner := flags.String("judge-runner", "codex", "semantic evaluator runner")
		judgeModel := flags.String("judge-model", "", "semantic evaluator model; enables blinded judgment")
		judgeSeed := flags.Int64("judge-seed", 1, "semantic candidate randomization seed")
		judgeTimeout := flags.Duration("judge-timeout", 5*time.Minute, "per-candidate semantic evaluator timeout")
		maxFailures := flags.Int("max-failures", 1, "stop after this many evaluator failures; zero means unlimited")
		stopOnTimeout := flags.Bool("stop-on-timeout", true, "stop immediately when a candidate reaches its timeout")
		root := flags.String("root", "", "repository root for freeze verification")
		freezePath := flags.String("freeze", "", "release-candidate freeze lock")
		privateHoldout := flags.String("private-holdout", "", "private holdout JSON")
		holdoutKey := flags.String("holdout-key", "", "holdout HMAC key file")
		publicOnly := flags.Bool("public-only", false, "score only disclosed development cells without opening the holdout")
		if err := flags.Parse(arguments[1:]); err != nil {
			return err
		}
		if *input == "" || *outputPath == "" {
			return errors.New("eval score requires -input and -output")
		}
		if *freezePath != "" {
			collection, err := corpus.Load(*root)
			if err != nil {
				return err
			}
			clientVersion, err := authenticatedCodexVersion()
			if err != nil {
				return err
			}
			lock, _, digest, err := evaluation.VerifyBenchmarkFreeze(collection, repositoryPath(collection.RepoRoot, *freezePath), evaluation.FreezeVerifyOptions{
				ClientVersion: clientVersion, ToolchainVersion: runtime.Version(),
				PrivateHoldoutPath: repositoryPath(collection.RepoRoot, *privateHoldout), HoldoutKeyPath: repositoryPath(collection.RepoRoot, *holdoutKey),
				PublicOnly: *publicOnly,
			})
			if err != nil {
				return err
			}
			if err := evaluation.ValidateResultArtifactFreeze(*input, lock, digest); err != nil {
				return err
			}
			if *judgeModel != "" {
				if err := evaluation.ValidateJudgmentFreeze(lock, *judgeRunner, *judgeModel, *judgeSeed, *judgeTimeout); err != nil {
					return err
				}
			}
		} else if *privateHoldout != "" || *holdoutKey != "" {
			return errors.New("private holdout scoring requires -freeze")
		}
		if *judgeModel != "" {
			if *judgmentsPath == "" {
				return errors.New("eval score with -judge-model requires -judgments")
			}
			written, err := evaluation.RunJudgments(context.Background(), evaluation.JudgmentOptions{
				Runner: *judgeRunner, Model: *judgeModel, InputPath: *input, OutputPath: *judgmentsPath,
				Seed: *judgeSeed, Timeout: *judgeTimeout, MaxFailures: *maxFailures, StopOnTimeout: *stopOnTimeout,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(output, "wrote %d blinded semantic judgments\n", written)
		}
		var count int
		var err error
		if *judgmentsPath == "" {
			count, err = evaluation.ScoreFile(*input, *outputPath)
		} else {
			count, err = evaluation.ScoreFileWithJudgments(*input, *judgmentsPath, *outputPath)
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "scored %d evaluation cells\n", count)
		return nil
	case "report":
		flags := flag.NewFlagSet("eval report", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		input := flags.String("input", "", "scored JSONL artifact")
		arm := flags.String("arm", "", "optional arm selected from a mixed matrix artifact")
		against := flags.String("against", "", "optional scored JSONL competitor artifact for paired comparison")
		againstArm := flags.String("against-arm", "", "optional competitor arm selected from -against, or from -input when -against is omitted")
		outputPath := flags.String("output", "", "optional JSON report path; stdout when omitted")
		if err := flags.Parse(arguments[1:]); err != nil {
			return err
		}
		if *input == "" {
			return errors.New("eval report requires -input")
		}
		destination := output
		var outputFile *os.File
		if *outputPath != "" {
			if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
				return err
			}
			created, err := os.Create(*outputPath)
			if err != nil {
				return err
			}
			outputFile = created
			defer outputFile.Close()
			destination = outputFile
		}
		competitorInput := *against
		if competitorInput == "" && *againstArm != "" {
			competitorInput = *input
		}
		if competitorInput != "" {
			report, err := evaluation.CompareArms(*input, *arm, competitorInput, *againstArm)
			if err != nil {
				return err
			}
			return evaluation.WriteComparison(destination, report)
		}
		report, err := evaluation.ReportFileForArm(*input, *arm)
		if err != nil {
			return err
		}
		return evaluation.WriteReport(destination, report)
	default:
		return fmt.Errorf("unknown eval command %q", arguments[0])
	}
}

func loadSkillMap(path string) (map[string]string, error) {
	if path == "" {
		return nil, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read skill map: %w", err)
	}
	var result map[string]string
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("decode skill map: %w", err)
	}
	for canonical, armSkill := range result {
		if canonical == "" || armSkill == "" {
			return nil, fmt.Errorf("skill map entry %q must name a non-empty arm skill", canonical)
		}
	}
	return result, nil
}

func loadRoutingMap(path string) (map[string][]string, error) {
	if path == "" {
		return nil, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read routing map: %w", err)
	}
	var result map[string][]string
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("decode routing map: %w", err)
	}
	for key, routes := range result {
		if key == "" || len(routes) == 0 {
			return nil, fmt.Errorf("routing map entry %q must have at least one route", key)
		}
	}
	return result, nil
}

func authenticatedCodexVersion() (string, error) {
	for _, status := range evaluation.Preflight(context.Background()) {
		if status.Client != "codex" {
			continue
		}
		if !status.Available || !status.Authenticated || status.Version == "" {
			return "", fmt.Errorf("Codex runner is not ready for a frozen benchmark: %s", status.Evidence)
		}
		return status.Version, nil
	}
	return "", errors.New("Codex runner status is unavailable")
}

func sequentialSeeds(first int64, count int) []int64 {
	result := make([]int64, count)
	for index := range result {
		result[index] = first + int64(index)
	}
	return result
}

func splitList(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func repositoryPath(root, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

func runRelease(arguments []string, output io.Writer) error {
	if len(arguments) == 0 || arguments[0] != "check" {
		return errors.New("usage: skillctl release check [-root path] [-report path]")
	}
	flags := flag.NewFlagSet("release check", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", "", "repository root")
	reportPath := flags.String("report", "evaluations/reports/release-gates.json", "machine-evaluable leadership gate report")
	if err := flags.Parse(arguments[1:]); err != nil {
		return err
	}
	collection, err := corpus.Load(*root)
	if err != nil {
		return err
	}
	metrics, err := corpus.Validate(collection)
	if err != nil {
		return err
	}
	outputs, err := corpus.Render(collection)
	if err != nil {
		return err
	}
	if err := corpus.CheckGenerated(collection, outputs); err != nil {
		return err
	}
	absoluteReport := filepath.Join(collection.RepoRoot, filepath.FromSlash(*reportPath))
	_, blockers, err := evaluation.CheckReleaseEvidence(collection.RepoRoot, absoluteReport)
	if err != nil {
		return fmt.Errorf("leadership evidence gate failed: %w", err)
	}
	if len(blockers) > 0 {
		return fmt.Errorf("leadership evidence gate failed: %s", strings.Join(blockers, "; "))
	}
	fmt.Fprintln(output, "all structural and leadership evidence gates pass")
	writeMetrics(output, metrics)
	return nil
}

func runAudit(arguments []string, output io.Writer) error {
	if len(arguments) == 0 || arguments[0] != "refs" {
		return errors.New("usage: skillctl audit refs -refs path [-root path] [-verified-on YYYY-MM-DD]")
	}
	flags := flag.NewFlagSet("audit refs", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	referenceRoot := flags.String("refs", "", "directory containing reference repositories")
	repositoryRoot := flags.String("root", "", "golangskills.com repository root")
	verifiedOn := flags.String("verified-on", time.Now().UTC().Format("2006-01-02"), "verification date")
	if err := flags.Parse(arguments[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	root, err := corpus.FindRepoRoot(*repositoryRoot)
	if err != nil {
		return err
	}
	lock, err := research.Audit(*referenceRoot, *verifiedOn)
	if err != nil {
		return err
	}
	outputPath := filepath.Join(root, "research", "corpus-lock.json")
	if err := research.WriteLock(outputPath, lock); err != nil {
		return err
	}
	dispositionPath := filepath.Join(root, "knowledge", "claims", "reference-dispositions.json")
	index := research.BuildDispositionIndex(lock)
	if err := research.WriteDispositionIndex(dispositionPath, index); err != nil {
		return err
	}
	files, skills, items := 0, 0, 0
	for _, repository := range lock.Repositories {
		files += len(repository.Files)
		skills += len(repository.Skills)
		items += len(repository.MaterialItems)
	}
	fmt.Fprintf(output, "locked %d repositories, %d files, %d skills, and %d material items in %s; mapped every item in %s\n", len(lock.Repositories), files, skills, items, outputPath, dispositionPath)
	return nil
}

func writeMetrics(output io.Writer, metrics corpus.Metrics) {
	fmt.Fprintf(output, "skills: %d\n", metrics.SkillCount)
	fmt.Fprintf(output, "discovery characters: %d\n", metrics.DiscoveryCharacters)
	for _, collection := range metrics.Collections {
		fmt.Fprintf(output, "%s: %d skills, %d discovery characters\n", collection.Name, collection.SkillCount, collection.DiscoveryCharacters)
	}
	fmt.Fprintf(output, "SKILL.md lines: %d\n", metrics.SkillLines)
	fmt.Fprintf(output, "SKILL.md words: %d\n", metrics.SkillWords)
	fmt.Fprintf(output, "sources: %d\n", metrics.SourceCount)
	fmt.Fprintf(output, "claims: %d\n", metrics.ClaimCount)
	fmt.Fprintf(output, "evaluation cases: %d\n", metrics.EvaluationCount)
	fmt.Fprintf(output, "maximum description overlap: %.3f\n", metrics.MaxDescriptionOverlap)
}
