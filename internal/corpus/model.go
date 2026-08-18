// Package corpus loads, validates, and renders the canonical skill corpus.
package corpus

// Frontmatter is the portable metadata embedded at the start of SKILL.md.
type Frontmatter struct {
	Name          string
	Description   string
	License       string
	Compatibility string
}

// Metadata is schema v2 for catalog, compatibility, ownership, and provenance.
type Metadata struct {
	SchemaVersion         int                     `json:"schema_version"`
	Collection            string                  `json:"collection"`
	DisplayName           string                  `json:"display_name"`
	ShortDescription      string                  `json:"short_description"`
	DefaultPrompt         string                  `json:"default_prompt"`
	Version               string                  `json:"version"`
	Maturity              string                  `json:"maturity"`
	Category              string                  `json:"category"`
	RiskDomains           []string                `json:"risk_domains"`
	Tags                  []string                `json:"tags"`
	GoVersions            GoVersions              `json:"go_versions"`
	ClaimIDs              []string                `json:"claim_ids"`
	Relations             Relations               `json:"relations"`
	CompatibilityEvidence []CompatibilityEvidence `json:"compatibility_evidence"`
	SourceProvenance      SourceProvenance        `json:"source_provenance"`
	Sources               []Source                `json:"sources"`
}

// GoVersions records the tooling floor, current guidance targets, and explicit
// legacy language versions discussed only for compatibility.
type GoVersions struct {
	Minimum  string   `json:"minimum"`
	Guidance []string `json:"guidance"`
	Legacy   []string `json:"legacy,omitempty"`
}

// Relations records intentionally adjacent skills.
type Relations struct {
	Complements []string `json:"complements"`
	Overlaps    []string `json:"overlaps"`
}

// CompatibilityEvidence records current client packaging evidence without
// implying behavioral superiority.
type CompatibilityEvidence struct {
	Client     string `json:"client"`
	Level      string `json:"level"`
	Contract   string `json:"contract"`
	VerifiedOn string `json:"verified_on"`
}

// SourceProvenance states how the canonical writing relates to references.
type SourceProvenance struct {
	Method                string   `json:"method"`
	ReferenceRepositories []string `json:"reference_repositories"`
	CorpusLock            string   `json:"corpus_lock"`
}

// Source describes primary evidence supporting a skill's claims.
type Source struct {
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	Publisher  string   `json:"publisher"`
	Kind       string   `json:"kind"`
	VerifiedOn string   `json:"verified_on"`
	Supports   []string `json:"supports"`
}

// Evaluations is schema v2 for routing, deterministic fixtures, and semantic
// rubrics. Development and holdout cases are kept distinguishable.
type Evaluations struct {
	SchemaVersion int        `json:"schema_version"`
	Skill         string     `json:"skill"`
	Cases         []EvalCase `json:"cases"`
}

// EvalCase is either a routing decision or a quality fixture.
type EvalCase struct {
	ID                 string   `json:"id"`
	Kind               string   `json:"kind"`
	Split              string   `json:"split"`
	Prompt             string   `json:"prompt"`
	ShouldActivate     *bool    `json:"should_activate,omitempty"`
	Reason             string   `json:"reason,omitempty"`
	ConfusesWith       []string `json:"confuses_with,omitempty"`
	Fixture            string   `json:"fixture,omitempty"`
	ExpectedInvariants []string `json:"expected_invariants,omitempty"`
	ForbiddenOutcomes  []string `json:"forbidden_outcomes,omitempty"`
	Graders            []Grader `json:"graders,omitempty"`
	SemanticRubric     []string `json:"semantic_rubric,omitempty"`
}

// Grader defines a deterministic check. SemanticRubric is deliberately kept
// separate so reports can label same-platform model judgment.
type Grader struct {
	ID        string   `json:"id"`
	Kind      string   `json:"kind"`
	Target    string   `json:"target,omitempty"`
	Required  []string `json:"required,omitempty"`
	Forbidden []string `json:"forbidden,omitempty"`
	Weight    float64  `json:"weight"`
}

// ClaimLedger is the human-adjudicated primary-evidence claim corpus.
type ClaimLedger struct {
	SchemaVersion int     `json:"schema_version"`
	VerifiedOn    string  `json:"verified_on"`
	Claims        []Claim `json:"claims"`
}

// Claim is the subset needed for ownership and source validation. Strict JSON
// loading keeps this synchronized with knowledge/claims/schema.json.
type Claim struct {
	ID              string          `json:"id"`
	Statement       string          `json:"statement"`
	Status          string          `json:"status"`
	Scope           string          `json:"scope"`
	Invariant       string          `json:"invariant"`
	Qualifications  []string        `json:"qualifications"`
	Counterexamples []string        `json:"counterexamples"`
	GoVersions      []string        `json:"go_versions"`
	Owners          []string        `json:"owners"`
	RiskDomains     []string        `json:"risk_domains"`
	PrimaryEvidence []ClaimEvidence `json:"primary_evidence"`
	ObservedIn      []string        `json:"observed_in"`
}

// ClaimEvidence is a primary or normative source attached to a claim.
type ClaimEvidence struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	Publisher  string `json:"publisher"`
	Kind       string `json:"kind"`
	VerifiedOn string `json:"verified_on"`
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
	Claims   ClaimLedger
}

// CollectionMetrics records discovery usage independently for each installable
// plugin.
type CollectionMetrics struct {
	Name                string `json:"name"`
	SkillCount          int    `json:"skill_count"`
	DiscoveryCharacters int    `json:"discovery_characters"`
}

// Metrics summarizes corpus size, evidence, routing, and discovery metadata.
type Metrics struct {
	SkillCount            int                 `json:"skill_count"`
	DiscoveryCharacters   int                 `json:"discovery_characters"`
	Collections           []CollectionMetrics `json:"collections"`
	SkillLines            int                 `json:"skill_lines"`
	SkillWords            int                 `json:"skill_words"`
	SourceCount           int                 `json:"source_count"`
	ClaimCount            int                 `json:"claim_count"`
	EvaluationCount       int                 `json:"evaluation_count"`
	MaxDescriptionOverlap float64             `json:"max_description_overlap"`
}
