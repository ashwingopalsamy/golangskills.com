package corpus

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var errNotFrontmatter = errors.New("SKILL.md must start with YAML frontmatter")

// FindRepoRoot walks upward from start until it finds go.mod and skills/.
func FindRepoRoot(start string) (string, error) {
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
	}

	root, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve start directory: %w", err)
	}
	for {
		if regularFile(filepath.Join(root, "go.mod")) && directory(filepath.Join(root, "skills")) {
			return root, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", fmt.Errorf("find repository root from %s", start)
		}
		root = parent
	}
}

// Load reads every canonical skill below repoRoot/skills.
func Load(repoRoot string) (Collection, error) {
	root, err := FindRepoRoot(repoRoot)
	if err != nil {
		return Collection{}, err
	}

	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		return Collection{}, fmt.Errorf("read skills directory: %w", err)
	}

	collection := Collection{RepoRoot: root}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		skill, err := loadSkill(filepath.Join(root, "skills", entry.Name()))
		if err != nil {
			return Collection{}, fmt.Errorf("load %s: %w", entry.Name(), err)
		}
		collection.Skills = append(collection.Skills, skill)
	}
	if len(collection.Skills) == 0 {
		return Collection{}, errors.New("no canonical skills found")
	}
	sort.Slice(collection.Skills, func(i, j int) bool {
		return collection.Skills[i].Name < collection.Skills[j].Name
	})
	return collection, nil
}

func loadSkill(dir string) (Skill, error) {
	name := filepath.Base(dir)
	files, err := readSkillFiles(dir)
	if err != nil {
		return Skill{}, err
	}

	markdown, ok := files["SKILL.md"]
	if !ok {
		return Skill{}, errors.New("missing SKILL.md")
	}
	frontmatter, body, err := parseFrontmatter(markdown)
	if err != nil {
		return Skill{}, err
	}

	var metadata Metadata
	if err := decodeStrict(files["skill.json"], &metadata); err != nil {
		return Skill{}, fmt.Errorf("skill.json: %w", err)
	}
	var evaluations Evaluations
	if err := decodeStrict(files["evals.json"], &evaluations); err != nil {
		return Skill{}, fmt.Errorf("evals.json: %w", err)
	}

	return Skill{
		Name:        name,
		Dir:         dir,
		Frontmatter: frontmatter,
		Body:        body,
		Metadata:    metadata,
		Evaluations: evaluations,
		Files:       files,
	}, nil
}

func readSkillFiles(dir string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == dir {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not allowed: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = content
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read skill files: %w", err)
	}
	return files, nil
}

func parseFrontmatter(content []byte) (Frontmatter, string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	if !scanner.Scan() || scanner.Text() != "---" {
		return Frontmatter{}, "", errNotFrontmatter
	}

	values := make(map[string]string)
	foundEnd := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			foundEnd = true
			break
		}
		key, raw, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) != key || key == "" {
			return Frontmatter{}, "", fmt.Errorf("unsupported frontmatter line %q", line)
		}
		switch key {
		case "name", "description", "license", "compatibility":
		default:
			return Frontmatter{}, "", fmt.Errorf("unsupported frontmatter key %q", key)
		}
		if _, duplicate := values[key]; duplicate {
			return Frontmatter{}, "", fmt.Errorf("duplicate frontmatter key %q", key)
		}
		value, err := parseScalar(strings.TrimSpace(raw))
		if err != nil {
			return Frontmatter{}, "", fmt.Errorf("frontmatter %s: %w", key, err)
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return Frontmatter{}, "", fmt.Errorf("scan SKILL.md: %w", err)
	}
	if !foundEnd {
		return Frontmatter{}, "", errors.New("unterminated frontmatter")
	}

	bodyStart := bytes.Index(content[len("---\n"):], []byte("\n---\n"))
	if bodyStart < 0 {
		return Frontmatter{}, "", errors.New("frontmatter must use newline-delimited markers")
	}
	bodyStart += len("---\n") + len("\n---\n")
	return Frontmatter{
		Name:          values["name"],
		Description:   values["description"],
		License:       values["license"],
		Compatibility: values["compatibility"],
	}, string(content[bodyStart:]), nil
}

func parseScalar(value string) (string, error) {
	if value == "" {
		return "", errors.New("value is empty")
	}
	if strings.HasPrefix(value, "\"") {
		unquoted, err := strconv.Unquote(value)
		if err != nil {
			return "", fmt.Errorf("invalid quoted value: %w", err)
		}
		return unquoted, nil
	}
	if strings.ContainsAny(value, "[]{}&*!|>#`'") {
		return "", errors.New("only plain or double-quoted scalar values are supported")
	}
	return value, nil
}

func decodeStrict(content []byte, target any) error {
	if len(content) == 0 {
		return errors.New("missing or empty file")
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("multiple JSON values")
	}
	return nil
}

func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func directory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
