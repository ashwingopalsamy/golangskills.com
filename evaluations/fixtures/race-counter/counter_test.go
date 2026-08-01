package racecounter

import (
	"sync"
	"testing"
)

func TestCounterConcurrentInvariant(t *testing.T) {
	const workers = 32
	const additions = 200
	var counter Counter
	var wait sync.WaitGroup
	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range additions {
				counter.Add()
			}
		}()
	}
	wait.Wait()
	accepted, posted, version := counter.Snapshot()
	want := workers * additions
	if accepted != want || posted != want || version != want {
		t.Fatalf("snapshot = (%d,%d,%d), want (%d,%d,%d)", accepted, posted, version, want, want, want)
	}
}
