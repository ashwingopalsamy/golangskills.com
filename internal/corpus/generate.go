package corpus

import (
	"bytes"
	_ "embed"
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

const collectionVersion = "0.4.0"

//go:embed assets/generated/composer-icon.png
var composerIconPNG []byte

//go:embed assets/generated/engineering-logo.png
var engineeringLogoPNG []byte

//go:embed assets/generated/distributed-systems-logo.png
var distributedSystemsLogoPNG []byte

//go:embed assets/generated/fintech-logo.png
var fintechLogoPNG []byte

type collectionSpec struct {
	Name             string
	NPMSlug          string
	DisplayName      string
	ShortDescription string
	LongDescription  string
	Keywords         []string
	Capabilities     []string
	DefaultPrompts   []string
	Logo             []byte
}

var collectionSpecs = []collectionSpec{
	{
		Name:             "engineering-skills-for-go",
		NPMSlug:          "engineering",
		DisplayName:      "Go: Production Engineering",
		ShortDescription: "Build, test & operate Go",
		LongDescription:  "Evidence-backed Go language, API, service, testing, performance, security, operations, and engineering-review skills.",
		Keywords:         []string{"go", "golang", "backend", "api-development", "testing", "performance", "security", "code-review"},
		Capabilities:     []string{"Build production Go systems", "Test and diagnose Go software", "Review Go changes"},
		DefaultPrompts: []string{
			"Review this Go change for correctness and production risk. Report concrete, causal findings only; skip style comments.",
			"Diagnose this Go production issue from evidence. Find the failure mechanism, what to measure next, and the smallest safe fix.",
			"Design or implement this Go change. Preserve behavior and compatibility; make the smallest coherent change and verify it.",
		},
		Logo: engineeringLogoPNG,
	},
	{
		Name:             "distributed-systems-skills-for-go",
		NPMSlug:          "distributed-systems",
		DisplayName:      "Go: Distributed Systems",
		ShortDescription: "Build resilient Go systems",
		LongDescription:  "Invariant-driven Go concurrency, consistency, messaging, resilience, coordination, and distributed-change review skills.",
		Keywords:         []string{"go", "golang", "distributed-systems", "concurrency", "messaging", "consistency", "resilience", "coordination"},
		Capabilities:     []string{"Design failure-safe Go systems", "Diagnose distributed failures", "Review distributed changes"},
		DefaultPrompts: []string{
			"Review this distributed Go change for consistency, ordering, retries, and partial-failure risk. Report concrete, causal findings only; skip style comments.",
			"Diagnose this distributed Go production issue from evidence. Find the failure mechanism, violated invariant, and smallest safe recovery.",
			"Design or implement this distributed Go change. Bound concurrency, retries, and ownership; preserve safety under ambiguity and failure.",
		},
		Logo: distributedSystemsLogoPNG,
	},
	{
		Name:             "fintech-skills-for-go",
		NPMSlug:          "fintech",
		DisplayName:      "Go: Fintech",
		ShortDescription: "Build financially correct Go",
		LongDescription:  "Financial-integrity skills for money, ledgers, payment lifecycles, idempotency, settlement, reconciliation, security, and compliance.",
		Keywords:         []string{"go", "golang", "fintech", "payments", "ledger", "idempotency", "reconciliation", "financial-integrity"},
		Capabilities:     []string{"Build payment and ledger systems", "Preserve financial integrity", "Review fintech changes"},
		DefaultPrompts: []string{
			"Review this Go payment or ledger change for financial-integrity risk. Report concrete, causal findings only; skip style comments.",
			"Diagnose this fintech production issue from evidence. Find how money could be lost, duplicated, misstated, or concealed.",
			"Design or implement this Go financial workflow. Preserve ledger balance, replay safety, lifecycle validity, and auditability.",
		},
		Logo: fintechLogoPNG,
	},
}

type catalog struct {
	SchemaVersion     int                 `json:"schema_version"`
	CollectionVersion string              `json:"collection_version"`
	GeneratedFrom     []string            `json:"generated_from"`
	Collections       []catalogCollection `json:"collections"`
	Skills            []catalogSkill      `json:"skills"`
	Compatibility     []clientStatus      `json:"compatibility"`
	Benchmark         benchmarkStatus     `json:"benchmark"`
}

type catalogCollection struct {
	Name             string   `json:"name"`
	DisplayName      string   `json:"display_name"`
	Description      string   `json:"description"`
	Version          string   `json:"version"`
	NPMName          string   `json:"npm_name"`
	NPMRegistry      string   `json:"npm_registry"`
	SkillCount       int      `json:"skill_count"`
	Skills           []string `json:"skills"`
	PluginPath       string   `json:"plugin_path"`
	DiscoveryPath    string   `json:"discovery_path"`
	InstallationPath string   `json:"installation_path"`
}

type catalogSkill struct {
	Name                  string                  `json:"name"`
	Collection            string                  `json:"collection"`
	Path                  string                  `json:"path"`
	Description           string                  `json:"description"`
	Compatibility         string                  `json:"compatibility"`
	DisplayName           string                  `json:"display_name"`
	ShortDescription      string                  `json:"short_description"`
	Version               string                  `json:"version"`
	Maturity              string                  `json:"maturity"`
	Category              string                  `json:"category"`
	RiskDomains           []string                `json:"risk_domains"`
	Tags                  []string                `json:"tags"`
	GoVersions            GoVersions              `json:"go_versions"`
	ClaimIDs              []string                `json:"claim_ids"`
	Relations             Relations               `json:"relations"`
	CompatibilityEvidence []CompatibilityEvidence `json:"compatibility_evidence"`
	Sources               []Source                `json:"sources"`
	SourceCount           int                     `json:"source_count"`
	EvaluationCount       int                     `json:"evaluation_count"`
}

type clientStatus struct {
	Client   string `json:"client"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

type benchmarkStatus struct {
	LeadershipClaimEligible bool     `json:"leadership_claim_eligible"`
	Status                  string   `json:"status"`
	BlockingGates           []string `json:"blocking_gates"`
	EvidencePath            string   `json:"evidence_path"`
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
	DisplayName       string   `json:"displayName"`
	ShortDescription  string   `json:"shortDescription"`
	LongDescription   string   `json:"longDescription"`
	DeveloperName     string   `json:"developerName"`
	Category          string   `json:"category"`
	Capabilities      []string `json:"capabilities"`
	WebsiteURL        string   `json:"websiteURL"`
	PrivacyPolicyURL  string   `json:"privacyPolicyURL"`
	TermsOfServiceURL string   `json:"termsOfServiceURL"`
	DefaultPrompt     []string `json:"defaultPrompt"`
	BrandColor        string   `json:"brandColor"`
	ComposerIcon      string   `json:"composerIcon"`
	Logo              string   `json:"logo"`
}

// Render builds every deterministic file derived from the canonical corpus.
// Map keys are slash-separated paths relative to the repository root.
func Render(collection Collection) (map[string][]byte, error) {
	outputs := make(map[string][]byte)
	byCollection := skillsByCollection(collection.Skills)

	for _, skill := range collection.Skills {
		openAI := renderOpenAI(skill)
		outputs["skills/"+skill.Name+"/agents/openai.yaml"] = openAI
		for _, spec := range collectionSpecs {
			if skill.Metadata.Collection != spec.Name {
				continue
			}
			for filename, content := range skill.Files {
				if filename == "agents/openai.yaml" {
					continue
				}
				outputs[pluginSkillPath(spec.Name, skill.Name, filename)] = content
				outputs[adapterSkillPath("cursor", spec.Name, skill.Name, filename)] = content
				outputs[adapterSkillPath("opencode", spec.Name, skill.Name, filename)] = content
			}
			outputs[pluginSkillPath(spec.Name, skill.Name, "agents/openai.yaml")] = openAI
		}
	}

	for _, spec := range collectionSpecs {
		manifest, err := renderJSON(buildCodexPluginManifest(spec))
		if err != nil {
			return nil, fmt.Errorf("render %s Codex manifest: %w", spec.Name, err)
		}
		outputs["plugins/"+spec.Name+"/.codex-plugin/plugin.json"] = manifest

		claudeManifest, err := renderJSON(buildClaudePlugin(spec))
		if err != nil {
			return nil, fmt.Errorf("render %s Claude manifest: %w", spec.Name, err)
		}
		outputs["plugins/"+spec.Name+"/.claude-plugin/plugin.json"] = claudeManifest

		agentPlugin, err := renderJSON(buildAgentPlugin(spec))
		if err != nil {
			return nil, fmt.Errorf("render %s Agent Plugin manifest: %w", spec.Name, err)
		}
		outputs["plugins/"+spec.Name+"/plugin.json"] = agentPlugin
		outputs["plugins/"+spec.Name+"/assets/composer-icon.png"] = composerIconPNG
		outputs["plugins/"+spec.Name+"/assets/logo.png"] = spec.Logo

		adapterReadme := renderAdapterReadme(spec, byCollection[spec.Name])
		outputs["adapters/cursor/"+spec.Name+"/README.md"] = adapterReadme
		outputs["adapters/opencode/"+spec.Name+"/README.md"] = adapterReadme
	}

	builtCatalog := buildCatalog(collection)
	if err := addCatalogOutputs(outputs, builtCatalog, collection); err != nil {
		return nil, err
	}

	codexMarketplace, err := renderJSON(buildCodexMarketplace())
	if err != nil {
		return nil, fmt.Errorf("render Codex marketplace: %w", err)
	}
	outputs[".agents/plugins/marketplace.json"] = codexMarketplace
	claudeMarketplace, err := renderJSON(buildClaudeMarketplace())
	if err != nil {
		return nil, fmt.Errorf("render Claude marketplace: %w", err)
	}
	outputs[".claude-plugin/marketplace.json"] = claudeMarketplace
	outputs["llms.txt"] = renderLLMs(builtCatalog)
	return outputs, nil
}

func addCatalogOutputs(outputs map[string][]byte, builtCatalog catalog, collection Collection) error {
	content, err := renderJSON(builtCatalog)
	if err != nil {
		return fmt.Errorf("render catalog: %w", err)
	}
	outputs["catalog/catalog.json"] = content
	outputs["site/data/catalog.json"] = content

	collections, err := renderJSON(builtCatalog.Collections)
	if err != nil {
		return err
	}
	outputs["catalog/collections.json"] = collections
	skills, err := renderJSON(builtCatalog.Skills)
	if err != nil {
		return err
	}
	outputs["catalog/skills.json"] = skills
	compatibility, err := renderJSON(builtCatalog.Compatibility)
	if err != nil {
		return err
	}
	outputs["catalog/compatibility.json"] = compatibility
	benchmark, err := renderJSON(builtCatalog.Benchmark)
	if err != nil {
		return err
	}
	outputs["catalog/benchmark-status.json"] = benchmark
	provenance, err := renderJSON(map[string]any{
		"schema_version": 2,
		"corpus_lock":    "research/corpus-lock.json",
		"claim_ledger":   "knowledge/claims/canonical.json",
		"claim_count":    len(collection.Claims.Claims),
		"notices":        "THIRD_PARTY_NOTICES.md",
	})
	if err != nil {
		return err
	}
	outputs["catalog/provenance.json"] = provenance
	search := make([]map[string]any, 0, len(builtCatalog.Skills))
	for _, skill := range builtCatalog.Skills {
		search = append(search, map[string]any{
			"name": skill.Name, "collection": skill.Collection, "description": skill.Description,
			"tags": skill.Tags, "risk_domains": skill.RiskDomains,
		})
	}
	searchContent, err := renderJSON(search)
	if err != nil {
		return err
	}
	outputs["catalog/search-index.json"] = searchContent
	return nil
}

// WriteGenerated replaces only known generated roots and writes rendered files
// atomically.
func WriteGenerated(collection Collection, outputs map[string][]byte) error {
	for _, relative := range []string{"plugins", "adapters", "catalog", "site/data"} {
		absolute := filepath.Join(collection.RepoRoot, filepath.FromSlash(relative))
		if filepath.Clean(absolute) == filepath.Clean(collection.RepoRoot) {
			return fmt.Errorf("refuse to replace repository root for %s", relative)
		}
		if err := os.RemoveAll(absolute); err != nil {
			return fmt.Errorf("remove stale generated directory %s: %w", relative, err)
		}
	}
	for _, relative := range sortedOutputPaths(outputs) {
		if err := writeFileAtomic(filepath.Join(collection.RepoRoot, filepath.FromSlash(relative)), outputs[relative]); err != nil {
			return fmt.Errorf("write %s: %w", relative, err)
		}
	}
	return nil
}

// CheckGenerated verifies that every generated file is current and no generated
// root contains stale files.
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
	for _, prefix := range []string{"plugins/", "adapters/", "catalog/", "site/data/"} {
		expected := make(map[string]struct{})
		for relative := range outputs {
			if strings.HasPrefix(relative, prefix) {
				expected[relative] = struct{}{}
			}
		}
		actual, err := walkRelativeFiles(collection.RepoRoot, filepath.Join(collection.RepoRoot, filepath.FromSlash(prefix)))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			issues = append(issues, fmt.Sprintf("inspect generated root %s: %v", prefix, err))
		}
		for _, relative := range actual {
			if _, exists := expected[relative]; !exists {
				issues = append(issues, fmt.Sprintf("generated file %s is stale", relative))
			}
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
	output.WriteString("policy:\n  allow_implicit_invocation: true\n")
	return []byte(output.String())
}

func buildCatalog(collection Collection) catalog {
	result := catalog{
		SchemaVersion:     2,
		CollectionVersion: collectionVersion,
		GeneratedFrom: []string{
			"skills/*/{SKILL.md,skill.json,evals.json,references/**}",
			"knowledge/claims/canonical.json",
			"research/corpus-lock.json",
		},
		Compatibility: []clientStatus{
			{Client: "codex", Status: "behaviorally-benchmarked", Evidence: "evaluations/runs/"},
			{Client: "claude-code", Status: "structurally-compatible", Evidence: "catalog/compatibility.json"},
			{Client: "cursor", Status: "structurally-compatible", Evidence: "catalog/compatibility.json"},
			{Client: "opencode", Status: "structurally-compatible", Evidence: "catalog/compatibility.json"},
		},
		Benchmark: benchmarkStatus{
			LeadershipClaimEligible: false,
			Status:                  "evidence-pending",
			BlockingGates: []string{
				"private holdout and unrelated-prompt routing evidence pending",
				"full baseline and competitor fixture matrix pending",
				"paired superiority, domain non-inferiority, and baseline improvement pending",
				"quality token-efficiency frontier and complete traceability pending",
			},
			EvidencePath: "evaluations/reports/release-gates.json",
		},
	}
	byCollection := skillsByCollection(collection.Skills)
	for _, spec := range collectionSpecs {
		entry := catalogCollection{
			Name: spec.Name, DisplayName: spec.DisplayName, Description: spec.LongDescription,
			Version: collectionVersion, NPMName: "@golangskills/" + spec.NPMSlug,
			NPMRegistry:   "https://www.npmjs.com/package/@golangskills/" + spec.NPMSlug,
			PluginPath:    "plugins/" + spec.Name,
			DiscoveryPath: "skills/", InstallationPath: "plugins/" + spec.Name,
		}
		for _, skill := range byCollection[spec.Name] {
			entry.Skills = append(entry.Skills, skill.Name)
		}
		entry.SkillCount = len(entry.Skills)
		result.Collections = append(result.Collections, entry)
	}
	for _, skill := range collection.Skills {
		result.Skills = append(result.Skills, catalogSkill{
			Name: skill.Name, Collection: skill.Metadata.Collection, Path: "skills/" + skill.Name,
			Description: skill.Frontmatter.Description, Compatibility: skill.Frontmatter.Compatibility,
			DisplayName: skill.Metadata.DisplayName, ShortDescription: skill.Metadata.ShortDescription,
			Version: skill.Metadata.Version, Maturity: skill.Metadata.Maturity, Category: skill.Metadata.Category,
			RiskDomains: skill.Metadata.RiskDomains, Tags: skill.Metadata.Tags, GoVersions: skill.Metadata.GoVersions,
			ClaimIDs: skill.Metadata.ClaimIDs, Relations: skill.Metadata.Relations,
			CompatibilityEvidence: skill.Metadata.CompatibilityEvidence, Sources: skill.Metadata.Sources,
			SourceCount: len(skill.Metadata.Sources), EvaluationCount: len(skill.Evaluations.Cases),
		})
	}
	return result
}

func buildCodexPluginManifest(spec collectionSpec) pluginManifest {
	return pluginManifest{
		Name: spec.Name, Version: collectionVersion, Description: spec.LongDescription,
		Author:   pluginAuthor{Name: "Ashwin Gopalsamy", URL: "https://ashwingopalsamy.in"},
		Homepage: "https://golangskills.com", Repository: "https://github.com/ashwingopalsamy/golangskills.com",
		License: "Apache-2.0", Keywords: spec.Keywords, Skills: "./skills/",
		Interface: pluginInterface{
			DisplayName: spec.DisplayName, ShortDescription: spec.ShortDescription, LongDescription: spec.LongDescription,
			DeveloperName: "Ashwin Gopalsamy", Category: "Developer Tools",
			WebsiteURL:        "https://github.com/ashwingopalsamy/golangskills.com",
			PrivacyPolicyURL:  "https://github.com/ashwingopalsamy/golangskills.com/blob/main/docs/privacy.md",
			TermsOfServiceURL: "https://github.com/ashwingopalsamy/golangskills.com/blob/main/docs/terms.md",
			Capabilities:      spec.Capabilities,
			DefaultPrompt:     spec.DefaultPrompts,
			BrandColor:        "#071A2B", ComposerIcon: "./assets/composer-icon.png", Logo: "./assets/logo.png",
		},
	}
}

func buildClaudePlugin(spec collectionSpec) map[string]any {
	return map[string]any{
		"name": spec.Name, "version": collectionVersion, "description": spec.LongDescription,
		"author":   map[string]string{"name": "Ashwin Gopalsamy", "url": "https://ashwingopalsamy.in"},
		"homepage": "https://golangskills.com", "repository": "https://github.com/ashwingopalsamy/golangskills.com",
		"license": "Apache-2.0", "keywords": spec.Keywords,
	}
}

func buildAgentPlugin(spec collectionSpec) map[string]any {
	return map[string]any{
		"$schema": "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json",
		"name":    spec.Name, "description": spec.LongDescription, "version": collectionVersion,
		"author": map[string]string{"name": "Ashwin Gopalsamy"},
	}
}

func buildCodexMarketplace() map[string]any {
	plugins := make([]any, 0, len(collectionSpecs))
	for _, spec := range collectionSpecs {
		plugins = append(plugins, map[string]any{
			"name":     spec.Name,
			"source":   map[string]string{"source": "local", "path": "./plugins/" + spec.Name},
			"policy":   map[string]string{"installation": "AVAILABLE", "authentication": "ON_INSTALL"},
			"category": "Developer Tools",
		})
	}
	return map[string]any{
		"name": "golangskills", "interface": map[string]string{"displayName": "Go Engineering Skills by Ashwin Gopalsamy"},
		"plugins": plugins,
	}
}

func buildClaudeMarketplace() map[string]any {
	plugins := make([]any, 0, len(collectionSpecs))
	for _, spec := range collectionSpecs {
		plugins = append(plugins, map[string]any{
			"name": spec.Name, "source": "./plugins/" + spec.Name, "description": spec.LongDescription,
			"version": collectionVersion, "category": "language-skills", "tags": spec.Keywords,
			"homepage": "https://golangskills.com", "license": "Apache-2.0",
		})
	}
	return map[string]any{
		"$schema": "https://raw.githubusercontent.com/anthropics/claude-code/main/schemas/marketplace.schema.json",
		"name":    "golangskills", "metadata": map[string]string{"description": "Go Engineering Skills by Ashwin Gopalsamy."},
		"owner":   map[string]string{"name": "Ashwin Gopalsamy", "url": "https://ashwingopalsamy.in"},
		"plugins": plugins,
	}
}

func renderAdapterReadme(spec collectionSpec, skills []Skill) []byte {
	var output strings.Builder
	output.WriteString("# ")
	output.WriteString(spec.DisplayName)
	output.WriteString("\n\nGenerated portable installation layout. Do not edit this adapter; edit canonical `skills/` and run `skillctl generate`.\n\nSkills:\n")
	for _, skill := range skills {
		output.WriteString("\n- `")
		output.WriteString(skill.Name)
		output.WriteString("`")
	}
	output.WriteByte('\n')
	return []byte(output.String())
}

func renderLLMs(value catalog) []byte {
	var output strings.Builder
	output.WriteString("# Go Engineering Skills by Ashwin Gopalsamy\n\n")
	output.WriteString("> Evidence-backed Agent Skills for production Go, distributed systems, and fintech. Not affiliated with Google or the Go project.\n\n")
	for _, collection := range value.Collections {
		output.WriteString("## ")
		output.WriteString(collection.DisplayName)
		output.WriteString("\n\n")
		output.WriteString(collection.Description)
		output.WriteString("\n\nNPM package: ")
		output.WriteString(collection.NPMName)
		output.WriteString("\n\nCanonical catalog: catalog/catalog.json\n\n")
	}
	output.WriteString("## Evidence\n\n- Claim ledger: knowledge/claims/canonical.json\n- Reference lock: research/corpus-lock.json\n- Benchmark status: catalog/benchmark-status.json\n")
	return []byte(output.String())
}

func skillsByCollection(skills []Skill) map[string][]Skill {
	result := make(map[string][]Skill)
	for _, skill := range skills {
		result[skill.Metadata.Collection] = append(result[skill.Metadata.Collection], skill)
	}
	return result
}

func pluginSkillPath(collection, skill, filename string) string {
	return "plugins/" + collection + "/skills/" + skill + "/" + filename
}

func adapterSkillPath(client, collection, skill, filename string) string {
	directory := ".agents/skills/"
	if client == "cursor" {
		directory = ".cursor/skills/"
	}
	return "adapters/" + client + "/" + collection + "/" + directory + skill + "/" + filename
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

func walkRelativeFiles(root, directory string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(directory, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, filename)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
