package reconciliationnetting

import "testing"

func TestNetSettlementBuildsOneExplainedGroup(t *testing.T) {
	internal := []Item{
		{ID: "capture-1", Currency: "USD", Amount: 60},
		{ID: "capture-2", Currency: "USD", Amount: 50},
		{ID: "refund", Currency: "USD", Amount: -7},
		{ID: "fee", Currency: "USD", Amount: -5},
	}
	external := []Item{{ID: "bank-credit", Currency: "USD", Amount: 98}}
	groups := Reconcile(internal, external)
	if len(groups) != 1 || len(groups[0].Internal) != 4 || len(groups[0].External) != 1 {
		t.Fatalf("groups = %#v, want one 4:1 group", groups)
	}
	seen := map[string]bool{}
	var internalTotal, externalTotal int64
	for _, item := range groups[0].Internal {
		if seen[item.ID] {
			t.Fatalf("duplicate internal item %q", item.ID)
		}
		seen[item.ID] = true
		internalTotal += item.Amount
	}
	for _, item := range groups[0].External {
		externalTotal += item.Amount
	}
	if internalTotal != externalTotal {
		t.Fatalf("group equation %d != %d", internalTotal, externalTotal)
	}
}
