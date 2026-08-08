package reconciliationnetting

import "testing"

func TestReconcileSupportsOneToOneAndManyToOne(t *testing.T) {
	tests := []struct {
		name         string
		internal     []Item
		external     []Item
		wantInternal int
		wantExternal int
	}{
		{
			name: "one to one", internal: []Item{{ID: "capture", Currency: "USD", Amount: 100}},
			external: []Item{{ID: "bank", Currency: "USD", Amount: 100}}, wantInternal: 1, wantExternal: 1,
		},
		{
			name:     "many internal to one external",
			internal: []Item{{ID: "a", Currency: "USD", Amount: 60}, {ID: "b", Currency: "USD", Amount: 40}},
			external: []Item{{ID: "bank", Currency: "USD", Amount: 100}}, wantInternal: 2, wantExternal: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			groups := Reconcile(test.internal, test.external)
			if len(groups) != 1 || len(groups[0].Internal) != test.wantInternal || len(groups[0].External) != test.wantExternal {
				t.Fatalf("groups = %#v", groups)
			}
			assertBalancedUnique(t, groups[0])
		})
	}
}

func TestReconcileNeverNetsCurrenciesOrReusesItems(t *testing.T) {
	internal := []Item{
		{ID: "usd-a", Currency: "USD", Amount: 60},
		{ID: "usd-b", Currency: "USD", Amount: 40},
		{ID: "eur", Currency: "EUR", Amount: 100},
	}
	external := []Item{
		{ID: "usd-bank", Currency: "USD", Amount: 100},
		{ID: "duplicate-demand", Currency: "USD", Amount: 100},
		{ID: "eur-bank", Currency: "EUR", Amount: 100},
	}
	groups := Reconcile(internal, external)
	if len(groups) != 2 {
		t.Fatalf("groups = %#v, want one USD and one EUR match", groups)
	}
	seen := map[string]bool{}
	for _, group := range groups {
		assertBalancedUnique(t, group)
		for _, item := range group.Internal {
			if seen["i:"+item.ID] {
				t.Fatalf("reused internal item %q", item.ID)
			}
			seen["i:"+item.ID] = true
		}
		for _, item := range group.External {
			if seen["e:"+item.ID] {
				t.Fatalf("reused external item %q", item.ID)
			}
			seen["e:"+item.ID] = true
		}
	}
}

func assertBalancedUnique(t *testing.T, group Group) {
	t.Helper()
	if len(group.Internal) == 0 || len(group.External) == 0 {
		t.Fatalf("empty reconciliation side: %#v", group)
	}
	seen := map[string]bool{}
	var currency string
	var internalTotal, externalTotal int64
	for _, item := range group.Internal {
		if seen["i:"+item.ID] {
			t.Fatalf("duplicate internal item %q", item.ID)
		}
		seen["i:"+item.ID] = true
		if currency == "" {
			currency = item.Currency
		}
		if item.Currency != currency {
			t.Fatalf("mixed internal currency %q with %q", item.Currency, currency)
		}
		internalTotal += item.Amount
	}
	for _, item := range group.External {
		if seen["e:"+item.ID] {
			t.Fatalf("duplicate external item %q", item.ID)
		}
		seen["e:"+item.ID] = true
		if currency == "" {
			currency = item.Currency
		}
		if item.Currency != currency {
			t.Fatalf("mixed external currency %q with %q", item.Currency, currency)
		}
		externalTotal += item.Amount
	}
	if internalTotal != externalTotal {
		t.Fatalf("group equation %d != %d", internalTotal, externalTotal)
	}
}
