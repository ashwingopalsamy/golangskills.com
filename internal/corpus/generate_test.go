package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderIsDeterministic(t *testing.T) {
	t.Parallel()

	collection := validTestCollection()
	first, err := Render(collection)
	if err != nil {
		t.Fatalf("Render() first error = %v", err)
	}
	second, err := Render(collection)
	if err != nil {
		t.Fatalf("Render() second error = %v", err)
	}
	if string(first["catalog/catalog.json"]) != string(second["catalog/catalog.json"]) {
		t.Fatal("Render() catalog output is not deterministic")
	}
	if _, exists := first["skills/example-skill/agents/openai.yaml"]; !exists {
		t.Fatal("Render() omitted the Codex skill manifest")
	}
}

func TestWriteGeneratedRemovesStalePluginSkillFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	collection := validTestCollection()
	collection.RepoRoot = root
	stale := filepath.Join(root, "plugins", "engineering-skills-for-go", "skills", "removed-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	outputs, err := Render(collection)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if err := WriteGenerated(collection, outputs); err != nil {
		t.Fatalf("WriteGenerated() error = %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale file still exists; Stat() error = %v", err)
	}
	if err := CheckGenerated(collection, outputs); err != nil {
		t.Fatalf("CheckGenerated() error = %v", err)
	}
}
