package moneyrounding

import (
	"reflect"
	"testing"
)

func TestAllocationPreservesAmountAndStableResidual(t *testing.T) {
	recipients := []string{"charlie", "alice", "bob"}
	wantInput := append([]string(nil), recipients...)
	allocation := Allocate(100, recipients)
	if allocation["alice"]+allocation["bob"]+allocation["charlie"] != 100 {
		t.Fatalf("allocation = %v, does not preserve total", allocation)
	}
	if allocation["alice"] != 34 || allocation["bob"] != 33 || allocation["charlie"] != 33 {
		t.Fatalf("allocation = %v, want stable lexical residual", allocation)
	}
	if !reflect.DeepEqual(recipients, wantInput) {
		t.Fatalf("input mutated from %v to %v", wantInput, recipients)
	}

	negative := Allocate(-100, []string{"bob", "charlie", "alice"})
	if negative["alice"] != -34 || negative["bob"] != -33 || negative["charlie"] != -33 {
		t.Fatalf("negative allocation = %v", negative)
	}
	if got := Allocate(10, nil); len(got) != 0 {
		t.Fatalf("empty allocation = %v", got)
	}
}
