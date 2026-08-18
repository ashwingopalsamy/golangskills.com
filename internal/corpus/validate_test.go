package corpus

import (
	"os"
	"path/filepath"
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
	if metrics.SkillCount != 1 || metrics.EvaluationCount != 4 || metrics.SourceCount != 2 {
		t.Fatalf("metrics = %+v, want one skill, four evals, and two sources", metrics)
	}
}

func TestValidateAtRequiresNegativeRoutingCases(t *testing.T) {
	t.Parallel()

	collection := validTestCollection()
	collection.Skills[0].Evaluations.Cases = collection.Skills[0].Evaluations.Cases[:3]
	_, err := validateAt(collection, mustDate(t, "2026-08-18"))
	if err == nil || !strings.Contains(err.Error(), "one positive route, one negative route, and two quality cases") {
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

func TestValidateEvaluationsRejectsFixtureTestsVisibleToAgent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	fixture := filepath.Join(root, "evaluations", "fixtures", "example")
	oracle := filepath.Join(root, "evaluations", "oracles", "example")
	if err := os.MkdirAll(fixture, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(oracle, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture, "visible_test.go"), []byte("package example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oracle, "hidden_test.go"), []byte("package example\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	skill := validTestCollection().Skills[0]
	skill.Evaluations.Cases[2] = EvalCase{
		ID: "quality-one", Kind: "quality", Split: "development", Prompt: "Repair the example.",
		Fixture: "evaluations/fixtures/example", ExpectedInvariants: []string{"Find cause.", "Preserve invariant."},
		ForbiddenOutcomes: []string{"Guess."}, Graders: []Grader{{ID: "oracle", Kind: "go-test", Target: "./...", Weight: 1}},
	}
	issues := validateEvaluations("example: ", skill, root, map[string]struct{}{skill.Name: {}})
	if !containsIssue(issues, "fixture exposes a test before the agent exits") {
		t.Fatalf("validateEvaluations() issues = %v", issues)
	}
}

func containsIssue(issues []string, fragment string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, fragment) {
			return true
		}
	}
	return false
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
		{ID: "route-one", Kind: "routing", Split: "development", Prompt: "Example failure one", ShouldActivate: &activate, Reason: "In scope."},
		{ID: "avoid-one", Kind: "routing", Split: "development", Prompt: "Unrelated work one", ShouldActivate: &doNotActivate, Reason: "Out of scope."},
		{ID: "quality-one", Kind: "quality", Split: "development", Prompt: "Solve example one", ExpectedInvariants: []string{"Find cause.", "Preserve invariant."}, ForbiddenOutcomes: []string{"Guess."}, Graders: []Grader{{ID: "cause", Kind: "contains", Required: []string{"cause"}, Weight: 1}}},
		{ID: "quality-two", Kind: "quality", Split: "development", Prompt: "Solve example two", ExpectedInvariants: []string{"Bound work.", "Handle failure."}, ForbiddenOutcomes: []string{"Ignore failure."}, Graders: []Grader{{ID: "bound", Kind: "contains", Required: []string{"bound"}, Weight: 1}}},
	}
	return Collection{
		RepoRoot: "",
		Claims: ClaimLedger{SchemaVersion: 2, VerifiedOn: "2026-08-18", Claims: []Claim{{
			ID: "example-invariant", Statement: "Preserve example behavior.", Status: "adopted-with-qualifications",
			Scope: "Example", Invariant: "Example remains bounded.", Owners: []string{"example-skill"}, RiskDomains: []string{"correctness"},
			PrimaryEvidence: []ClaimEvidence{{Title: "Normative source", URL: "https://example.com/normative", Publisher: "Example", Kind: "normative", VerifiedOn: "2026-08-18"}},
		}}},
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
					SchemaVersion:    2,
					Collection:       "engineering-skills-for-go",
					DisplayName:      "Example Skill",
					ShortDescription: "Bound example production failures",
					DefaultPrompt:    "Use $example-skill to review this path.",
					Version:          "0.1.0",
					Maturity:         "beta",
					Category:         "language",
					RiskDomains:      []string{"correctness"},
					Tags:             []string{"example", "failure", "bounded"},
					GoVersions: GoVersions{
						Minimum:  "1.24",
						Guidance: []string{"1.25", "1.26"},
					},
					ClaimIDs: []string{"example-invariant"},
					CompatibilityEvidence: []CompatibilityEvidence{
						{Client: "codex", Level: "behaviorally-benchmarked", Contract: "test", VerifiedOn: "2026-08-18"},
						{Client: "claude-code", Level: "structurally-compatible", Contract: "test", VerifiedOn: "2026-08-18"},
						{Client: "cursor", Level: "structurally-compatible", Contract: "test", VerifiedOn: "2026-08-18"},
						{Client: "opencode", Level: "structurally-compatible", Contract: "test", VerifiedOn: "2026-08-18"},
					},
					SourceProvenance: SourceProvenance{Method: "independent-rewrite-from-primary-evidence", CorpusLock: "research/corpus-lock.json"},
					Sources:          sources,
				},
				Evaluations: Evaluations{SchemaVersion: 2, Skill: "example-skill", Cases: evals},
				Files: map[string][]byte{
					"SKILL.md":   markdown,
					"skill.json": []byte("{}"),
					"evals.json": []byte("{}"),
				},
			},
		},
	}
}

func TestValidateAtRequiresOwnedClaimEvidence(t *testing.T) {
	t.Parallel()

	collection := validTestCollection()
	collection.Skills[0].Metadata.Sources[0].URL = "https://example.com/unrelated"
	_, err := validateAt(collection, mustDate(t, "2026-08-18"))
	if err == nil || !strings.Contains(err.Error(), "has no matching primary evidence") {
		t.Fatalf("validateAt() error = %v, want claim-evidence error", err)
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
