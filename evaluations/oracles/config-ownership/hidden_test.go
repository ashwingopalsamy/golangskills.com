package configownership

import (
	"bytes"
	"sync"
	"testing"
)

func TestStoreOwnsInputAndOutputBytes(t *testing.T) {
	store := NewStore()
	input := []byte("production")
	store.Put("service", input)
	input[0] = 'X'

	got, ok := store.Get("service")
	if !ok || string(got) != "production" {
		t.Fatalf("stored payload = %q, ok=%v", got, ok)
	}
	got[1] = 'X'
	got, _ = store.Get("service")
	if string(got) != "production" {
		t.Fatalf("Get result aliases stored payload: %q", got)
	}

	snapshot := store.Snapshot()
	snapshot["service"][2] = 'X'
	snapshot["new"] = []byte("value")
	got, _ = store.Get("service")
	if string(got) != "production" {
		t.Fatalf("Snapshot value aliases stored payload: %q", got)
	}
	if _, ok := store.Get("new"); ok {
		t.Fatal("Snapshot map aliases registry map")
	}
}

func TestCallerMutationDoesNotRaceRegistryReads(t *testing.T) {
	store := NewStore()
	input := bytes.Repeat([]byte{'a'}, 4096)
	store.Put("service", input)

	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer wait.Done()
		for index := 0; index < 10_000; index++ {
			input[index%len(input)]++
		}
	}()
	go func() {
		defer wait.Done()
		for range 10_000 {
			payload, _ := store.Get("service")
			_ = bytes.Count(payload, []byte{'a'})
		}
	}()
	wait.Wait()
}
