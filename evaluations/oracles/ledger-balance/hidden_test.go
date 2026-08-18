package ledgerbalance

import (
	"math"
	"testing"
)

func TestJournalBalancesPerCurrencyAndRejectsOverflow(t *testing.T) {
	entries := []Entry{{Currency: "USD", Amount: 100}, {Currency: "JPY", Amount: -100}}
	if err := Validate(entries); err == nil {
		t.Fatal("cross-currency netting was accepted")
	}
	balanced := []Entry{{Currency: "USD", Amount: 100}, {Currency: "USD", Amount: -100}}
	if err := Validate(balanced); err != nil {
		t.Fatalf("balanced journal rejected: %v", err)
	}
	overflowToZero := []Entry{
		{Currency: "USD", Amount: math.MaxInt64},
		{Currency: "USD", Amount: math.MaxInt64},
		{Currency: "USD", Amount: 2},
	}
	if err := Validate(overflowToZero); err == nil {
		t.Fatal("integer overflow manufactured a balanced journal")
	}
}
