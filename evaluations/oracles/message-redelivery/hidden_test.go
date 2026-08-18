package messageredelivery

import (
	"errors"
	"testing"
)

type fakeStore struct {
	seen  map[string]bool
	order *[]string
	err   error
}

func (s *fakeStore) ClaimAndApply(id string) (bool, error) {
	*s.order = append(*s.order, "commit")
	if s.err != nil {
		return false, s.err
	}
	if s.seen[id] {
		return false, nil
	}
	s.seen[id] = true
	return true, nil
}

func TestHandleOrdersDurableEffectAndAck(t *testing.T) {
	var order []string
	store := &fakeStore{seen: map[string]bool{}, order: &order}
	ack := func() error { order = append(order, "ack"); return nil }
	if err := Handle("m1", store, ack); err != nil {
		t.Fatal(err)
	}
	if len(order) != 2 || order[0] != "commit" || order[1] != "ack" {
		t.Fatalf("order = %v, want commit then ack", order)
	}
	if err := Handle("m1", store, ack); err != nil {
		t.Fatal(err)
	}
	if len(store.seen) != 1 {
		t.Fatalf("effects = %d, want one", len(store.seen))
	}
}

func TestHandleDoesNotAckFailedCommitAndReturnsAckFailure(t *testing.T) {
	commitErr := errors.New("commit failed")
	var order []string
	store := &fakeStore{seen: map[string]bool{}, order: &order, err: commitErr}
	acked := false
	if err := Handle("m1", store, func() error { acked = true; return nil }); !errors.Is(err, commitErr) {
		t.Fatalf("commit error = %v", err)
	}
	if acked {
		t.Fatal("acknowledged after failed commit")
	}

	ackErr := errors.New("ack failed")
	store.err = nil
	if err := Handle("m2", store, func() error { return ackErr }); !errors.Is(err, ackErr) {
		t.Fatalf("ack error = %v", err)
	}
}
