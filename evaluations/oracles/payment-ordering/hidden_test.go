package paymentordering

import "testing"

func TestFutureEvidenceIsReevaluatedAndInvalidEvidenceRejected(t *testing.T) {
	var projector Projector
	if err := projector.Apply(Refunded); err != nil {
		t.Fatalf("future refund rejected: %v", err)
	}
	if captured, refunded, pending := projector.State(); captured || refunded || pending != 1 {
		t.Fatalf("future state = captured:%v refunded:%v pending:%d", captured, refunded, pending)
	}
	if err := projector.Apply(Captured); err != nil {
		t.Fatal(err)
	}
	captured, refunded, pending := projector.State()
	if !captured || !refunded || pending != 0 {
		t.Fatalf("state = captured:%v refunded:%v pending:%d", captured, refunded, pending)
	}
	if err := projector.Apply(Event("unknown")); err == nil {
		t.Fatal("unknown evidence was silently accepted")
	}
}
