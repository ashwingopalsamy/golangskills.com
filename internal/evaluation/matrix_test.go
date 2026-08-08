package evaluation

import (
	"reflect"
	"testing"

	"github.com/ashwingopalsamy/golangskills.com/internal/corpus"
)

func TestPlanMatrixCellsSelectsSameCasesAndRandomizesAllArms(t *testing.T) {
	t.Parallel()
	activate := true
	collection := corpus.Collection{Skills: []corpus.Skill{
		{Name: "skill-a", Evaluations: corpus.Evaluations{Cases: []corpus.EvalCase{{ID: "one", Kind: "routing", ShouldActivate: &activate}}}},
		{Name: "skill-b", Evaluations: corpus.Evaluations{Cases: []corpus.EvalCase{{ID: "two", Kind: "routing", ShouldActivate: &activate}}}},
	}}
	arms := []RunOptions{{Arm: "baseline"}, {Arm: "ours"}}
	options := MatrixOptions{Runner: "codex", Kind: "routing", Limit: 1, Seed: 42}
	first, err := planMatrixCells(collection, arms, options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := planMatrixCells(collection, arms, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 2 || len(second) != 2 {
		t.Fatalf("cell counts = %d and %d; want two arms times one case", len(first), len(second))
	}
	keys := func(cells []matrixCell) []string {
		result := make([]string, 0, len(cells))
		for _, cell := range cells {
			result = append(result, cell.options.Arm+"/"+cell.item.skill.Name+"/"+cell.item.eval.ID)
		}
		return result
	}
	if !reflect.DeepEqual(keys(first), keys(second)) {
		t.Fatalf("same seed produced different plans: %v and %v", keys(first), keys(second))
	}
	if first[0].item.eval.ID != first[1].item.eval.ID || first[0].options.Arm == first[1].options.Arm {
		t.Fatalf("matrix did not select one shared case across distinct arms: %v", keys(first))
	}
}

func TestBenchmarkIdentitySeparatesModeRepetitionAndOpaqueArms(t *testing.T) {
	t.Parallel()
	native := benchmarkKey("ours", "skill", "case", "native", 0)
	explicit := benchmarkKey("ours", "skill", "case", "explicit", 0)
	repeated := benchmarkKey("ours", "skill", "case", "native", 1)
	if native == explicit || native == repeated || explicit == repeated {
		t.Fatalf("benchmark identities collided: %q %q %q", native, explicit, repeated)
	}
	if opaqueLabel(7, "ours") == opaqueLabel(7, "competitor") {
		t.Fatal("opaque labels collided across arms")
	}
}
