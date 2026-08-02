package evaluation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckReleaseEvidenceRequiresEveryGateAndConsistentEligibility(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "evidence.txt"), []byte("evidence"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := ReleaseEvidence{SchemaVersion: 2, GeneratedOn: "2026-08-18", LeadershipClaimEligible: true}
	for _, id := range requiredReleaseGates {
		report.Gates = append(report.Gates, ReleaseGate{ID: id, Status: "pass", Summary: "verified", Evidence: []string{"evidence.txt"}})
	}
	path := writeReleaseEvidence(t, root, report)
	_, blockers, err := CheckReleaseEvidence(root, path)
	if err != nil || len(blockers) != 0 {
		t.Fatalf("CheckReleaseEvidence() blockers = %v, err = %v", blockers, err)
	}

	report.LeadershipClaimEligible = false
	report.Gates[len(report.Gates)-1].Status = "blocked"
	report.Gates[len(report.Gates)-1].Summary = "not measured"
	path = writeReleaseEvidence(t, root, report)
	_, blockers, err = CheckReleaseEvidence(root, path)
	if err != nil || len(blockers) != 1 {
		t.Fatalf("CheckReleaseEvidence() blockers = %v, err = %v", blockers, err)
	}
}

func TestCheckReleaseEvidenceRejectsBooleanOnlyManifest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "release.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":2,"generated_on":"2026-08-18","leadership_claim_eligible":true,"gates":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := CheckReleaseEvidence(root, path); err == nil {
		t.Fatal("CheckReleaseEvidence() accepted a manifest without required gates")
	}
}

func writeReleaseEvidence(t *testing.T, root string, report ReleaseEvidence) string {
	t.Helper()
	path := filepath.Join(root, "release.json")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(file).Encode(report); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
