package artifact

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	npmScope            = "@golangskills"
	npmRepositoryURL    = "git+https://github.com/ashwingopalsamy/golangskills.com.git"
	npmHomepage         = "https://golangskills.com"
	npmAuthorName       = "Ashwin Gopalsamy"
	npmAuthorURL        = "https://ashwingopalsamy.in"
	npmPackageSchema    = 1
	npmPackageDirectory = "dist/npm"
)

type npmCollection struct {
	Name        string   `json:"name"`
	NPMName     string   `json:"npm_name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
}

type npmPluginManifest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

type npmPackageJSON struct {
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	Description string           `json:"description"`
	Author      npmAuthor        `json:"author"`
	License     string           `json:"license"`
	Keywords    []string         `json:"keywords"`
	Repository  npmRepository    `json:"repository"`
	Homepage    string           `json:"homepage"`
	Files       []string         `json:"files"`
	Publish     npmPublishConfig `json:"publishConfig"`
}

type npmAuthor struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type npmRepository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type npmPublishConfig struct {
	Access string `json:"access"`
}

type npmPackageManifest struct {
	SchemaVersion    int               `json:"schema_version"`
	PackageName      string            `json:"package_name"`
	Collection       string            `json:"collection"`
	Version          string            `json:"version"`
	SourceRepository string            `json:"source_repository"`
	SourceCommit     string            `json:"source_commit"`
	Skills           []string          `json:"skills"`
	ClientLayouts    map[string]string `json:"client_layouts"`
	InstallSafety    []string          `json:"install_safety"`
}

// PackageNPM renders the three public, scoped npm package directories. The
// output is a data-only distribution: npm installation never runs a script or
// modifies an agent's configuration. Each package contains the generated
// plugin and client layouts plus a narrow package.json files allowlist.
func PackageNPM(repoRoot, version string) ([]string, error) {
	if strings.TrimSpace(version) == "" {
		return nil, errors.New("npm package version is required")
	}
	collections, err := readNPMCollections(repoRoot)
	if err != nil {
		return nil, err
	}

	distRoot := filepath.Join(repoRoot, npmPackageDirectory)
	if err := os.RemoveAll(distRoot); err != nil {
		return nil, fmt.Errorf("remove stale npm staging directory: %w", err)
	}
	if err := os.MkdirAll(distRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create npm staging directory: %w", err)
	}

	commit := gitHead(repoRoot)
	var paths []string
	for _, collection := range collections {
		if err := validateNPMCollection(collection); err != nil {
			return nil, err
		}
		stage := filepath.Join(distRoot, collection.Name)
		if err := renderNPMCollection(repoRoot, stage, collection, version, commit); err != nil {
			return nil, fmt.Errorf("render npm package %s: %w", collection.Name, err)
		}
		files, err := walkFiles(stage)
		if err != nil {
			return nil, err
		}
		paths = append(paths, files...)
	}
	sort.Strings(paths)
	return paths, nil
}

// CheckNPM verifies the generated package boundary before npm pack or publish.
// It deliberately rejects lifecycle scripts, dependencies, symlinks, and
// local-environment disclosures from the staged public packages.
func CheckNPM(repoRoot, version string) error {
	if strings.TrimSpace(version) == "" {
		return errors.New("npm package version is required")
	}
	collections, err := readNPMCollections(repoRoot)
	if err != nil {
		return err
	}
	for _, collection := range collections {
		stage := filepath.Join(repoRoot, npmPackageDirectory, collection.Name)
		if err := checkNPMCollection(stage, collection, version); err != nil {
			return fmt.Errorf("check npm package %s: %w", collection.Name, err)
		}
	}
	return nil
}

func readNPMCollections(repoRoot string) ([]npmCollection, error) {
	content, err := os.ReadFile(filepath.Join(repoRoot, "catalog", "collections.json"))
	if err != nil {
		return nil, fmt.Errorf("read catalog collections: %w", err)
	}
	var collections []npmCollection
	if err := json.Unmarshal(content, &collections); err != nil {
		return nil, fmt.Errorf("decode catalog collections: %w", err)
	}
	if len(collections) != 3 {
		return nil, fmt.Errorf("catalog collections: got %d, want 3", len(collections))
	}
	return collections, nil
}

func validateNPMCollection(collection npmCollection) error {
	if collection.Name == "" || filepath.Base(collection.Name) != collection.Name || strings.Contains(collection.Name, "..") {
		return fmt.Errorf("invalid collection name %q", collection.Name)
	}
	packageSlug := strings.TrimPrefix(collection.NPMName, npmScope+"/")
	if !strings.HasPrefix(collection.NPMName, npmScope+"/") || packageSlug == "" || strings.ContainsAny(packageSlug, "/\\") || strings.Contains(packageSlug, "..") {
		return fmt.Errorf("invalid npm package name %q", collection.NPMName)
	}
	if collection.DisplayName == "" || collection.Description == "" || len(collection.Skills) == 0 {
		return fmt.Errorf("collection %q has incomplete metadata", collection.Name)
	}
	return nil
}

func renderNPMCollection(repoRoot, stage string, collection npmCollection, version, commit string) error {
	pluginRoot := filepath.Join(repoRoot, "plugins", collection.Name)
	if err := copyTree(pluginRoot, stage); err != nil {
		return fmt.Errorf("copy generated plugin: %w", err)
	}

	for client, sourceRelative := range map[string]string{
		"cursor":   filepath.Join("adapters", "cursor", collection.Name, ".cursor"),
		"opencode": filepath.Join("adapters", "opencode", collection.Name, ".agents"),
	} {
		source := filepath.Join(repoRoot, sourceRelative)
		destination := filepath.Join(stage, client, filepath.Base(sourceRelative))
		if err := copyTree(source, destination); err != nil {
			return fmt.Errorf("copy %s layout: %w", client, err)
		}
	}

	pluginContent, err := os.ReadFile(filepath.Join(pluginRoot, ".codex-plugin", "plugin.json"))
	if err != nil {
		return fmt.Errorf("read plugin manifest: %w", err)
	}
	var plugin npmPluginManifest
	if err := json.Unmarshal(pluginContent, &plugin); err != nil {
		return fmt.Errorf("decode plugin manifest: %w", err)
	}
	keywords := append([]string{}, plugin.Keywords...)
	keywords = append(keywords, "agent-skills", "codex", "claude-code", "cursor", "opencode")
	keywords = uniqueStrings(keywords)

	packageName := collection.NPMName
	packageJSON := npmPackageJSON{
		Name:        packageName,
		Version:     version,
		Description: plugin.Description,
		Author:      npmAuthor{Name: npmAuthorName, URL: npmAuthorURL},
		License:     "Apache-2.0",
		Keywords:    keywords,
		Repository:  npmRepository{Type: "git", URL: npmRepositoryURL},
		Homepage:    npmHomepage,
		Files: []string{
			"skills/", ".codex-plugin/", ".claude-plugin/", "plugin.json",
			"cursor/", "opencode/", "manifest.json", "README.md",
			"LICENSE", "THIRD_PARTY_NOTICES.md",
		},
		Publish: npmPublishConfig{Access: "public"},
	}
	if err := writeJSON(filepath.Join(stage, "package.json"), packageJSON); err != nil {
		return fmt.Errorf("write package.json: %w", err)
	}

	manifest := npmPackageManifest{
		SchemaVersion:    npmPackageSchema,
		PackageName:      packageName,
		Collection:       collection.Name,
		Version:          version,
		SourceRepository: "https://github.com/ashwingopalsamy/golangskills.com",
		SourceCommit:     commit,
		Skills:           append([]string{}, collection.Skills...),
		ClientLayouts: map[string]string{
			"codex":       ".codex-plugin/ plus skills/",
			"claude-code": ".claude-plugin/ plus skills/",
			"cursor":      "cursor/.cursor/skills/",
			"opencode":    "opencode/.agents/skills/",
		},
		InstallSafety: []string{
			"data-only package with no npm lifecycle scripts",
			"no network access after package download",
			"agent configuration is unchanged until the user copies or installs a layout",
		},
	}
	if err := writeJSON(filepath.Join(stage, "manifest.json"), manifest); err != nil {
		return fmt.Errorf("write manifest.json: %w", err)
	}

	for _, filename := range []string{"LICENSE", "THIRD_PARTY_NOTICES.md"} {
		content, err := os.ReadFile(filepath.Join(repoRoot, filename))
		if err != nil {
			return fmt.Errorf("read %s: %w", filename, err)
		}
		if filename == "THIRD_PARTY_NOTICES.md" {
			content = []byte(strings.ReplaceAll(string(content),
				"](research/corpus-lock.json)",
				"](https://github.com/ashwingopalsamy/golangskills.com/blob/main/research/corpus-lock.json)"))
		}
		if err := writeFile(filepath.Join(stage, filename), content); err != nil {
			return fmt.Errorf("write %s: %w", filename, err)
		}
	}
	if err := writeFile(filepath.Join(stage, "README.md"), []byte(renderNPMReadme(collection, packageName, version))); err != nil {
		return fmt.Errorf("write README.md: %w", err)
	}
	return nil
}

func renderNPMReadme(collection npmCollection, packageName, version string) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# %s\n\n", collection.DisplayName)
	fmt.Fprintf(&output, "`%s@%s` is a versioned Agent Skills collection for production Go work. It is authored by [Ashwin Gopalsamy](%s) and distributed under Apache-2.0.\n\n", packageName, version, npmAuthorURL)
	output.WriteString("## Install\n\n")
	fmt.Fprintf(&output, "```sh\nnpm install %s@%s\n```\n\n", packageName, version)
	output.WriteString("The package is data-only. npm installation does not run a lifecycle script or change an agent configuration.\n\n")
	output.WriteString("## Client layouts\n\n")
	output.WriteString("- Codex: use the package root as a plugin source; it contains `.codex-plugin/plugin.json` and `skills/`.\n")
	output.WriteString("- Claude Code: use the package root as a plugin source; it contains `.claude-plugin/plugin.json` and `skills/`.\n")
	output.WriteString("- Cursor: copy `cursor/.cursor/skills/` into the project’s `.cursor/skills/` directory.\n")
	output.WriteString("- OpenCode: copy `opencode/.agents/skills/` into the project’s `.agents/skills/` directory.\n\n")
	output.WriteString("## GitHub discovery\n\n")
	output.WriteString("The open Skills CLI installs directly from the canonical repository:\n\n")
	output.WriteString("```sh\nnpx skills add ashwingopalsamy/golangskills.com\n```\n\n")
	output.WriteString("## Contents\n\n")
	for _, skill := range collection.Skills {
		fmt.Fprintf(&output, "- `%s`\n", skill)
	}
	output.WriteString("\nSource repository: https://github.com/ashwingopalsamy/golangskills.com\n")
	return output.String()
}

func checkNPMCollection(stage string, collection npmCollection, version string) error {
	packagePath := filepath.Join(stage, "package.json")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return fmt.Errorf("read package.json: %w", err)
	}
	var packageJSON npmPackageJSON
	if err := json.Unmarshal(content, &packageJSON); err != nil {
		return fmt.Errorf("decode package.json: %w", err)
	}
	if packageJSON.Name != collection.NPMName {
		return fmt.Errorf("package name %q is not %q", packageJSON.Name, collection.NPMName)
	}
	if packageJSON.Version != version {
		return fmt.Errorf("package version %q is not %q", packageJSON.Version, version)
	}
	if packageJSON.Publish.Access != "public" {
		return fmt.Errorf("package access is %q, want public", packageJSON.Publish.Access)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(content, &raw); err != nil {
		return fmt.Errorf("decode package.json fields: %w", err)
	}
	for _, forbidden := range []string{"scripts", "dependencies", "devDependencies", "optionalDependencies", "peerDependencies", "bin"} {
		if _, exists := raw[forbidden]; exists {
			return fmt.Errorf("package.json contains forbidden field %q", forbidden)
		}
	}
	for _, required := range []string{
		"README.md", "LICENSE", "THIRD_PARTY_NOTICES.md", "manifest.json", "plugin.json",
		".codex-plugin/plugin.json", ".claude-plugin/plugin.json", "assets/composer-icon.png", "assets/logo.png", "skills",
		"cursor/.cursor/skills", "opencode/.agents/skills",
	} {
		info, err := os.Stat(filepath.Join(stage, filepath.FromSlash(required)))
		if err != nil {
			return fmt.Errorf("required path %s: %w", required, err)
		}
		if required != "skills" && !strings.HasSuffix(required, "/skills") && info.IsDir() {
			return fmt.Errorf("required file %s is a directory", required)
		}
	}
	allowedPrefixes := []string{
		"skills/", ".codex-plugin/", ".claude-plugin/", "assets/", "cursor/", "opencode/",
		"plugin.json", "manifest.json", "package.json", "README.md", "LICENSE", "THIRD_PARTY_NOTICES.md",
	}
	forbiddenContent := []string{
		"/Users/", "CodexSwitch", ".private/", "authenticated as", "Cursor CLI: absent", "OpenCode CLI: absent",
	}
	return filepath.WalkDir(stage, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic link %s is not allowed", path)
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(stage, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		allowed := false
		for _, prefix := range allowedPrefixes {
			if relative == strings.TrimSuffix(prefix, "/") || strings.HasPrefix(relative, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file %s is outside the package allowlist", relative)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, forbidden := range forbiddenContent {
			if strings.Contains(string(content), forbidden) {
				return fmt.Errorf("file %s contains forbidden content %q", relative, forbidden)
			}
		}
		return nil
	})
}

func copyTree(source, destination string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("source %s is not a directory", source)
	}
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not allowed: %s", path)
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return os.MkdirAll(destination, 0o755)
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeFile(target, content)
	})
}

func writeJSON(path string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, append(content, '\n'))
}

func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func walkFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
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
		files = append(files, filepath.ToSlash(filepath.Join(root, relative)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
