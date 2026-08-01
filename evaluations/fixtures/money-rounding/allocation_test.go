package moneyrounding

import "testing"

func TestAllocationPreservesAmountAndStableResidual(t *testing.T) {
	allocation := Allocate(100, []string{"charlie", "alice", "bob"})
	if allocation["alice"]+allocation["bob"]+allocation["charlie"] != 100 {
		t.Fatalf("allocation = %v, does not preserve total", allocation)
	}
	if allocation["alice"] != 34 || allocation["bob"] != 33 || allocation["charlie"] != 33 {
		t.Fatalf("allocation = %v, want stable lexical residual", allocation)
	}
}
