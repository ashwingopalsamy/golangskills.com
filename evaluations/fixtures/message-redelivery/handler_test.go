package messageredelivery

import "testing"

type fakeStore struct {
	seen  map[string]bool
	order *[]string
}

func (s *fakeStore) ClaimAndApply(id string) (bool, error) {
	*s.order = append(*s.order, "commit")
	if s.seen[id] {
		return false, nil
	}
	s.seen[id] = true
	return true, nil
}

func TestHandleCommitsBeforeAcknowledgementAndDeduplicates(t *testing.T) {
	var order []string
	store := &fakeStore{seen: map[string]bool{}, order: &order}
	ack := func() error { order = append(order, "ack"); return nil }
	if err := Handle("m1", store, ack); err != nil {
		t.Fatal(err)
	}
	if len(order) < 2 || order[0] != "commit" || order[1] != "ack" {
		t.Fatalf("order = %v, want commit before ack", order)
	}
	if err := Handle("m1", store, ack); err != nil {
		t.Fatal(err)
	}
	if len(store.seen) != 1 {
		t.Fatalf("effects = %d, want one", len(store.seen))
	}
}
