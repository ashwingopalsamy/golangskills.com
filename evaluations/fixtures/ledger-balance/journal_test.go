package ledgerbalance

import "testing"

func TestJournalBalancesPerCurrency(t *testing.T) {
	entries := []Entry{{Currency: "USD", Amount: 100}, {Currency: "JPY", Amount: -100}}
	if err := Validate(entries); err == nil {
		t.Fatal("cross-currency netting was accepted")
	}
	balanced := []Entry{{Currency: "USD", Amount: 100}, {Currency: "USD", Amount: -100}}
	if err := Validate(balanced); err != nil {
		t.Fatalf("balanced journal rejected: %v", err)
	}
}
