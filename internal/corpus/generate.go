package corpus

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	collectionVersion = "0.1.0"
	pluginName        = "engineering-skills-for-go"
	pluginRoot        = "plugins/engineering-skills-for-go"
)

type catalog struct {
	SchemaVersion     int            `json:"schema_version"`
	CollectionVersion string         `json:"collection_version"`
	GeneratedFrom     string         `json:"generated_from"`
	Skills            []catalogSkill `json:"skills"`
}

type catalogSkill struct {
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	Description      string     `json:"description"`
	Compatibility    string     `json:"compatibility"`
	DisplayName      string     `json:"display_name"`
	ShortDescription string     `json:"short_description"`
	Version          string     `json:"version"`
	Status           string     `json:"status"`
	Category         string     `json:"category"`
	Tags             []string   `json:"tags"`
	GoVersions       GoVersions `json:"go_versions"`
	Relations        Relations  `json:"relations"`
	Sources          []Source   `json:"sources"`
	SourceCount      int        `json:"source_count"`
	EvaluationCount  int        `json:"evaluation_count"`
}

type pluginManifest struct {
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description"`
	Author      pluginAuthor    `json:"author"`
	Homepage    string          `json:"homepage"`
	Repository  string          `json:"repository"`
	License     string          `json:"license"`
	Keywords    []string        `json:"keywords"`
	Skills      string          `json:"skills"`
	Interface   pluginInterface `json:"interface"`
}

type pluginAuthor struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type pluginInterface struct {
	DisplayName      string   `json:"displayName"`
	ShortDescription string   `json:"shortDescription"`
	LongDescription  string   `json:"longDescription"`
	DeveloperName    string   `json:"developerName"`
	Category         string   `json:"category"`
	Capabilities     []string `json:"capabilities"`
	WebsiteURL       string   `json:"websiteURL"`
	DefaultPrompt    string   `json:"defaultPrompt"`
}

// Render builds every deterministic file derived from the canonical corpus.
// Map keys are slash-separated paths relative to the repository root.
func Render(collection Collection) (map[string][]byte, error) {
	outputs := make(map[string][]byte)

	for _, skill := range collection.Skills {
		openAI := renderOpenAI(skill)
		outputs["skills/"+skill.Name+"/agents/openai.yaml"] = openAI
		for filename, content := range skill.Files {
			if filename == "agents/openai.yaml" {
				continue
			}
			outputs[pluginRoot+"/skills/"+skill.Name+"/"+filename] = content
		}
		outputs[pluginRoot+"/skills/"+skill.Name+"/agents/openai.yaml"] = openAI
	}

	catalogContent, err := renderJSON(buildCatalog(collection))
	if err != nil {
		return nil, fmt.Errorf("render catalog: %w", err)
	}
	outputs["catalog/catalog.json"] = catalogContent

	manifestContent, err := renderJSON(buildPluginManifest())
	if err != nil {
		return nil, fmt.Errorf("render Codex plugin manifest: %w", err)
	}
	outputs[pluginRoot+"/.codex-plugin/plugin.json"] = manifestContent

	claudePluginContent, err := renderJSON(buildClaudePlugin())
	if err != nil {
		return nil, fmt.Errorf("render Claude plugin manifest: %w", err)
	}
	outputs[pluginRoot+"/.claude-plugin/plugin.json"] = claudePluginContent

	codexMarketplaceContent, err := renderJSON(buildCodexMarketplace())
	if err != nil {
		return nil, fmt.Errorf("render Codex marketplace: %w", err)
	}
	outputs[".agents/plugins/marketplace.json"] = codexMarketplaceContent

	claudeMarketplaceContent, err := renderJSON(buildClaudeMarketplace())
	if err != nil {
		return nil, fmt.Errorf("render Claude marketplace: %w", err)
	}
	outputs[".claude-plugin/marketplace.json"] = claudeMarketplaceContent
	return outputs, nil
}

// WriteGenerated replaces only generated artifacts and writes all rendered
// files atomically.
func WriteGenerated(collection Collection, outputs map[string][]byte) error {
	generatedSkills := filepath.Join(collection.RepoRoot, filepath.FromSlash(pluginRoot), "skills")
	if filepath.Base(generatedSkills) != "skills" || filepath.Base(filepath.Dir(generatedSkills)) != pluginName {
		return fmt.Errorf("refuse to replace unexpected generated directory %s", generatedSkills)
	}
	if err := os.RemoveAll(generatedSkills); err != nil {
		return fmt.Errorf("remove stale generated plugin skills: %w", err)
	}

	paths := sortedOutputPaths(outputs)
	for _, relative := range paths {
		absolute := filepath.Join(collection.RepoRoot, filepath.FromSlash(relative))
		if err := writeFileAtomic(absolute, outputs[relative]); err != nil {
			return fmt.Errorf("write %s: %w", relative, err)
		}
	}
	return nil
}

// CheckGenerated verifies that generated files are current and that the plugin
// skill tree contains no stale files.
func CheckGenerated(collection Collection, outputs map[string][]byte) error {
	var issues []string
	for _, relative := range sortedOutputPaths(outputs) {
		actual, err := os.ReadFile(filepath.Join(collection.RepoRoot, filepath.FromSlash(relative)))
		if err != nil {
			issues = append(issues, fmt.Sprintf("generated file %s: %v", relative, err))
			continue
		}
		if !bytes.Equal(actual, outputs[relative]) {
			issues = append(issues, fmt.Sprintf("generated file %s is stale", relative))
		}
	}

	prefix := pluginRoot + "/skills/"
	expectedPluginFiles := make(map[string]struct{})
	for relative := range outputs {
		if strings.HasPrefix(relative, prefix) {
			expectedPluginFiles[relative] = struct{}{}
		}
	}
	actualPluginFiles, err := walkRelativeFiles(collection.RepoRoot, filepath.Join(collection.RepoRoot, filepath.FromSlash(prefix)))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		issues = append(issues, fmt.Sprintf("inspect generated plugin skills: %v", err))
	}
	for _, relative := range actualPluginFiles {
		if _, expected := expectedPluginFiles[relative]; !expected {
			issues = append(issues, fmt.Sprintf("generated plugin file %s is stale", relative))
		}
	}

	sort.Strings(issues)
	if len(issues) > 0 {
		return &ValidationError{Issues: issues}
	}
	return nil
}

func renderOpenAI(skill Skill) []byte {
	var output strings.Builder
	output.WriteString("interface:\n")
	output.WriteString("  display_name: ")
	output.WriteString(strconv.Quote(skill.Metadata.DisplayName))
	output.WriteByte('\n')
	output.WriteString("  short_description: ")
	output.WriteString(strconv.Quote(skill.Metadata.ShortDescription))
	output.WriteByte('\n')
	output.WriteString("  default_prompt: ")
	output.WriteString(strconv.Quote(skill.Metadata.DefaultPrompt))
	output.WriteByte('\n')
	output.WriteString("policy:\n")
	output.WriteString("  allow_implicit_invocation: true\n")
	return []byte(output.String())
}

func buildCatalog(collection Collection) catalog {
	result := catalog{
		SchemaVersion:     1,
		CollectionVersion: collectionVersion,
		GeneratedFrom:     "skills/*/{SKILL.md,skill.json,evals.json,references/**}",
	}
	for _, skill := range collection.Skills {
		result.Skills = append(result.Skills, catalogSkill{
			Name:             skill.Name,
			Path:             "skills/" + skill.Name,
			Description:      skill.Frontmatter.Description,
			Compatibility:    skill.Frontmatter.Compatibility,
			DisplayName:      skill.Metadata.DisplayName,
			ShortDescription: skill.Metadata.ShortDescription,
			Version:          skill.Metadata.Version,
			Status:           skill.Metadata.Status,
			Category:         skill.Metadata.Category,
			Tags:             skill.Metadata.Tags,
			GoVersions:       skill.Metadata.GoVersions,
			Relations:        skill.Metadata.Relations,
			Sources:          skill.Metadata.Sources,
			SourceCount:      len(skill.Metadata.Sources),
			EvaluationCount:  len(skill.Evaluations.Cases),
		})
	}
	return result
}

func buildPluginManifest() pluginManifest {
	return pluginManifest{
		Name:        pluginName,
		Version:     collectionVersion,
		Description: "Focused production engineering skills for Go systems.",
		Author: pluginAuthor{
			Name: "Ashwin Gopalsamy",
			URL:  "https://github.com/ashwingopalsamy",
		},
		Homepage:   "https://golangskills.com",
		Repository: "https://github.com/ashwingopalsamy/golangskills.com",
		License:    "Apache-2.0",
		Keywords:   []string{"go", "golang", "distributed-systems", "reliability", "agent-skills"},
		Skills:     "./skills/",
		Interface: pluginInterface{
			DisplayName:      "Engineering Skills for Go",
			ShortDescription: "Production failure-boundary skills for Go.",
			LongDescription:  "Focused skills for concurrent execution, HTTP, SQL transactions, message delivery, resilience, and production change review in Go systems.",
			DeveloperName:    "Ashwin Gopalsamy",
			Category:         "Developer Tools",
			Capabilities:     []string{"Analyze Go production paths", "Design failure boundaries", "Review Go changes"},
			WebsiteURL:       "https://golangskills.com",
			DefaultPrompt:    "Review this Go production path with the most relevant Engineering Skills for Go skill.",
		},
	}
}

func buildClaudePlugin() map[string]any {
	return map[string]any{
		"name":        pluginName,
		"version":     collectionVersion,
		"description": "Focused production engineering skills for Go systems.",
		"author": map[string]string{
			"name": "Ashwin Gopalsamy",
			"url":  "https://github.com/ashwingopalsamy",
		},
		"homepage":   "https://golangskills.com",
		"repository": "https://github.com/ashwingopalsamy/golangskills.com",
		"license":    "Apache-2.0",
		"keywords":   []string{"go", "golang", "distributed-systems", "reliability", "agent-skills"},
	}
}

func buildCodexMarketplace() map[string]any {
	return map[string]any{
		"name": "golangskills",
		"interface": map[string]string{
			"displayName": "golangskills.com",
		},
		"plugins": []any{
			map[string]any{
				"name": pluginName,
				"source": map[string]string{
					"source": "local",
					"path":   "./" + pluginRoot,
				},
				"policy": map[string]string{
					"installation":   "AVAILABLE",
					"authentication": "ON_INSTALL",
				},
				"category": "Developer Tools",
			},
		},
	}
}

func buildClaudeMarketplace() map[string]any {
	return map[string]any{
		"$schema": "https://raw.githubusercontent.com/anthropics/claude-code/main/schemas/marketplace.schema.json",
		"name":    pluginName,
		"metadata": map[string]string{
			"description": "Production engineering skills for Go systems.",
		},
		"owner": map[string]string{
			"name": "Ashwin Gopalsamy",
			"url":  "https://github.com/ashwingopalsamy",
		},
		"plugins": []any{
			map[string]any{
				"name":        pluginName,
				"source":      "./" + pluginRoot,
				"description": "Focused production engineering skills for Go systems.",
				"version":     collectionVersion,
				"category":    "language-skills",
				"tags":        []string{"go", "golang", "distributed-systems", "reliability", "agent-skills"},
				"homepage":    "https://golangskills.com",
				"license":     "Apache-2.0",
			},
		},
	}
}

func renderJSON(value any) ([]byte, error) {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(content, '\n'), nil
}

func sortedOutputPaths(outputs map[string][]byte) []string {
	paths := make([]string, 0, len(outputs))
	for relative := range outputs {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	return paths
}

func writeFileAtomic(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".skillctl-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, path)
}

func walkRelativeFiles(root, start string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(start, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	sort.Strings(files)
	return files, err
}
