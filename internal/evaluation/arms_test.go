package evaluation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
	"github.com/ashwingopalsamy/golangskills.com/internal/research"
)

func TestArmValidationSeparatesPortableManifestFromInstalledSnapshots(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	referenceRoot := filepath.Join(repoRoot, "references")
	manifest := armManifest{SchemaVersion: 1, FrozenOn: "2026-08-24"}
	lock := research.CorpusLock{SchemaVersion: 1, VerifiedOn: "2026-08-24", ReferenceRoot: referenceRoot}

	for index := 0; index < 5; index++ {
		name := "competitor-" + string(rune('a'+index))
		repositoryRoot := filepath.Join(referenceRoot, name)
		skillsRoot := filepath.Join(repositoryRoot, "skills")
		mapPath := filepath.ToSlash(filepath.Join("evaluations", "arms", name+".skill-map.json"))
		manifest.Arms = append(manifest.Arms, armSpec{
			Name:             name,
			RepositoryCommit: "commit-" + name,
			SkillsRoot:       skillsRoot,
			SkillMap:         mapPath,
		})
		lock.Repositories = append(lock.Repositories, research.Repository{
			Name: name, Path: repositoryRoot, Commit: "commit-" + name, BenchmarkEligible: true,
			Skills: []research.Skill{{Name: "target-skill"}},
		})
		writeJSONForTest(t, filepath.Join(repoRoot, filepath.FromSlash(mapPath)), map[string]string{"canonical-skill": "target-skill"})
	}
	writeJSONForTest(t, filepath.Join(repoRoot, "evaluations", "arms", "manifest.json"), manifest)
	writeJSONForTest(t, filepath.Join(repoRoot, "research", "corpus-lock.json"), lock)
	collection := corpus.Collection{RepoRoot: repoRoot, Skills: []corpus.Skill{{Name: "canonical-skill"}}}

	if err := ValidateArmManifest(collection); err != nil {
		t.Fatalf("ValidateArmManifest() error = %v", err)
	}
	if err := ValidateArmFiles(collection); err == nil || !strings.Contains(err.Error(), "missing skill target-skill") {
		t.Fatalf("ValidateArmFiles() error = %v, want missing installed skill", err)
	}

	for _, arm := range manifest.Arms {
		skillPath := filepath.Join(arm.SkillsRoot, "target-skill", "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(skillPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(skillPath, []byte("# Target\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := ValidateArmFiles(collection); err != nil {
		t.Fatalf("ValidateArmFiles() after installation error = %v", err)
	}
}

func TestValidateArmManifestRejectsSkillsRootOutsideLockedRepository(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	referenceRoot := filepath.Join(repoRoot, "references")
	manifest := armManifest{SchemaVersion: 1, FrozenOn: "2026-08-24"}
	lock := research.CorpusLock{SchemaVersion: 1, VerifiedOn: "2026-08-24", ReferenceRoot: referenceRoot}
	for index := 0; index < 5; index++ {
		name := "competitor-" + string(rune('a'+index))
		repositoryRoot := filepath.Join(referenceRoot, name)
		skillsRoot := filepath.Join(repositoryRoot, "skills")
		if index == 0 {
			skillsRoot = filepath.Join(referenceRoot, "another-repository", "skills")
		}
		mapPath := filepath.ToSlash(filepath.Join("evaluations", "arms", name+".skill-map.json"))
		manifest.Arms = append(manifest.Arms, armSpec{Name: name, RepositoryCommit: "commit-" + name, SkillsRoot: skillsRoot, SkillMap: mapPath})
		lock.Repositories = append(lock.Repositories, research.Repository{
			Name: name, Path: repositoryRoot, Commit: "commit-" + name, BenchmarkEligible: true,
			Skills: []research.Skill{{Name: "target-skill"}},
		})
		writeJSONForTest(t, filepath.Join(repoRoot, filepath.FromSlash(mapPath)), map[string]string{"canonical-skill": "target-skill"})
	}
	writeJSONForTest(t, filepath.Join(repoRoot, "evaluations", "arms", "manifest.json"), manifest)
	writeJSONForTest(t, filepath.Join(repoRoot, "research", "corpus-lock.json"), lock)

	err := ValidateArmManifest(corpus.Collection{RepoRoot: repoRoot, Skills: []corpus.Skill{{Name: "canonical-skill"}}})
	if err == nil || !strings.Contains(err.Error(), "skills_root is outside its locked repository") {
		t.Fatalf("ValidateArmManifest() error = %v, want path-containment error", err)
	}
}

func TestReferenceRootOverrideRebasesInstalledSnapshots(t *testing.T) {
	repoRoot := t.TempDir()
	committedRoot := filepath.Join(string(filepath.Separator), "reference-checkouts")
	runtimeRoot := filepath.Join(t.TempDir(), "go-refs")
	manifest := armManifest{SchemaVersion: 1, FrozenOn: "2026-08-29"}
	lock := research.CorpusLock{SchemaVersion: 1, VerifiedOn: "2026-08-29", ReferenceRoot: committedRoot}

	for index := 0; index < 5; index++ {
		name := "competitor-" + string(rune('a'+index))
		committedRepositoryRoot := filepath.Join(committedRoot, name)
		mapPath := filepath.ToSlash(filepath.Join("evaluations", "arms", name+".skill-map.json"))
		manifest.Arms = append(manifest.Arms, armSpec{
			Name:             name,
			RepositoryCommit: "commit-" + name,
			SkillsRoot:       filepath.Join(committedRepositoryRoot, "skills"),
			SkillMap:         mapPath,
		})
		lock.Repositories = append(lock.Repositories, research.Repository{
			Name: name, Path: committedRepositoryRoot, Commit: "commit-" + name, BenchmarkEligible: true,
			Skills: []research.Skill{{Name: "target-skill"}},
		})
		writeJSONForTest(t, filepath.Join(repoRoot, filepath.FromSlash(mapPath)), map[string]string{"canonical-skill": "target-skill"})
		skillPath := filepath.Join(runtimeRoot, name, "skills", "target-skill", "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(skillPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(skillPath, []byte("# Target\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeJSONForTest(t, filepath.Join(repoRoot, "evaluations", "arms", "manifest.json"), manifest)
	writeJSONForTest(t, filepath.Join(repoRoot, "research", "corpus-lock.json"), lock)
	t.Setenv(referenceRootOverrideEnv, runtimeRoot)

	collection := corpus.Collection{RepoRoot: repoRoot, Skills: []corpus.Skill{{Name: "canonical-skill"}}}
	if err := ValidateArmFiles(collection); err != nil {
		t.Fatalf("ValidateArmFiles() with runtime override error = %v", err)
	}
	arms, err := FrozenMatrixArms(collection)
	if err != nil {
		t.Fatalf("FrozenMatrixArms() error = %v", err)
	}
	if len(arms) != 7 {
		t.Fatalf("FrozenMatrixArms() returned %d arms, want 7", len(arms))
	}
	for _, arm := range arms[2:] {
		if !pathWithin(runtimeRoot, arm.SkillsRoot) {
			t.Errorf("arm %s skills root %q is outside runtime root %q", arm.Arm, arm.SkillsRoot, runtimeRoot)
		}
	}
}

func TestReferenceRootOverrideRejectsRelativePath(t *testing.T) {
	repoRoot := t.TempDir()
	writeJSONForTest(t, filepath.Join(repoRoot, "evaluations", "arms", "manifest.json"), armManifest{SchemaVersion: 1, FrozenOn: "2026-08-29"})
	writeJSONForTest(t, filepath.Join(repoRoot, "research", "corpus-lock.json"), research.CorpusLock{
		SchemaVersion: 1,
		VerifiedOn:    "2026-08-29",
		ReferenceRoot: filepath.Join(string(filepath.Separator), "reference-checkouts"),
	})
	t.Setenv(referenceRootOverrideEnv, "relative/go-refs")

	_, _, err := loadRuntimeArmInputs(repoRoot)
	if err == nil || !strings.Contains(err.Error(), "must be an absolute path") {
		t.Fatalf("loadRuntimeArmInputs() error = %v, want absolute-path error", err)
	}
}

func writeJSONForTest(t *testing.T, path string, value any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
