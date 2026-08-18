// Package research locks local reference repositories without copying their prose.
package research

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const lockSchemaVersion = 1

var (
	materialExtensions = map[string]struct{}{
		".go": {}, ".json": {}, ".md": {}, ".py": {}, ".sh": {}, ".txt": {}, ".yaml": {}, ".yml": {},
	}
	normativePattern  = regexp.MustCompile(`(?i)\b(must|should|prefer|avoid|never|always|recommend(?:ed)?|require[sd]?|ensure|do not|don't)\b`)
	evaluationPattern = regexp.MustCompile(`(?i)\b(assert(?:ion)?|benchmark|eval(?:uation)?|grader|pass rate|score|accuracy|baseline|rubric)\b`)
)

// CorpusLock identifies every local snapshot and every file that contributed to
// the research corpus. MaterialItems store hashes and coordinates, never copied
// competitor prose.
type CorpusLock struct {
	SchemaVersion int          `json:"schema_version"`
	VerifiedOn    string       `json:"verified_on"`
	ReferenceRoot string       `json:"reference_root"`
	Repositories  []Repository `json:"repositories"`
}

// Repository is one locked local reference snapshot.
type Repository struct {
	Name                string         `json:"name"`
	Path                string         `json:"path"`
	Remote              string         `json:"remote,omitempty"`
	Commit              string         `json:"commit,omitempty"`
	CommitDate          string         `json:"commit_date,omitempty"`
	Dirty               bool           `json:"dirty,omitempty"`
	SnapshotSHA256      string         `json:"snapshot_sha256"`
	License             License        `json:"license"`
	ReusePolicy         string         `json:"reuse_policy"`
	BenchmarkEligible   bool           `json:"benchmark_eligible"`
	BenchmarkExclusion  string         `json:"benchmark_exclusion,omitempty"`
	Files               []File         `json:"files"`
	Skills              []Skill        `json:"skills"`
	MaterialItems       []MaterialItem `json:"material_items"`
	MaterialItemCount   int            `json:"material_item_count"`
	CanonicalSkillCount int            `json:"canonical_skill_count"`
}

// License records exact local notice files and the conservative reuse policy.
type License struct {
	SPDX   string   `json:"spdx,omitempty"`
	Status string   `json:"status"`
	Files  []string `json:"files,omitempty"`
}

// File locks a repository-relative regular file.
type File struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// Skill identifies a canonical skill entrypoint found in a reference.
type Skill struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// MaterialItem identifies a recommendation, taught code block, or scoring
// assertion by location and hash. The claim ledger owns its disposition.
type MaterialItem struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Kind      string `json:"kind"`
	SHA256    string `json:"sha256"`
}

// DispositionIndex proves that every material item in a corpus lock has an
// explicit technical or scope disposition. Canonical claims provide the
// detailed evidence and qualifications.
type DispositionIndex struct {
	SchemaVersion   int           `json:"schema_version"`
	CorpusVerified  string        `json:"corpus_verified_on"`
	CorpusSHA256    string        `json:"corpus_sha256"`
	MaterialItems   int           `json:"material_items"`
	DispositionRows []Disposition `json:"dispositions"`
}

// Disposition maps one hashed source item to a canonical claim or a documented
// exclusion. It intentionally contains no copied source text.
type Disposition struct {
	Repository string `json:"repository"`
	ItemID     string `json:"item_id"`
	Status     string `json:"status"`
	ClaimID    string `json:"claim_id"`
	Rationale  string `json:"rationale"`
}

// Audit locks every immediate child snapshot under referenceRoot.
func Audit(referenceRoot, verifiedOn string) (CorpusLock, error) {
	if referenceRoot == "" {
		return CorpusLock{}, errors.New("reference root is required")
	}
	if _, err := time.Parse("2006-01-02", verifiedOn); err != nil {
		return CorpusLock{}, fmt.Errorf("verified-on must use YYYY-MM-DD: %w", err)
	}
	absoluteRoot, err := filepath.Abs(referenceRoot)
	if err != nil {
		return CorpusLock{}, fmt.Errorf("resolve reference root: %w", err)
	}
	entries, err := os.ReadDir(absoluteRoot)
	if err != nil {
		return CorpusLock{}, fmt.Errorf("read reference root: %w", err)
	}

	lock := CorpusLock{
		SchemaVersion: lockSchemaVersion,
		VerifiedOn:    verifiedOn,
		ReferenceRoot: absoluteRoot,
	}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		repository, err := auditRepository(filepath.Join(absoluteRoot, entry.Name()))
		if err != nil {
			return CorpusLock{}, fmt.Errorf("audit %s: %w", entry.Name(), err)
		}
		lock.Repositories = append(lock.Repositories, repository)
	}
	if len(lock.Repositories) == 0 {
		return CorpusLock{}, errors.New("reference root contains no repositories")
	}
	sort.Slice(lock.Repositories, func(i, j int) bool {
		return lock.Repositories[i].Name < lock.Repositories[j].Name
	})
	return lock, nil
}

func auditRepository(root string) (Repository, error) {
	repository := Repository{Name: filepath.Base(root), Path: root}
	if directory(filepath.Join(root, ".git")) {
		repository.Remote, _ = gitOutput(root, "config", "--get", "remote.origin.url")
		repository.Commit, _ = gitOutput(root, "rev-parse", "HEAD")
		repository.CommitDate, _ = gitOutput(root, "show", "-s", "--format=%cI", "HEAD")
		status, statusErr := gitOutput(root, "status", "--porcelain")
		repository.Dirty = statusErr == nil && status != ""
	}

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		repository.Files = append(repository.Files, File{
			Path: relative, Size: info.Size(), SHA256: hash(content),
		})
		if strings.EqualFold(filepath.Base(path), "SKILL.md") {
			repository.Skills = append(repository.Skills, Skill{
				Name: filepath.Base(filepath.Dir(path)), Path: relative,
			})
		}
		if isLicenseFile(relative) {
			repository.License.Files = append(repository.License.Files, relative)
			if repository.License.SPDX == "" {
				repository.License.SPDX = detectLicense(content)
			}
		}
		if _, relevant := materialExtensions[strings.ToLower(filepath.Ext(relative))]; relevant {
			repository.MaterialItems = append(repository.MaterialItems, extractMaterial(relative, content)...)
		}
		return nil
	})
	if err != nil {
		return Repository{}, err
	}
	sort.Slice(repository.Files, func(i, j int) bool { return repository.Files[i].Path < repository.Files[j].Path })
	sort.Slice(repository.Skills, func(i, j int) bool { return repository.Skills[i].Path < repository.Skills[j].Path })
	repository.Skills = canonicalSkills(repository.Skills)
	sort.Slice(repository.MaterialItems, func(i, j int) bool {
		left, right := repository.MaterialItems[i], repository.MaterialItems[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.StartLine != right.StartLine {
			return left.StartLine < right.StartLine
		}
		return left.Kind < right.Kind
	})
	repository.SnapshotSHA256 = snapshotHash(repository.Files)
	repository.CanonicalSkillCount = len(repository.Skills)
	repository.MaterialItemCount = len(repository.MaterialItems)
	if repository.License.SPDX == "" {
		repository.License.Status = "no-local-license-file"
		repository.ReusePolicy = "reference-only"
		repository.BenchmarkEligible = true
	} else {
		repository.License.Status = "local-license-file-verified"
		repository.ReusePolicy = "independent-rewrite-with-attribution"
		repository.BenchmarkEligible = true
	}
	if len(repository.Skills) == 0 {
		repository.BenchmarkEligible = false
		repository.BenchmarkExclusion = "not an installable agent-skill collection"
	}
	return repository, nil
}

func canonicalSkills(skills []Skill) []Skill {
	hasTopLevelSkills := false
	for _, skill := range skills {
		if strings.HasPrefix(skill.Path, "skills/") {
			hasTopLevelSkills = true
			break
		}
	}
	if !hasTopLevelSkills {
		return skills
	}
	canonical := make([]Skill, 0, len(skills))
	for _, skill := range skills {
		if strings.HasPrefix(skill.Path, "skills/") {
			canonical = append(canonical, skill)
		}
	}
	return canonical
}

func extractMaterial(relative string, content []byte) []MaterialItem {
	var items []MaterialItem
	scanner := bufio.NewScanner(bytes.NewReader(content))
	// Large generated lines should be hashed rather than making the audit fail.
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	lineNumber := 0
	inFence := false
	fenceStart := 0
	var fence bytes.Buffer
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inFence {
				fence.WriteString(line)
				items = append(items, materialItem(relative, fenceStart, lineNumber, "taught-code-pattern", fence.Bytes()))
				fence.Reset()
				inFence = false
			} else {
				inFence = true
				fenceStart = lineNumber
				fence.WriteString(line)
				fence.WriteByte('\n')
			}
			continue
		}
		if inFence {
			fence.WriteString(line)
			fence.WriteByte('\n')
			continue
		}
		kind := ""
		switch {
		case evaluationPattern.MatchString(line):
			kind = "scoring-assertion"
		case normativePattern.MatchString(line):
			kind = "normative-recommendation"
		}
		if kind != "" {
			items = append(items, materialItem(relative, lineNumber, lineNumber, kind, []byte(strings.TrimSpace(line))))
		}
	}
	if inFence {
		items = append(items, materialItem(relative, fenceStart, lineNumber, "taught-code-pattern", fence.Bytes()))
	}
	return items
}

func materialItem(relative string, start, end int, kind string, content []byte) MaterialItem {
	digest := hash(content)
	idDigest := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%d:%s:%s", relative, start, end, kind, digest)))
	return MaterialItem{
		ID:        "mi-" + hex.EncodeToString(idDigest[:8]),
		Path:      relative,
		StartLine: start,
		EndLine:   end,
		Kind:      kind,
		SHA256:    digest,
	}
}

func snapshotHash(files []File) string {
	hasher := sha256.New()
	for _, file := range files {
		fmt.Fprintf(hasher, "%s\x00%d\x00%s\n", file.Path, file.Size, file.SHA256)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func hash(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func gitOutput(root string, arguments ...string) (string, error) {
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	content, err := command.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

func isLicenseFile(relative string) bool {
	base := strings.ToLower(filepath.Base(relative))
	return base == "license" || strings.HasPrefix(base, "license.") || base == "copying" || strings.HasPrefix(base, "copying.")
}

func detectLicense(content []byte) string {
	lower := strings.ToLower(string(content))
	switch {
	case strings.Contains(lower, "apache license") && strings.Contains(lower, "version 2.0"):
		return "Apache-2.0"
	case strings.Contains(lower, "permission is hereby granted, free of charge"):
		return "MIT"
	case strings.Contains(lower, "gnu general public license"):
		return "GPL"
	default:
		return "UNKNOWN"
	}
}

func directory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// WriteLock writes a deterministic indented lock file atomically.
func WriteLock(path string, lock CorpusLock) error {
	return writeJSONAtomic(path, lock, ".corpus-lock-*")
}

// BuildDispositionIndex maps every material item to a primary-evidence claim
// family or an explicit exclusion. The domain mapping is deterministic and is
// reviewed alongside the human-authored claim ledger.
func BuildDispositionIndex(lock CorpusLock) DispositionIndex {
	index := DispositionIndex{
		SchemaVersion:  1,
		CorpusVerified: lock.VerifiedOn,
		CorpusSHA256:   corpusLockHash(lock),
	}
	for _, repository := range lock.Repositories {
		for _, item := range repository.MaterialItems {
			status, claimID, rationale := classifyDisposition(item)
			index.DispositionRows = append(index.DispositionRows, Disposition{
				Repository: repository.Name,
				ItemID:     item.ID,
				Status:     status,
				ClaimID:    claimID,
				Rationale:  rationale,
			})
		}
	}
	index.MaterialItems = len(index.DispositionRows)
	return index
}

// WriteDispositionIndex writes the complete item-to-claim coverage map.
func WriteDispositionIndex(path string, index DispositionIndex) error {
	return writeJSONAtomic(path, index, ".reference-dispositions-*")
}

func classifyDisposition(item MaterialItem) (string, string, string) {
	path := strings.ToLower(item.Path)
	mappings := []struct {
		terms   []string
		claimID string
	}{
		{[]string{"concurr", "goroutine", "channel", "context", "race", "worker-pool"}, "dist-bounded-concurrency"},
		{[]string{"money", "currency", "decimal", "rounding"}, "fin-money-representation"},
		{[]string{"ledger", "accounting", "double-entry"}, "fin-ledger-invariants"},
		{[]string{"payment", "refund", "chargeback", "authorization", "capture"}, "fin-payment-state-machine"},
		{[]string{"idempot", "dedup", "replay"}, "fin-idempotency-record"},
		{[]string{"settlement", "clearing"}, "fin-settlement-evidence"},
		{[]string{"reconcil"}, "fin-reconciliation-control"},
		{[]string{"pci", "compliance", "audit-log", "sensitive-data"}, "fin-data-boundaries"},
		{[]string{"sql", "database", "transaction", "isolation", "cache", "consistency"}, "dist-transaction-invariant"},
		{[]string{"message", "kafka", "nats", "outbox", "inbox", "queue"}, "dist-delivery-semantics"},
		{[]string{"retry", "resilien", "circuit", "bulkhead", "timeout", "backpressure"}, "dist-retry-budget"},
		{[]string{"lease", "lock", "coordination", "saga", "distributed"}, "dist-coordination-safety"},
		{[]string{"security", "crypto", "auth", "vulnerab"}, "eng-security-boundary"},
		{[]string{"benchmark", "performance", "profil", "pprof", "alloc"}, "eng-measure-before-optimize"},
		{[]string{"test", "fuzz", "verify", "lint", "review"}, "eng-verification-by-risk"},
		{[]string{"http", "grpc", "graphql", "swagger", "openapi", "service-bound"}, "eng-protocol-boundary"},
		{[]string{"observability", "logging", "metric", "tracing", "kubernetes", "container", "ci", "release"}, "eng-operational-lifecycle"},
		{[]string{"project", "package", "module", "dependency", "injection", "architecture", "cobra", "viper", "wire", "fx", "dig"}, "eng-dependency-direction"},
		{[]string{"error", "interface", "generic", "function", "naming", "declaration", "control-flow", "data-structure", "code-style", "style"}, "eng-language-clarity"},
	}
	for _, mapping := range mappings {
		for _, term := range mapping.terms {
			if strings.Contains(path, term) {
				return "adopted-with-qualifications", mapping.claimID, "Mapped to a conditional canonical claim; primary evidence and counterexamples control, not source wording."
			}
		}
	}
	if strings.Contains(path, "eval") || item.Kind == "scoring-assertion" {
		return "organizational-preference", "benchmark-methodology", "Retained as benchmark-design input; competitor scoring choices are not Go correctness claims."
	}
	if strings.Contains(path, "skill.md") || strings.Contains(path, "reference") || strings.Contains(path, "example") {
		return "adopted-with-qualifications", "eng-language-clarity", "General Go guidance is retained for claim-level review and is subordinate to primary Go sources."
	}
	return "outside-project-scope", "repository-mechanics", "Repository-specific prose or tooling does not define canonical Go, distributed-systems, or fintech guidance."
}

func corpusLockHash(lock CorpusLock) string {
	content, err := json.Marshal(lock)
	if err != nil {
		panic(err)
	}
	return hash(content)
}

func writeJSONAtomic(path string, value any, pattern string) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", filepath.Base(path), err)
	}
	content = append(content, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create lock directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), pattern)
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
	if err := os.Rename(temporaryName, path); err != nil {
		return err
	}
	return nil
}
