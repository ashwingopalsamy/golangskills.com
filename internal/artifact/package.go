// Package artifact creates reproducible skill-plugin archives and provenance.
package artifact

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var plugins = []string{
	"distributed-systems-skills-for-go",
	"engineering-skills-for-go",
	"fintech-skills-for-go",
}

// Package writes deterministic tar.gz archives, checksums, and provenance.
func Package(repoRoot, version string) ([]string, error) {
	dist := filepath.Join(repoRoot, "dist")
	if err := os.MkdirAll(dist, 0o755); err != nil {
		return nil, err
	}
	checksums := make(map[string]string)
	var paths []string
	for _, plugin := range plugins {
		filename := plugin + "-" + version + ".tar.gz"
		path := filepath.Join(dist, filename)
		if err := archive(filepath.Join(repoRoot, "plugins", plugin), path, plugin); err != nil {
			return nil, err
		}
		digest, err := fileHash(path)
		if err != nil {
			return nil, err
		}
		checksums[filename] = digest
		paths = append(paths, path)
	}
	var checksumText strings.Builder
	for _, plugin := range plugins {
		filename := plugin + "-" + version + ".tar.gz"
		fmt.Fprintf(&checksumText, "%s  %s\n", checksums[filename], filename)
	}
	checksumPath := filepath.Join(dist, "SHA256SUMS")
	if err := os.WriteFile(checksumPath, []byte(checksumText.String()), 0o644); err != nil {
		return nil, err
	}
	paths = append(paths, checksumPath)
	provenance := map[string]any{
		"schema_version": 1, "version": version, "source_commit": gitHead(repoRoot),
		"claims": "knowledge/claims/canonical.json", "corpus_lock": "research/corpus-lock.json",
		"archives": checksums, "reproducible": true,
	}
	content, err := json.MarshalIndent(provenance, "", "  ")
	if err != nil {
		return nil, err
	}
	provenancePath := filepath.Join(dist, "provenance.json")
	if err := os.WriteFile(provenancePath, append(content, '\n'), 0o644); err != nil {
		return nil, err
	}
	return append(paths, provenancePath), nil
}

func archive(source, destination, prefix string) error {
	var files []string
	if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refuse symlink %s", path)
		}
		if !entry.IsDir() {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(files)
	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer output.Close()
	gzipWriter, err := gzip.NewWriterLevel(output, gzip.BestCompression)
	if err != nil {
		return err
	}
	gzipWriter.Header.ModTime = time.Unix(0, 0).UTC()
	gzipWriter.Header.OS = 255
	tarWriter := tar.NewWriter(gzipWriter)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, file)
		if err != nil {
			return err
		}
		header := &tar.Header{
			Name: filepath.ToSlash(filepath.Join(prefix, relative)), Mode: 0o644,
			Size: int64(len(content)), ModTime: time.Unix(0, 0).UTC(), Format: tar.FormatPAX,
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tarWriter.Write(content); err != nil {
			return err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return err
	}
	return gzipWriter.Close()
}

func fileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func gitHead(root string) string {
	content, err := os.ReadFile(filepath.Join(root, ".git", "HEAD"))
	if err != nil {
		return "unknown"
	}
	value := strings.TrimSpace(string(content))
	if !strings.HasPrefix(value, "ref: ") {
		return value
	}
	ref := strings.TrimPrefix(value, "ref: ")
	content, err = os.ReadFile(filepath.Join(root, ".git", filepath.FromSlash(ref)))
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(content))
}
