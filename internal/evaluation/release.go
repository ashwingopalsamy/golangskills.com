package evaluation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var requiredReleaseGates = []string{
	"reference-inventory",
	"claim-evidence",
	"structural-validation",
	"hidden-routing",
	"discovery-budget",
	"critical-correctness",
	"fixture-superiority",
	"paired-superiority",
	"domain-non-inferiority",
	"baseline-improvement",
	"token-efficiency",
	"artifact-traceability",
}

// ReleaseEvidence is the strict, machine-evaluable publication gate manifest.
// A boolean without this complete gate inventory is deliberately insufficient.
type ReleaseEvidence struct {
	SchemaVersion           int           `json:"schema_version"`
	GeneratedOn             string        `json:"generated_on"`
	LeadershipClaimEligible bool          `json:"leadership_claim_eligible"`
	Gates                   []ReleaseGate `json:"gates"`
}

// ReleaseGate links one required decision to repository-local evidence.
type ReleaseGate struct {
	ID       string   `json:"id"`
	Status   string   `json:"status"`
	Summary  string   `json:"summary"`
	Evidence []string `json:"evidence"`
}

// CheckReleaseEvidence validates completeness, internal consistency, and all
// evidence links. It returns blocked/failed gate descriptions separately from
// malformed evidence so callers cannot confuse "not yet proven" with a broken
// manifest.
func CheckReleaseEvidence(repoRoot, reportPath string) (ReleaseEvidence, []string, error) {
	file, err := os.Open(reportPath)
	if err != nil {
		return ReleaseEvidence{}, nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var report ReleaseEvidence
	if err := decoder.Decode(&report); err != nil {
		return ReleaseEvidence{}, nil, fmt.Errorf("decode release evidence: %w", err)
	}
	if report.SchemaVersion != 2 {
		return ReleaseEvidence{}, nil, fmt.Errorf("release evidence schema_version = %d; want 2", report.SchemaVersion)
	}
	if _, err := time.Parse("2006-01-02", report.GeneratedOn); err != nil {
		return ReleaseEvidence{}, nil, fmt.Errorf("release evidence generated_on must use YYYY-MM-DD: %w", err)
	}
	required := make(map[string]struct{}, len(requiredReleaseGates))
	for _, id := range requiredReleaseGates {
		required[id] = struct{}{}
	}
	seen := make(map[string]struct{}, len(report.Gates))
	var blockers, issues []string
	allPassed := true
	for _, gate := range report.Gates {
		if _, exists := required[gate.ID]; !exists {
			issues = append(issues, "unknown release gate "+gate.ID)
			continue
		}
		if _, duplicate := seen[gate.ID]; duplicate {
			issues = append(issues, "duplicate release gate "+gate.ID)
			continue
		}
		seen[gate.ID] = struct{}{}
		if gate.Summary == "" {
			issues = append(issues, gate.ID+" has no summary")
		}
		switch gate.Status {
		case "pass":
		case "blocked", "fail":
			allPassed = false
			blockers = append(blockers, gate.ID+": "+gate.Summary)
		default:
			issues = append(issues, gate.ID+" has invalid status "+gate.Status)
			allPassed = false
		}
		if len(gate.Evidence) == 0 {
			issues = append(issues, gate.ID+" has no evidence links")
		}
		for _, relative := range gate.Evidence {
			if relative == "" || filepath.IsAbs(relative) || filepath.Clean(relative) != filepath.FromSlash(relative) || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || relative == ".." {
				issues = append(issues, gate.ID+" has unsafe evidence path "+relative)
				continue
			}
			if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relative))); err != nil {
				issues = append(issues, fmt.Sprintf("%s evidence %s: %v", gate.ID, relative, err))
			}
		}
	}
	for _, id := range requiredReleaseGates {
		if _, exists := seen[id]; !exists {
			issues = append(issues, "missing release gate "+id)
			allPassed = false
		}
	}
	if report.LeadershipClaimEligible != allPassed {
		issues = append(issues, fmt.Sprintf("leadership_claim_eligible = %t but all gates pass = %t", report.LeadershipClaimEligible, allPassed))
	}
	sort.Strings(issues)
	sort.Strings(blockers)
	if len(issues) > 0 {
		return ReleaseEvidence{}, nil, fmt.Errorf("invalid release evidence: %s", strings.Join(issues, "; "))
	}
	return report, blockers, nil
}
