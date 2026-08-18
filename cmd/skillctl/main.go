// Command skillctl validates and generates the Engineering Skills for Go corpus.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
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
	fmt.Fprintf(output, "SKILL.md lines: %d\n", metrics.SkillLines)
	fmt.Fprintf(output, "SKILL.md words: %d\n", metrics.SkillWords)
	fmt.Fprintf(output, "sources: %d\n", metrics.SourceCount)
	fmt.Fprintf(output, "evaluation cases: %d\n", metrics.EvaluationCount)
	fmt.Fprintf(output, "maximum description overlap: %.3f\n", metrics.MaxDescriptionOverlap)
}
