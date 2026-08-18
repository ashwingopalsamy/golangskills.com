// Package corpus loads, validates, and renders the canonical skill corpus.
package corpus

// Frontmatter is the portable metadata embedded at the start of SKILL.md.
type Frontmatter struct {
	Name          string
	Description   string
	License       string
	Compatibility string
}

// Metadata contains repository-specific catalog and provenance fields.
type Metadata struct {
	SchemaVersion    int        `json:"schema_version"`
	DisplayName      string     `json:"display_name"`
	ShortDescription string     `json:"short_description"`
	DefaultPrompt    string     `json:"default_prompt"`
	Version          string     `json:"version"`
	Status           string     `json:"status"`
	Category         string     `json:"category"`
	Tags             []string   `json:"tags"`
	GoVersions       GoVersions `json:"go_versions"`
	Relations        Relations  `json:"relations"`
	Sources          []Source   `json:"sources"`
}

// GoVersions records the minimum supported language version and newer versions
// whose features the skill discusses.
type GoVersions struct {
	Minimum  string   `json:"minimum"`
	Guidance []string `json:"guidance"`
}

// Relations records intentionally adjacent skills.
type Relations struct {
	Complements []string `json:"complements"`
	Overlaps    []string `json:"overlaps"`
}

// Source describes evidence supporting a skill's guidance.
type Source struct {
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	Publisher  string   `json:"publisher"`
	Kind       string   `json:"kind"`
	VerifiedOn string   `json:"verified_on"`
	Supports   []string `json:"supports"`
}

// Evaluations contains routing and response-quality cases for one skill.
type Evaluations struct {
	SchemaVersion int        `json:"schema_version"`
	Skill         string     `json:"skill"`
	Cases         []EvalCase `json:"cases"`
}

// EvalCase is either a routing decision or a quality rubric.
type EvalCase struct {
	ID             string   `json:"id"`
	Kind           string   `json:"kind"`
	Prompt         string   `json:"prompt"`
	ShouldActivate *bool    `json:"should_activate,omitempty"`
	Reason         string   `json:"reason,omitempty"`
	Criteria       []string `json:"criteria,omitempty"`
	AntiCriteria   []string `json:"anti_criteria,omitempty"`
}

// Skill is a fully loaded canonical skill.
type Skill struct {
	Name        string
	Dir         string
	Frontmatter Frontmatter
	Body        string
	Metadata    Metadata
	Evaluations Evaluations
	Files       map[string][]byte
}

// Collection is the complete canonical corpus rooted at RepoRoot.
type Collection struct {
	RepoRoot string
	Skills   []Skill
}

// Metrics summarizes corpus size and routing metadata.
type Metrics struct {
	SkillCount            int
	DiscoveryCharacters   int
	SkillLines            int
	SkillWords            int
	SourceCount           int
	EvaluationCount       int
	MaxDescriptionOverlap float64
}
