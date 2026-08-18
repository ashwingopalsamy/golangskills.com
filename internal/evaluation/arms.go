package evaluation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
	"github.com/ashwingopalsamy/golangskills.com/internal/research"
)

type armManifest struct {
	SchemaVersion int       `json:"schema_version"`
	FrozenOn      string    `json:"frozen_on"`
	Arms          []armSpec `json:"arms"`
}

type armSpec struct {
	Name             string `json:"name"`
	RepositoryCommit string `json:"repository_commit"`
	SkillsRoot       string `json:"skills_root"`
	SkillMap         string `json:"skill_map"`
}

// ValidateArmFiles ensures frozen competitor mappings match the audited
// repository commits and reference actual skill directories.
func ValidateArmFiles(collection corpus.Collection) error {
	manifestPath := filepath.Join(collection.RepoRoot, "evaluations", "arms", "manifest.json")
	var manifest armManifest
	if err := decodeStrictFile(manifestPath, &manifest); err != nil {
		return err
	}
	if manifest.SchemaVersion != 1 || manifest.FrozenOn == "" {
		return fmt.Errorf("evaluation arm manifest requires schema version 1 and frozen_on")
	}

	var lock research.CorpusLock
	if err := decodeStrictFile(filepath.Join(collection.RepoRoot, "research", "corpus-lock.json"), &lock); err != nil {
		return err
	}
	locked := make(map[string]research.Repository, len(lock.Repositories))
	for _, repository := range lock.Repositories {
		locked[repository.Name] = repository
	}
	wantSkills := make(map[string]struct{}, len(collection.Skills))
	for _, skill := range collection.Skills {
		wantSkills[skill.Name] = struct{}{}
	}

	var issues []string
	seen := make(map[string]struct{}, len(manifest.Arms))
	for _, arm := range manifest.Arms {
		if _, duplicate := seen[arm.Name]; duplicate {
			issues = append(issues, "duplicate evaluation arm "+arm.Name)
		}
		seen[arm.Name] = struct{}{}
		repository, exists := locked[arm.Name]
		if !exists || repository.Commit != arm.RepositoryCommit || !repository.BenchmarkEligible {
			issues = append(issues, fmt.Sprintf("arm %s does not match an eligible corpus lock", arm.Name))
		}
		if !filepath.IsAbs(arm.SkillsRoot) {
			issues = append(issues, fmt.Sprintf("arm %s skills_root must be absolute", arm.Name))
		}
		var skillMap map[string]string
		mapPath := filepath.Join(collection.RepoRoot, filepath.FromSlash(arm.SkillMap))
		if err := decodeStrictFile(mapPath, &skillMap); err != nil {
			issues = append(issues, fmt.Sprintf("arm %s: %v", arm.Name, err))
			continue
		}
		for canonical := range wantSkills {
			target, exists := skillMap[canonical]
			if !exists || target == "" {
				issues = append(issues, fmt.Sprintf("arm %s is missing mapping for %s", arm.Name, canonical))
				continue
			}
			if target == "NONE" {
				continue
			}
			if _, err := os.Stat(filepath.Join(arm.SkillsRoot, target, "SKILL.md")); err != nil {
				issues = append(issues, fmt.Sprintf("arm %s maps %s to missing skill %s", arm.Name, canonical, target))
			}
		}
		for canonical := range skillMap {
			if _, exists := wantSkills[canonical]; !exists {
				issues = append(issues, fmt.Sprintf("arm %s has unknown canonical skill %s", arm.Name, canonical))
			}
		}
	}
	if len(manifest.Arms) != 4 {
		issues = append(issues, fmt.Sprintf("evaluation arm manifest has %d competitors; want 4", len(manifest.Arms)))
	}
	if len(issues) > 0 {
		sort.Strings(issues)
		return fmt.Errorf("evaluation arm validation failed: %s", strings.Join(issues, "; "))
	}
	return nil
}

func decodeStrictFile(path string, destination any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}
