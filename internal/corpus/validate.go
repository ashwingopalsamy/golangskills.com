package corpus

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ashwingopalsamy/golangskills.com/internal/research"
)

const (
	maxDescriptionCharacters = 500
	maxDiscoveryCharacters   = 7800
	maxCollectionDiscovery   = 4000
	minLocationCharacters    = 160
	maxSkillLines            = 250
	maxSkillWords            = 1800
	maxSourceAge             = 400 * 24 * time.Hour
	maxDescriptionOverlap    = 0.68
)

var collectionNames = []string{
	"distributed-systems-skills-for-go",
	"engineering-skills-for-go",
	"fintech-skills-for-go",
}

var (
	skillNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	semverPattern    = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)
	goVersionPattern = regexp.MustCompile(`^1\.[0-9]+$`)
	datePattern      = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
	markdownLink     = regexp.MustCompile(`\[[^]]*\]\(([^)]+)\)`)
)

// ValidationError reports all corpus violations in deterministic order.
type ValidationError struct {
	Issues []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Issues, "\n")
}

// Validate checks open Agent Skills constraints and repository policy.
func Validate(collection Collection) (Metrics, error) {
	return validateAt(collection, time.Now().UTC())
}

func validateAt(collection Collection, now time.Time) (Metrics, error) {
	var issues []string
	metrics := Metrics{
		SkillCount:          len(collection.Skills),
		DiscoveryCharacters: len("<available_skills>\n</available_skills>\n"),
		ClaimCount:          len(collection.Claims.Claims),
	}
	names := make(map[string]struct{}, len(collection.Skills))
	displayNames := make(map[string]string, len(collection.Skills))
	claimIDs := make(map[string]struct{}, len(collection.Claims.Claims))
	collectionDiscovery := make(map[string]int, len(collectionNames))
	collectionSkillCount := make(map[string]int, len(collectionNames))

	for _, skill := range collection.Skills {
		names[skill.Name] = struct{}{}
		if previous, exists := displayNames[skill.Metadata.DisplayName]; exists {
			issues = append(issues, fmt.Sprintf("%s: display_name duplicates %s", skill.Name, previous))
		}
		displayNames[skill.Metadata.DisplayName] = skill.Name
	}
	issues = append(issues, validateClaims(collection.Claims, names, claimIDs, now)...)
	issues = append(issues, validateReferenceCoverage(collection.RepoRoot, claimIDs, now)...)

	for _, skill := range collection.Skills {
		prefix := skill.Name + ": "
		issues = append(issues, validateFrontmatter(prefix, skill)...)
		issues = append(issues, validateMetadata(prefix, skill, names, claimIDs, now)...)
		issues = append(issues, validateEvaluations(prefix, skill, collection.RepoRoot, names)...)
		issues = append(issues, validateFiles(prefix, skill)...)

		skillMarkdown := skill.Files["SKILL.md"]
		discovery := discoveryCharacters(skill)
		metrics.DiscoveryCharacters += discovery
		collectionDiscovery[skill.Metadata.Collection] += discovery
		collectionSkillCount[skill.Metadata.Collection]++
		metrics.SkillLines += lineCount(skillMarkdown)
		metrics.SkillWords += len(strings.Fields(string(skillMarkdown)))
		metrics.SourceCount += len(skill.Metadata.Sources)
		metrics.EvaluationCount += len(skill.Evaluations.Cases)
	}
	issues = append(issues, validateSkillClaimEvidence(collection.Skills, collection.Claims.Claims)...)
	for _, name := range collectionNames {
		characters := collectionDiscovery[name] + len("<available_skills>\n</available_skills>\n")
		metrics.Collections = append(metrics.Collections, CollectionMetrics{
			Name: name, SkillCount: collectionSkillCount[name], DiscoveryCharacters: characters,
		})
		if characters > maxCollectionDiscovery {
			issues = append(issues, fmt.Sprintf("collection %s: discovery metadata has %d characters; limit is %d", name, characters, maxCollectionDiscovery))
		}
	}

	if metrics.DiscoveryCharacters > maxDiscoveryCharacters {
		issues = append(issues, fmt.Sprintf(
			"collection: discovery metadata has %d characters; limit is %d",
			metrics.DiscoveryCharacters,
			maxDiscoveryCharacters,
		))
	}
	metrics.MaxDescriptionOverlap, issues = validateDescriptionOverlap(collection.Skills, issues)
	issues = append(issues, validateRepositoryLinks(collection.RepoRoot)...)

	sort.Strings(issues)
	if len(issues) > 0 {
		return metrics, &ValidationError{Issues: issues}
	}
	return metrics, nil
}

func validateSkillClaimEvidence(skills []Skill, claims []Claim) []string {
	claimByID := make(map[string]Claim, len(claims))
	for _, claim := range claims {
		claimByID[claim.ID] = claim
	}

	var issues []string
	for _, skill := range skills {
		sourceURLs := make(map[string]struct{}, len(skill.Metadata.Sources))
		for _, source := range skill.Metadata.Sources {
			sourceURLs[source.URL] = struct{}{}
		}
		for _, claimID := range skill.Metadata.ClaimIDs {
			claim, exists := claimByID[claimID]
			if !exists || !containsString(claim.Owners, skill.Name) || len(claim.PrimaryEvidence) == 0 {
				continue
			}
			covered := false
			for _, evidence := range claim.PrimaryEvidence {
				if _, exists := sourceURLs[evidence.URL]; exists {
					covered = true
					break
				}
			}
			if !covered {
				issues = append(issues, fmt.Sprintf("%s: owned claim %q has no matching primary evidence in skill.json sources", skill.Name, claimID))
			}
		}
	}
	return issues
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func validateReferenceCoverage(repoRoot string, claimIDs map[string]struct{}, now time.Time) []string {
	if repoRoot == "" {
		return nil
	}
	var issues []string
	lockContent, err := os.ReadFile(filepath.Join(repoRoot, "research", "corpus-lock.json"))
	if err != nil {
		return []string{"research: corpus lock: " + err.Error()}
	}
	var lock research.CorpusLock
	if err := decodeStrict(lockContent, &lock); err != nil {
		return []string{"research: decode corpus lock: " + err.Error()}
	}
	dispositionContent, err := os.ReadFile(filepath.Join(repoRoot, "knowledge", "claims", "reference-dispositions.json"))
	if err != nil {
		return []string{"research: disposition index: " + err.Error()}
	}
	var index research.DispositionIndex
	if err := decodeStrict(dispositionContent, &index); err != nil {
		return []string{"research: decode disposition index: " + err.Error()}
	}
	if lock.SchemaVersion != 1 || index.SchemaVersion != 1 {
		issues = append(issues, "research: corpus lock and disposition schemas must be version 1")
	}
	issues = append(issues, validateDate("research corpus: ", lock.VerifiedOn, now)...)
	if index.CorpusSHA256 != research.LockSHA256(lock) {
		issues = append(issues, "research: disposition index does not match corpus lock")
	}
	want := make(map[string]struct{})
	for _, repository := range lock.Repositories {
		for _, item := range repository.MaterialItems {
			key := repository.Name + "/" + item.ID
			if _, duplicate := want[key]; duplicate {
				issues = append(issues, "research: duplicate material item "+key)
			}
			want[key] = struct{}{}
		}
	}
	seen := make(map[string]struct{})
	for _, disposition := range index.DispositionRows {
		key := disposition.Repository + "/" + disposition.ItemID
		if _, exists := want[key]; !exists {
			issues = append(issues, "research: disposition has no material item "+key)
		}
		if _, duplicate := seen[key]; duplicate {
			issues = append(issues, "research: duplicate disposition "+key)
		}
		seen[key] = struct{}{}
		if _, exists := claimIDs[disposition.ClaimID]; !exists {
			issues = append(issues, fmt.Sprintf("research: disposition %s references missing claim %q", key, disposition.ClaimID))
		}
	}
	if len(want) != len(seen) || index.MaterialItems != len(seen) {
		issues = append(issues, fmt.Sprintf("research: material coverage mismatch: %d items, %d dispositions", len(want), len(seen)))
	}
	return issues
}

func validateFrontmatter(prefix string, skill Skill) []string {
	var issues []string
	frontmatter := skill.Frontmatter
	if !skillNamePattern.MatchString(skill.Name) || len(skill.Name) > 64 {
		issues = append(issues, prefix+"directory name must be lowercase kebab-case and at most 64 characters")
	}
	if frontmatter.Name != skill.Name {
		issues = append(issues, prefix+"frontmatter name must match its directory")
	}
	if len(frontmatter.Description) == 0 || len(frontmatter.Description) > maxDescriptionCharacters {
		issues = append(issues, fmt.Sprintf("%sdescription must contain 1-%d characters", prefix, maxDescriptionCharacters))
	}
	if !strings.Contains(frontmatter.Description, "Use ") {
		issues = append(issues, prefix+"description must state when to use the skill")
	}
	if !strings.Contains(frontmatter.Description, "Do not use ") {
		issues = append(issues, prefix+"description must state when not to use the skill")
	}
	if frontmatter.License != "Apache-2.0" {
		issues = append(issues, prefix+"license must be Apache-2.0")
	}
	if len(frontmatter.Compatibility) == 0 || len(frontmatter.Compatibility) > 500 {
		issues = append(issues, prefix+"compatibility must contain 1-500 characters")
	}

	markdown := skill.Files["SKILL.md"]
	if lines := lineCount(markdown); lines > maxSkillLines {
		issues = append(issues, fmt.Sprintf("%sSKILL.md has %d lines; limit is %d", prefix, lines, maxSkillLines))
	}
	if words := len(strings.Fields(string(markdown))); words > maxSkillWords {
		issues = append(issues, fmt.Sprintf("%sSKILL.md has %d words; limit is %d", prefix, words, maxSkillWords))
	}
	return issues
}

func validateMetadata(prefix string, skill Skill, names, claimIDs map[string]struct{}, now time.Time) []string {
	metadata := skill.Metadata
	var issues []string
	if metadata.SchemaVersion != 2 {
		issues = append(issues, prefix+"skill.json schema_version must be 2")
	}
	if !oneOf(metadata.Collection, collectionNames...) {
		issues = append(issues, prefix+"collection is not an installable collection")
	}
	if len(metadata.DisplayName) < 3 || len(metadata.DisplayName) > 64 {
		issues = append(issues, prefix+"display_name must contain 3-64 characters")
	}
	if len(metadata.ShortDescription) < 20 || len(metadata.ShortDescription) > 64 {
		issues = append(issues, prefix+"short_description must contain 20-64 characters")
	}
	if !strings.Contains(metadata.DefaultPrompt, "$"+skill.Name) {
		issues = append(issues, prefix+"default_prompt must explicitly invoke the skill")
	}
	if len(metadata.DefaultPrompt) > 200 {
		issues = append(issues, prefix+"default_prompt must not exceed 200 characters")
	}
	if !semverPattern.MatchString(metadata.Version) {
		issues = append(issues, prefix+"version must be semantic versioning")
	}
	if !oneOf(metadata.Maturity, "experimental", "beta", "stable") {
		issues = append(issues, prefix+"maturity must be experimental, beta, or stable")
	}
	if !oneOf(metadata.Category, "language", "design", "boundary", "verification", "performance", "security", "operations", "review", "concurrency", "consistency", "messaging", "resilience", "coordination", "money", "payments", "idempotency", "settlement", "compliance") {
		issues = append(issues, prefix+"category is not part of the repository taxonomy")
	}
	issues = append(issues, validateStringSet(prefix+"risk_domains", metadata.RiskDomains, 1, 8)...)
	issues = append(issues, validateTags(prefix, metadata.Tags)...)

	minimumMinor, minimumValid := parseGoMinor(metadata.GoVersions.Minimum)
	if !minimumValid {
		issues = append(issues, prefix+"go_versions.minimum must use 1.N form")
	}
	seenVersions := make(map[string]struct{})
	for _, version := range metadata.GoVersions.Guidance {
		minor, valid := parseGoMinor(version)
		if !valid {
			issues = append(issues, fmt.Sprintf("%sunsupported guidance version %q", prefix, version))
		} else if minimumValid && minor < minimumMinor {
			issues = append(issues, fmt.Sprintf("%sguidance version %q predates the minimum", prefix, version))
		}
		if _, duplicate := seenVersions[version]; duplicate {
			issues = append(issues, fmt.Sprintf("%sduplicate guidance version %q", prefix, version))
		}
		seenVersions[version] = struct{}{}
	}
	for _, required := range []string{"1.25", "1.26"} {
		if _, exists := seenVersions[required]; !exists {
			issues = append(issues, fmt.Sprintf("%sguidance must include current target %s", prefix, required))
		}
	}
	if metadata.GoVersions.Minimum != "1.24" {
		issues = append(issues, prefix+"go_versions.minimum must remain 1.24 for repository tooling compatibility")
	}
	for _, legacy := range metadata.GoVersions.Legacy {
		minor, valid := parseGoMinor(legacy)
		if !valid || minor >= 25 {
			issues = append(issues, fmt.Sprintf("%slegacy version %q must be valid and older than 1.25", prefix, legacy))
		}
	}

	issues = append(issues, validateRelations(prefix, skill.Name, metadata.Relations, names)...)
	if len(metadata.ClaimIDs) == 0 {
		issues = append(issues, prefix+"claim_ids must identify at least one canonical claim")
	}
	for _, claimID := range metadata.ClaimIDs {
		if _, exists := claimIDs[claimID]; !exists {
			issues = append(issues, fmt.Sprintf("%sclaim_id %q does not exist", prefix, claimID))
		}
	}
	issues = append(issues, validateCompatibility(prefix, metadata.CompatibilityEvidence, now)...)
	if metadata.SourceProvenance.Method != "independent-rewrite-from-primary-evidence" {
		issues = append(issues, prefix+"source_provenance.method must require an independent primary-evidence rewrite")
	}
	if metadata.SourceProvenance.CorpusLock != "research/corpus-lock.json" {
		issues = append(issues, prefix+"source_provenance.corpus_lock must name research/corpus-lock.json")
	}
	if len(metadata.Sources) < 2 {
		issues = append(issues, prefix+"at least two evidence sources are required")
	}
	seenSources := make(map[string]struct{}, len(metadata.Sources))
	for index, source := range metadata.Sources {
		sourcePrefix := fmt.Sprintf("%ssource %d: ", prefix, index+1)
		issues = append(issues, validateSource(sourcePrefix, source, now)...)
		if _, duplicate := seenSources[source.URL]; duplicate {
			issues = append(issues, sourcePrefix+"URL is duplicated")
		}
		seenSources[source.URL] = struct{}{}
	}
	return issues
}

func validateTags(prefix string, tags []string) []string {
	var issues []string
	if len(tags) < 3 || len(tags) > 10 {
		issues = append(issues, prefix+"tags must contain 3-10 entries")
	}
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		if !skillNamePattern.MatchString(tag) {
			issues = append(issues, fmt.Sprintf("%sinvalid tag %q", prefix, tag))
		}
		if _, duplicate := seen[tag]; duplicate {
			issues = append(issues, fmt.Sprintf("%sduplicate tag %q", prefix, tag))
		}
		seen[tag] = struct{}{}
	}
	return issues
}

func validateStringSet(label string, values []string, minimum, maximum int) []string {
	var issues []string
	if len(values) < minimum || len(values) > maximum {
		issues = append(issues, fmt.Sprintf("%s must contain %d-%d entries", label, minimum, maximum))
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !skillNamePattern.MatchString(value) {
			issues = append(issues, fmt.Sprintf("%s contains invalid value %q", label, value))
		}
		if _, duplicate := seen[value]; duplicate {
			issues = append(issues, fmt.Sprintf("%s contains duplicate value %q", label, value))
		}
		seen[value] = struct{}{}
	}
	return issues
}

func validateCompatibility(prefix string, evidence []CompatibilityEvidence, now time.Time) []string {
	var issues []string
	clients := make(map[string]struct{}, len(evidence))
	for index, item := range evidence {
		itemPrefix := fmt.Sprintf("%scompatibility evidence %d: ", prefix, index+1)
		if !oneOf(item.Client, "codex", "claude-code", "cursor", "opencode") {
			issues = append(issues, itemPrefix+"client is not supported")
		}
		if !oneOf(item.Level, "behaviorally-benchmarked", "structurally-compatible") {
			issues = append(issues, itemPrefix+"level is not recognized")
		}
		if item.Contract == "" {
			issues = append(issues, itemPrefix+"contract is required")
		}
		if _, duplicate := clients[item.Client]; duplicate {
			issues = append(issues, itemPrefix+"client is duplicated")
		}
		clients[item.Client] = struct{}{}
		issues = append(issues, validateDate(itemPrefix, item.VerifiedOn, now)...)
	}
	for _, client := range []string{"codex", "claude-code", "cursor", "opencode"} {
		if _, exists := clients[client]; !exists {
			issues = append(issues, prefix+"compatibility_evidence must cover "+client)
		}
	}
	return issues
}

func validateRelations(prefix, name string, relations Relations, names map[string]struct{}) []string {
	var issues []string
	seen := make(map[string]struct{})
	for _, relation := range append(append([]string{}, relations.Complements...), relations.Overlaps...) {
		if relation == name {
			issues = append(issues, prefix+"a skill cannot relate to itself")
		}
		if _, exists := names[relation]; !exists {
			issues = append(issues, fmt.Sprintf("%srelation %q does not exist", prefix, relation))
		}
		if _, duplicate := seen[relation]; duplicate {
			issues = append(issues, fmt.Sprintf("%srelation %q is duplicated", prefix, relation))
		}
		seen[relation] = struct{}{}
	}
	return issues
}

func validateSource(prefix string, source Source, now time.Time) []string {
	var issues []string
	if source.Title == "" || source.Publisher == "" {
		issues = append(issues, prefix+"title and publisher are required")
	}
	parsedURL, err := url.Parse(source.URL)
	if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" {
		issues = append(issues, prefix+"URL must be absolute HTTPS")
	}
	if !oneOf(source.Kind, "normative", "primary", "operational", "organization-specific") {
		issues = append(issues, prefix+"kind is not recognized")
	}
	if len(source.Supports) == 0 {
		issues = append(issues, prefix+"supports must identify at least one claim")
	}
	issues = append(issues, validateDate(prefix, source.VerifiedOn, now)...)
	return issues
}

func validateDate(prefix, value string, now time.Time) []string {
	var issues []string
	if !datePattern.MatchString(value) {
		return append(issues, prefix+"verified_on must use YYYY-MM-DD")
	}
	verified, err := time.Parse("2006-01-02", value)
	if err != nil {
		return append(issues, prefix+"verified_on is not a calendar date")
	}
	now = now.Truncate(24 * time.Hour)
	if verified.After(now) {
		issues = append(issues, prefix+"verified_on is in the future")
	} else if now.Sub(verified) > maxSourceAge {
		issues = append(issues, prefix+"source verification is older than 400 days")
	}
	return issues
}

func validateClaims(ledger ClaimLedger, skillNames, claimIDs map[string]struct{}, now time.Time) []string {
	var issues []string
	if ledger.SchemaVersion != 2 {
		issues = append(issues, "claims: schema_version must be 2")
	}
	issues = append(issues, validateDate("claims: ", ledger.VerifiedOn, now)...)
	for index, claim := range ledger.Claims {
		prefix := fmt.Sprintf("claim %d (%s): ", index+1, claim.ID)
		if !skillNamePattern.MatchString(claim.ID) {
			issues = append(issues, prefix+"id must be lowercase kebab-case")
		}
		if _, duplicate := claimIDs[claim.ID]; duplicate {
			issues = append(issues, prefix+"id is duplicated")
		}
		claimIDs[claim.ID] = struct{}{}
		if !oneOf(claim.Status, "adopted", "adopted-with-qualifications", "organizational-preference", "version-specific", "rejected", "outside-project-scope") {
			issues = append(issues, prefix+"status is not recognized")
		}
		if claim.Statement == "" || claim.Scope == "" || claim.Invariant == "" {
			issues = append(issues, prefix+"statement, scope, and invariant are required")
		}
		if len(claim.Owners) == 0 {
			issues = append(issues, prefix+"at least one owner is required")
		}
		for _, owner := range claim.Owners {
			if _, exists := skillNames[owner]; !exists {
				issues = append(issues, fmt.Sprintf("%sowner %q does not exist", prefix, owner))
			}
		}
		sensitive := strings.HasPrefix(claim.ID, "fin-") || strings.Contains(claim.ID, "security") || claim.Status == "adopted" || claim.Status == "version-specific"
		if sensitive && len(claim.PrimaryEvidence) == 0 {
			issues = append(issues, prefix+"normative, security, and financial claims require primary evidence")
		}
		for evidenceIndex, evidence := range claim.PrimaryEvidence {
			evidencePrefix := fmt.Sprintf("%sevidence %d: ", prefix, evidenceIndex+1)
			parsedURL, err := url.Parse(evidence.URL)
			if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" {
				issues = append(issues, evidencePrefix+"URL must be absolute HTTPS")
			}
			if evidence.Title == "" || evidence.Publisher == "" || !oneOf(evidence.Kind, "normative", "primary", "protocol", "operational", "regulatory") {
				issues = append(issues, evidencePrefix+"title, publisher, and recognized kind are required")
			}
			issues = append(issues, validateDate(evidencePrefix, evidence.VerifiedOn, now)...)
		}
	}
	return issues
}

func validateEvaluations(prefix string, skill Skill, repoRoot string, names map[string]struct{}) []string {
	evaluations := skill.Evaluations
	var issues []string
	if evaluations.SchemaVersion != 2 {
		issues = append(issues, prefix+"evals.json schema_version must be 2")
	}
	if evaluations.Skill != skill.Name {
		issues = append(issues, prefix+"evals.json skill must match its directory")
	}

	seen := make(map[string]struct{}, len(evaluations.Cases))
	positiveRoutes, negativeRoutes, qualityCases := 0, 0, 0
	for index, eval := range evaluations.Cases {
		evalPrefix := fmt.Sprintf("%seval %d: ", prefix, index+1)
		if !skillNamePattern.MatchString(eval.ID) {
			issues = append(issues, evalPrefix+"id must be lowercase kebab-case")
		}
		if _, duplicate := seen[eval.ID]; duplicate {
			issues = append(issues, evalPrefix+"id is duplicated")
		}
		seen[eval.ID] = struct{}{}
		if strings.TrimSpace(eval.Prompt) == "" {
			issues = append(issues, evalPrefix+"prompt is required")
		}
		if !oneOf(eval.Split, "development", "holdout") {
			issues = append(issues, evalPrefix+"split must be development or holdout")
		}
		switch eval.Kind {
		case "routing":
			if eval.ShouldActivate == nil {
				issues = append(issues, evalPrefix+"routing cases require should_activate")
			} else if *eval.ShouldActivate {
				positiveRoutes++
			} else {
				negativeRoutes++
			}
			if eval.Reason == "" {
				issues = append(issues, evalPrefix+"routing cases require a reason")
			}
			for _, alternative := range eval.ConfusesWith {
				if _, exists := names[alternative]; !exists || alternative == skill.Name {
					issues = append(issues, fmt.Sprintf("%sconfusion route %q must name another canonical skill", evalPrefix, alternative))
				}
			}
			if eval.Fixture != "" || len(eval.ExpectedInvariants) != 0 || len(eval.ForbiddenOutcomes) != 0 || len(eval.Graders) != 0 || len(eval.SemanticRubric) != 0 {
				issues = append(issues, evalPrefix+"routing cases cannot contain quality criteria")
			}
		case "quality":
			qualityCases++
			if eval.ShouldActivate != nil || eval.Reason != "" {
				issues = append(issues, evalPrefix+"quality cases cannot contain routing fields")
			}
			if len(eval.ExpectedInvariants) < 2 || len(eval.ForbiddenOutcomes) < 1 {
				issues = append(issues, evalPrefix+"quality cases require at least two expected invariants and one forbidden outcome")
			}
			if len(eval.Graders) == 0 {
				issues = append(issues, evalPrefix+"quality cases require at least one deterministic grader")
			}
			hasGoTest := false
			for graderIndex, grader := range eval.Graders {
				graderPrefix := fmt.Sprintf("%sgrader %d: ", evalPrefix, graderIndex+1)
				if !skillNamePattern.MatchString(grader.ID) || !oneOf(grader.Kind, "contains", "not-contains", "go-test", "json") {
					issues = append(issues, graderPrefix+"id or kind is invalid")
				}
				if grader.Kind == "go-test" {
					hasGoTest = true
					if eval.Fixture == "" {
						issues = append(issues, graderPrefix+"go-test requires a fixture")
					}
					if grader.Target != "" && grader.Target != "./..." && grader.Target != "-race ./..." {
						issues = append(issues, graderPrefix+"go-test target must be empty, ./..., or -race ./...")
					}
				}
				if grader.Weight <= 0 {
					issues = append(issues, graderPrefix+"weight must be positive")
				}
			}
			if eval.Fixture != "" {
				clean := path.Clean(eval.Fixture)
				if !strings.HasPrefix(clean, "evaluations/fixtures/") || clean != eval.Fixture {
					issues = append(issues, evalPrefix+"fixture must be a clean path under evaluations/fixtures")
				} else if repoRoot != "" {
					info, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(clean)))
					if err != nil || !info.IsDir() {
						issues = append(issues, evalPrefix+"fixture directory does not exist")
					}
				}
				if !hasGoTest {
					issues = append(issues, evalPrefix+"fixture requires at least one go-test grader")
				}
			}
		default:
			issues = append(issues, evalPrefix+"kind must be routing or quality")
		}
	}
	if positiveRoutes < 1 || negativeRoutes < 1 || qualityCases < 2 {
		issues = append(issues, prefix+"evals require at least one positive route, one negative route, and two quality cases")
	}
	return issues
}

func validateFiles(prefix string, skill Skill) []string {
	var issues []string
	references := make(map[string]struct{})
	for filename := range skill.Files {
		switch {
		case filename == "SKILL.md", filename == "skill.json", filename == "evals.json":
		case filename == "agents/openai.yaml":
		case strings.HasPrefix(filename, "references/") && path.Ext(filename) == ".md" && strings.Count(filename, "/") == 1:
			references[filename] = struct{}{}
		default:
			issues = append(issues, fmt.Sprintf("%sunexpected file %s", prefix, filename))
		}
		if strings.HasPrefix(filename, "scripts/") || strings.Contains(filename, "/scripts/") {
			issues = append(issues, fmt.Sprintf("%sexecutable skill scripts are not accepted: %s", prefix, filename))
		}
	}

	linkedReferences := make(map[string]struct{})
	for filename, content := range skill.Files {
		if path.Ext(filename) != ".md" {
			continue
		}
		for _, match := range markdownLink.FindAllStringSubmatch(string(content), -1) {
			target := strings.Fields(match[1])[0]
			target = strings.Trim(target, "<>")
			if strings.HasPrefix(target, "#") {
				continue
			}
			if parsed, err := url.Parse(target); err == nil && parsed.Scheme != "" {
				continue
			}
			target = strings.SplitN(target, "#", 2)[0]
			resolved := path.Clean(path.Join(path.Dir(filename), target))
			if resolved == ".." || strings.HasPrefix(resolved, "../") {
				issues = append(issues, fmt.Sprintf("%s%s link escapes the skill directory: %s", prefix, filename, target))
				continue
			}
			if _, exists := skill.Files[resolved]; !exists {
				issues = append(issues, fmt.Sprintf("%s%s has missing link target %s", prefix, filename, target))
			}
			if filename == "SKILL.md" && strings.HasPrefix(resolved, "references/") {
				linkedReferences[resolved] = struct{}{}
			}
		}
	}
	for reference := range references {
		if _, linked := linkedReferences[reference]; !linked {
			issues = append(issues, fmt.Sprintf("%s%s is not linked from SKILL.md", prefix, reference))
		}
	}
	return issues
}

func validateRepositoryLinks(repoRoot string) []string {
	if repoRoot == "" {
		return nil
	}
	var issues []string
	err := filepath.WalkDir(repoRoot, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "plugins", "skills":
				if filename != repoRoot {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if filepath.Ext(filename) != ".md" {
			return nil
		}
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(repoRoot, filename)
		if err != nil {
			return err
		}
		for _, match := range markdownLink.FindAllStringSubmatch(string(content), -1) {
			target := strings.Trim(strings.Fields(match[1])[0], "<>")
			parsed, err := url.Parse(target)
			if err != nil || parsed.Scheme != "" || strings.HasPrefix(target, "#") {
				continue
			}
			linkPath, err := url.PathUnescape(parsed.Path)
			if err != nil || linkPath == "" {
				continue
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(filename), filepath.FromSlash(linkPath)))
			withinRepo, err := filepath.Rel(repoRoot, resolved)
			if err != nil || withinRepo == ".." || strings.HasPrefix(withinRepo, ".."+string(filepath.Separator)) {
				issues = append(issues, fmt.Sprintf("%s link escapes the repository: %s", filepath.ToSlash(relative), target))
				continue
			}
			if _, err := os.Stat(resolved); err != nil {
				issues = append(issues, fmt.Sprintf("%s has missing link target %s", filepath.ToSlash(relative), target))
			}
		}
		return nil
	})
	if err != nil {
		issues = append(issues, fmt.Sprintf("repository links: %v", err))
	}
	return issues
}

func validateDescriptionOverlap(skills []Skill, issues []string) (float64, []string) {
	maximum := 0.0
	for left := 0; left < len(skills); left++ {
		for right := left + 1; right < len(skills); right++ {
			overlap := jaccard(descriptionTerms(skills[left].Frontmatter.Description), descriptionTerms(skills[right].Frontmatter.Description))
			if overlap > maximum {
				maximum = overlap
			}
			if overlap > maxDescriptionOverlap {
				issues = append(issues, fmt.Sprintf(
					"collection: descriptions for %s and %s overlap %.2f; limit is %.2f",
					skills[left].Name,
					skills[right].Name,
					overlap,
					maxDescriptionOverlap,
				))
			}
		}
	}
	return maximum, issues
}

func descriptionTerms(description string) map[string]struct{} {
	stopWords := map[string]struct{}{
		"a": {}, "an": {}, "and": {}, "by": {}, "do": {}, "for": {}, "from": {}, "in": {},
		"not": {}, "of": {}, "on": {}, "or": {}, "the": {}, "to": {}, "use": {}, "when": {}, "with": {},
	}
	terms := make(map[string]struct{})
	for _, term := range strings.FieldsFunc(strings.ToLower(description), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-'
	}) {
		if len(term) < 3 {
			continue
		}
		if _, stop := stopWords[term]; !stop {
			terms[term] = struct{}{}
		}
	}
	return terms
}

func discoveryCharacters(skill Skill) int {
	locationCharacters := len(skill.Dir) + len("/SKILL.md")
	if locationCharacters < minLocationCharacters {
		locationCharacters = minLocationCharacters
	}
	return len("<skill>\n<name>\n</name>\n<description>\n</description>\n<location>\n</location>\n</skill>\n") +
		len(skill.Name) + len(skill.Frontmatter.Description) + locationCharacters
}

func jaccard(left, right map[string]struct{}) float64 {
	if len(left) == 0 && len(right) == 0 {
		return 1
	}
	intersection := 0
	for term := range left {
		if _, exists := right[term]; exists {
			intersection++
		}
	}
	union := len(left) + len(right) - intersection
	return float64(intersection) / float64(union)
}

func lineCount(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := bytes.Count(content, []byte{'\n'})
	if content[len(content)-1] != '\n' {
		lines++
	}
	return lines
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func parseGoMinor(version string) (int, bool) {
	if !goVersionPattern.MatchString(version) {
		return 0, false
	}
	minor, err := strconv.Atoi(strings.TrimPrefix(version, "1."))
	return minor, err == nil
}
