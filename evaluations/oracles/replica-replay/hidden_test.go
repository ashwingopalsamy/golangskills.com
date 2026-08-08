package replicareplay

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestConcurrentReplayConvergesAcrossServices(t *testing.T) {
	backend := NewBackend()
	services := []*Service{NewService(NewStore(backend)), NewService(NewStore(backend))}
	request := Request{Key: "transfer-1", Account: "acct-1", Amount: 1250}

	const calls = 128
	start := make(chan struct{})
	results := make(chan Result, calls)
	errorsFound := make(chan error, calls)
	var wait sync.WaitGroup
	for index := 0; index < calls; index++ {
		wait.Add(1)
		go func(service *Service) {
			defer wait.Done()
			<-start
			result, err := service.Post(context.Background(), request)
			results <- result
			errorsFound <- err
		}(services[index%len(services)])
	}
	close(start)
	wait.Wait()
	close(results)
	close(errorsFound)

	for err := range errorsFound {
		if err != nil {
			t.Fatalf("Post returned error: %v", err)
		}
	}
	var first Result
	for result := range results {
		if first == (Result{}) {
			first = result
		}
		if result != first {
			t.Fatalf("replay result = %+v, want %+v", result, first)
		}
	}
	if entries := backendEntries(backend); len(entries) != 1 || entries[0] != first {
		t.Fatalf("entries = %+v, want one entry %+v", entries, first)
	}
}

func TestConflictingPayloadReturnsNoPriorResult(t *testing.T) {
	backend := NewBackend()
	service := NewService(NewStore(backend))
	original := Request{Key: "transfer-1", Account: "acct-1", Amount: 1250}
	if _, err := service.Post(context.Background(), original); err != nil {
		t.Fatal(err)
	}

	for _, conflict := range []Request{
		{Key: original.Key, Account: "acct-2", Amount: original.Amount},
		{Key: original.Key, Account: original.Account, Amount: original.Amount + 1},
	} {
		result, err := service.Post(context.Background(), conflict)
		if !errors.Is(err, ErrKeyConflict) {
			t.Fatalf("Post conflict error = %v, want ErrKeyConflict", err)
		}
		if result != (Result{}) {
			t.Fatalf("Post conflict result = %+v, want zero Result", result)
		}
	}
	if entries := backendEntries(backend); len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
}

func TestCanceledCallDoesNotWrite(t *testing.T) {
	backend := NewBackend()
	service := NewService(NewStore(backend))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := service.Post(ctx, Request{Key: "transfer-1", Account: "acct-1", Amount: 1250})
	if !errors.Is(err, context.Canceled) || result != (Result{}) {
		t.Fatalf("Post = (%+v, %v), want zero Result and context.Canceled", result, err)
	}
	if entries := backendEntries(backend); len(entries) != 0 {
		t.Fatalf("entries = %d, want 0", len(entries))
	}
}

func backendEntries(backend *Backend) []Result {
	return NewStore(backend).Entries()
}
