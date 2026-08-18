package goroutinelifecycle

import (
	"context"
	"errors"
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
		for observed := peak.Load(); current > observed && !peak.CompareAndSwap(observed, current); observed = peak.Load() {
		}
		time.Sleep(time.Millisecond)
		completed.Add(1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if completed.Load() != int32(len(items)) || active.Load() != 0 {
		t.Fatalf("completed=%d active=%d", completed.Load(), active.Load())
	}
	if peak.Load() > 3 {
		t.Fatalf("peak concurrency = %d, want at most 3", peak.Load())
	}
}

func TestProcessAllHonorsCancellationAndErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var called atomic.Int32
	if err := ProcessAll(ctx, []int{1, 2, 3}, 2, func(context.Context, int) error {
		called.Add(1)
		return nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled call error = %v", err)
	}
	if called.Load() != 0 {
		t.Fatalf("started %d calls after cancellation", called.Load())
	}

	want := errors.New("process failed")
	if err := ProcessAll(context.Background(), []int{1, 2, 3}, 2, func(context.Context, int) error { return want }); !errors.Is(err, want) {
		t.Fatalf("process error = %v, want %v", err, want)
	}
	if err := ProcessAll(context.Background(), []int{1}, 0, func(context.Context, int) error { return nil }); err == nil {
		t.Fatal("non-positive concurrency limit accepted")
	}
}
