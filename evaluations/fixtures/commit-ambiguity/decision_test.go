package commitambiguity

import "testing"

func TestLostCommitResponseIsNotDefiniteRollback(t *testing.T) {
	if got := Decide(true, true); got != Committed {
		t.Fatalf("found operation: got %q, want %q", got, Committed)
	}
	if got := Decide(true, false); got != Reconcile {
		t.Fatalf("unknown operation: got %q, want %q", got, Reconcile)
	}
}
