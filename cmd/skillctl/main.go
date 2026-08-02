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
		return errors.New("usage: skillctl eval <preflight|run|score|report> [options]")
	}
	switch arguments[0] {
	case "preflight":
		if len(arguments) != 1 {
			return errors.New("usage: skillctl eval preflight")
		}
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")
		return encoder.Encode(evaluation.SortedStatuses(evaluation.Preflight(context.Background())))
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
		selectedCase := flags.String("case", "", "one canonical skill/case key to execute")
		explicit := flags.Bool("explicit", false, "install and invoke only the expected canonical skill")
		limit := flags.Int("limit", 0, "maximum cases")
		seed := flags.Int64("seed", 1, "randomization seed")
		timeout := flags.Duration("timeout", 5*time.Minute, "per-case timeout")
		outputPath := flags.String("output", "evaluations/runs/manual.jsonl", "resumable JSONL artifact")
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
		routingMap, err := loadRoutingMap(*routingMapPath)
		if err != nil {
			return err
		}
		skillMap, err := loadSkillMap(*skillMapPath)
		if err != nil {
			return err
		}
		written, err := evaluation.Run(context.Background(), collection, evaluation.RunOptions{
			Runner: *runner, Model: *model, Arm: *arm, SkillsRoot: *skillsRoot, Kind: *kind, Case: *selectedCase,
			ExplicitSkill: *explicit, RoutingMap: routingMap, SkillMap: skillMap, Limit: *limit, Seed: *seed, Timeout: *timeout,
			OutputPath: filepath.Join(collection.RepoRoot, filepath.FromSlash(*outputPath)),
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "wrote %d isolated evaluation cells\n", written)
		return nil
	case "score":
		flags := flag.NewFlagSet("eval score", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		input := flags.String("input", "", "raw JSONL artifact")
		outputPath := flags.String("output", "", "scored JSONL artifact")
		if err := flags.Parse(arguments[1:]); err != nil {
			return err
		}
		if *input == "" || *outputPath == "" {
			return errors.New("eval score requires -input and -output")
		}
		count, err := evaluation.ScoreFile(*input, *outputPath)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "scored %d evaluation cells\n", count)
		return nil
	case "report":
		flags := flag.NewFlagSet("eval report", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		input := flags.String("input", "", "scored JSONL artifact")
		against := flags.String("against", "", "optional scored JSONL competitor artifact for paired comparison")
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
		if *against != "" {
			report, err := evaluation.CompareFiles(*input, *against)
			if err != nil {
				return err
			}
			return evaluation.WriteComparison(destination, report)
		}
		report, err := evaluation.ReportFile(*input)
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
