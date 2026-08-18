package artifact

import (
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveIsDeterministic(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	source := filepath.Join(root, "plugin")
	if err := os.MkdirAll(filepath.Join(source, "skills", "example"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "skills", "example", "SKILL.md"), []byte("example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	left, right := filepath.Join(root, "left.tar.gz"), filepath.Join(root, "right.tar.gz")
	if err := archive(source, left, "plugin"); err != nil {
		t.Fatal(err)
	}
	if err := archive(source, right, "plugin"); err != nil {
		t.Fatal(err)
	}
	leftHash, _ := fileHash(left)
	rightHash, _ := fileHash(right)
	if leftHash != rightHash {
		t.Fatalf("hashes differ: %s != %s", leftHash, rightHash)
	}
}
