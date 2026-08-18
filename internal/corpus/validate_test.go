package corpus

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestValidateAtAcceptsCompleteSkill(t *testing.T) {
	t.Parallel()

	collection := validTestCollection()
	metrics, err := validateAt(collection, mustDate(t, "2026-08-18"))
	if err != nil {
		t.Fatalf("validateAt() error = %v", err)
	}
	if metrics.SkillCount != 1 || metrics.EvaluationCount != 6 || metrics.SourceCount != 2 {
		t.Fatalf("metrics = %+v, want one skill, six evals, and two sources", metrics)
	}
}

func TestValidateAtRequiresNegativeRoutingCases(t *testing.T) {
	t.Parallel()

	collection := validTestCollection()
	collection.Skills[0].Evaluations.Cases = collection.Skills[0].Evaluations.Cases[:4]
	_, err := validateAt(collection, mustDate(t, "2026-08-18"))
	if err == nil || !strings.Contains(err.Error(), "two positive routes, two negative routes, and two quality cases") {
		t.Fatalf("validateAt() error = %v, want balanced-eval error", err)
	}
}

func TestValidateAtRejectsStaleSource(t *testing.T) {
	t.Parallel()

	collection := validTestCollection()
	collection.Skills[0].Metadata.Sources[0].VerifiedOn = "2025-01-01"
	_, err := validateAt(collection, mustDate(t, "2026-08-18"))
	if err == nil || !strings.Contains(err.Error(), "older than 400 days") {
		t.Fatalf("validateAt() error = %v, want stale-source error", err)
	}
}

func TestJaccardIgnoresRoutingBoilerplate(t *testing.T) {
	t.Parallel()

	left := descriptionTerms("Use for Go HTTP client timeouts. Do not use for message delivery.")
	right := descriptionTerms("Use for SQL transaction isolation. Do not use for HTTP clients.")
	if overlap := jaccard(left, right); overlap >= 0.25 {
		t.Fatalf("jaccard() = %.3f, want low semantic overlap", overlap)
	}
}

func TestValidateRepositoryLinksRejectsMissingTarget(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(root+"/README.md", []byte("[missing](docs/missing.md)\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	issues := validateRepositoryLinks(root)
	if len(issues) != 1 || !strings.Contains(issues[0], "missing link target") {
		t.Fatalf("validateRepositoryLinks() issues = %v, want one missing-target issue", issues)
	}
}

func validTestCollection() Collection {
	activate := true
	doNotActivate := false
	description := "Design or review bounded example Go work. Use for example failures. Do not use for unrelated sequential formatting."
	markdown := []byte("---\nname: example-skill\ndescription: \"" + description + "\"\nlicense: Apache-2.0\ncompatibility: Go 1.24 or newer.\n---\n\n# Example skill\n")
	sources := []Source{
		{
			Title:      "Normative source",
			URL:        "https://example.com/normative",
			Publisher:  "Example",
			Kind:       "normative",
			VerifiedOn: "2026-08-18",
			Supports:   []string{"example invariant"},
		},
		{
			Title:      "Primary source",
			URL:        "https://example.com/primary",
			Publisher:  "Example",
			Kind:       "primary",
			VerifiedOn: "2026-08-18",
			Supports:   []string{"example failure"},
		},
	}
	evals := []EvalCase{
		{ID: "route-one", Kind: "routing", Prompt: "Example failure one", ShouldActivate: &activate, Reason: "In scope."},
		{ID: "route-two", Kind: "routing", Prompt: "Example failure two", ShouldActivate: &activate, Reason: "In scope."},
		{ID: "avoid-one", Kind: "routing", Prompt: "Unrelated work one", ShouldActivate: &doNotActivate, Reason: "Out of scope."},
		{ID: "avoid-two", Kind: "routing", Prompt: "Unrelated work two", ShouldActivate: &doNotActivate, Reason: "Out of scope."},
		{ID: "quality-one", Kind: "quality", Prompt: "Solve example one", Criteria: []string{"Find cause.", "Preserve invariant."}, AntiCriteria: []string{"Guess."}},
		{ID: "quality-two", Kind: "quality", Prompt: "Solve example two", Criteria: []string{"Bound work.", "Handle failure."}, AntiCriteria: []string{"Ignore failure."}},
	}
	return Collection{
		RepoRoot: "",
		Skills: []Skill{
			{
				Name: "example-skill",
				Frontmatter: Frontmatter{
					Name:          "example-skill",
					Description:   description,
					License:       "Apache-2.0",
					Compatibility: "Go 1.24 or newer.",
				},
				Metadata: Metadata{
					SchemaVersion:    1,
					DisplayName:      "Example Skill",
					ShortDescription: "Bound example production failures",
					DefaultPrompt:    "Use $example-skill to review this path.",
					Version:          "0.1.0",
					Status:           "beta",
					Category:         "execution",
					Tags:             []string{"example", "failure", "bounded"},
					GoVersions: GoVersions{
						Minimum:  "1.24",
						Guidance: []string{"1.25", "1.26"},
					},
					Sources: sources,
				},
				Evaluations: Evaluations{SchemaVersion: 1, Skill: "example-skill", Cases: evals},
				Files: map[string][]byte{
					"SKILL.md":   markdown,
					"skill.json": []byte("{}"),
					"evals.json": []byte("{}"),
				},
			},
		},
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}
	return parsed
}
