package retryamplification

import (
	"errors"
	"testing"
)

func TestOneLayerOwnsRetryBudget(t *testing.T) {
	attempts := 0
	err := Call(3, 3, func() error { attempts++; return errors.New("down") })
	if err == nil {
		t.Fatal("expected failure")
	}
	if attempts != 3 {
		t.Fatalf("leaf attempts = %d, want 3", attempts)
	}
}
