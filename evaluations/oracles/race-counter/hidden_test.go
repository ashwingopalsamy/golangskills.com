package racecounter

import (
	"sync"
	"testing"
)

func TestCounterConcurrentInvariant(t *testing.T) {
	const workers = 32
	const additions = 200
	var counter Counter
	var writers sync.WaitGroup
	for range workers {
		writers.Add(1)
		go func() {
			defer writers.Done()
			for range additions {
				counter.Add()
			}
		}()
	}
	stopReaders := make(chan struct{})
	var readers sync.WaitGroup
	readers.Add(1)
	go func() {
		defer readers.Done()
		for {
			select {
			case <-stopReaders:
				return
			default:
			}
			accepted, posted, version := counter.Snapshot()
			if accepted != posted || posted != version {
				t.Errorf("torn snapshot = (%d,%d,%d)", accepted, posted, version)
				return
			}
		}
	}()
	writers.Wait()
	close(stopReaders)
	readers.Wait()
	accepted, posted, version := counter.Snapshot()
	want := workers * additions
	if accepted != want || posted != want || version != want {
		t.Fatalf("snapshot = (%d,%d,%d), want (%d,%d,%d)", accepted, posted, version, want, want, want)
	}
}
