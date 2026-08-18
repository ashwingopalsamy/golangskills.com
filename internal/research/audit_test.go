package research

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuditLocksFilesAndMaterialItems(t *testing.T) {
	t.Parallel()

	referenceRoot := t.TempDir()
	repository := filepath.Join(referenceRoot, "sample")
	skillDir := filepath.Join(repository, "skills", "go-concurrency")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(repository, "LICENSE"), "Permission is hereby granted, free of charge, to any person obtaining a copy")
	writeTestFile(t, filepath.Join(skillDir, "SKILL.md"), "# Concurrency\n\nYou should bound every worker pool.\n\n```go\ngo work()\n```\n")

	lock, err := Audit(referenceRoot, "2026-08-18")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(lock.Repositories); got != 1 {
		t.Fatalf("repositories = %d, want 1", got)
	}
	got := lock.Repositories[0]
	if got.License.SPDX != "MIT" || got.CanonicalSkillCount != 1 {
		t.Fatalf("repository metadata = %#v", got)
	}
	if got.MaterialItemCount != 2 {
		t.Fatalf("material items = %d, want 2", got.MaterialItemCount)
	}

	index := BuildDispositionIndex(lock)
	if index.MaterialItems != got.MaterialItemCount {
		t.Fatalf("dispositions = %d, want %d", index.MaterialItems, got.MaterialItemCount)
	}
	for _, disposition := range index.DispositionRows {
		if disposition.ClaimID != "dist-bounded-concurrency" {
			t.Fatalf("claim = %q, want dist-bounded-concurrency", disposition.ClaimID)
		}
	}
}

func TestCanonicalSkillsExcludesGeneratedCopies(t *testing.T) {
	t.Parallel()

	skills := canonicalSkills([]Skill{
		{Name: "go-context", Path: ".claude-plugin/skills/go-context/SKILL.md"},
		{Name: "go-context", Path: "skills/go-context/SKILL.md"},
	})
	if len(skills) != 1 || skills[0].Path != "skills/go-context/SKILL.md" {
		t.Fatalf("canonical skills = %#v", skills)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
