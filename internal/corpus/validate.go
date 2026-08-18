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
)

const (
	maxDescriptionCharacters = 500
	maxDiscoveryCharacters   = 6000
	minLocationCharacters    = 160
	maxSkillLines            = 300
	maxSkillWords            = 3000
	maxSourceAge             = 400 * 24 * time.Hour
	maxDescriptionOverlap    = 0.68
)

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
	}
	names := make(map[string]struct{}, len(collection.Skills))
	displayNames := make(map[string]string, len(collection.Skills))

	for _, skill := range collection.Skills {
		names[skill.Name] = struct{}{}
		if previous, exists := displayNames[skill.Metadata.DisplayName]; exists {
			issues = append(issues, fmt.Sprintf("%s: display_name duplicates %s", skill.Name, previous))
		}
		displayNames[skill.Metadata.DisplayName] = skill.Name
	}

	for _, skill := range collection.Skills {
		prefix := skill.Name + ": "
		issues = append(issues, validateFrontmatter(prefix, skill)...)
		issues = append(issues, validateMetadata(prefix, skill, names, now)...)
		issues = append(issues, validateEvaluations(prefix, skill)...)
		issues = append(issues, validateFiles(prefix, skill)...)

		skillMarkdown := skill.Files["SKILL.md"]
		metrics.DiscoveryCharacters += discoveryCharacters(skill)
		metrics.SkillLines += lineCount(skillMarkdown)
		metrics.SkillWords += len(strings.Fields(string(skillMarkdown)))
		metrics.SourceCount += len(skill.Metadata.Sources)
		metrics.EvaluationCount += len(skill.Evaluations.Cases)
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

func validateMetadata(prefix string, skill Skill, names map[string]struct{}, now time.Time) []string {
	metadata := skill.Metadata
	var issues []string
	if metadata.SchemaVersion != 1 {
		issues = append(issues, prefix+"skill.json schema_version must be 1")
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
	if !oneOf(metadata.Status, "experimental", "beta", "stable") {
		issues = append(issues, prefix+"status must be experimental, beta, or stable")
	}
	if !oneOf(metadata.Category, "execution", "state", "reliability", "workflow") {
		issues = append(issues, prefix+"category is not part of the repository taxonomy")
	}
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

	issues = append(issues, validateRelations(prefix, skill.Name, metadata.Relations, names)...)
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
	if !datePattern.MatchString(source.VerifiedOn) {
		issues = append(issues, prefix+"verified_on must use YYYY-MM-DD")
		return issues
	}
	verified, err := time.Parse("2006-01-02", source.VerifiedOn)
	if err != nil {
		issues = append(issues, prefix+"verified_on is not a calendar date")
		return issues
	}
	now = now.Truncate(24 * time.Hour)
	if verified.After(now) {
		issues = append(issues, prefix+"verified_on is in the future")
	} else if now.Sub(verified) > maxSourceAge {
		issues = append(issues, prefix+"source verification is older than 400 days")
	}
	return issues
}

func validateEvaluations(prefix string, skill Skill) []string {
	evaluations := skill.Evaluations
	var issues []string
	if evaluations.SchemaVersion != 1 {
		issues = append(issues, prefix+"evals.json schema_version must be 1")
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
			if len(eval.Criteria) != 0 || len(eval.AntiCriteria) != 0 {
				issues = append(issues, evalPrefix+"routing cases cannot contain quality criteria")
			}
		case "quality":
			qualityCases++
			if eval.ShouldActivate != nil || eval.Reason != "" {
				issues = append(issues, evalPrefix+"quality cases cannot contain routing fields")
			}
			if len(eval.Criteria) < 2 || len(eval.AntiCriteria) < 1 {
				issues = append(issues, evalPrefix+"quality cases require at least two criteria and one anti-criterion")
			}
		default:
			issues = append(issues, evalPrefix+"kind must be routing or quality")
		}
	}
	if positiveRoutes < 2 || negativeRoutes < 2 || qualityCases < 2 {
		issues = append(issues, prefix+"evals require at least two positive routes, two negative routes, and two quality cases")
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
