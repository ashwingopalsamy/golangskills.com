package goroutinelifecycle

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestProcessAllBoundsAndJoins(t *testing.T) {
	items := make([]int, 20)
	var active atomic.Int32
	var peak atomic.Int32
	var completed atomic.Int32
	err := ProcessAll(context.Background(), items, 3, func(context.Context, int) error {
		current := active.Add(1)
		defer active.Add(-1)
		for current > peak.Load() && !peak.CompareAndSwap(peak.Load(), current) {
		}
		time.Sleep(time.Millisecond)
		completed.Add(1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if completed.Load() != int32(len(items)) {
		t.Fatalf("returned after %d of %d items", completed.Load(), len(items))
	}
	if peak.Load() > 3 {
		t.Fatalf("peak concurrency = %d, want at most 3", peak.Load())
	}
}
