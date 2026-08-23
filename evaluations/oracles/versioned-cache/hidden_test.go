package versionedcache

import (
	"fmt"
	"sync"
	"testing"
)

func TestApplyRejectsStaleAndDuplicateVersions(t *testing.T) {
	cache := New()
	cache.Apply(Event{Key: "account", Value: "v3", Version: 3})
	cache.Apply(Event{Key: "account", Value: "v2", Version: 2})
	cache.Apply(Event{Key: "account", Value: "conflict", Version: 3})
	if value, ok := cache.Get("account"); !ok || value != "v3" {
		t.Fatalf("value = %q, ok=%v; want v3", value, ok)
	}
}

func TestDeleteRetainsOrderingAndNewerValueCanRecreate(t *testing.T) {
	cache := New()
	cache.Apply(Event{Key: "account", Value: "v3", Version: 3})
	cache.Apply(Event{Key: "account", Version: 5, Deleted: true})
	cache.Apply(Event{Key: "account", Value: "v4", Version: 4})
	if value, ok := cache.Get("account"); ok {
		t.Fatalf("stale update resurrected deleted value %q", value)
	}
	cache.Apply(Event{Key: "account", Value: "v6", Version: 6})
	if value, ok := cache.Get("account"); !ok || value != "v6" {
		t.Fatalf("value = %q, ok=%v; want v6", value, ok)
	}
}

func TestConcurrentOutOfOrderDeliveryConvergesOnNewest(t *testing.T) {
	cache := New()
	var wait sync.WaitGroup
	for version := uint64(1); version <= 64; version++ {
		wait.Add(1)
		go func(version uint64) {
			defer wait.Done()
			cache.Apply(Event{Key: "account", Value: fmt.Sprintf("v%d", version), Version: version})
		}(version)
	}
	wait.Wait()
	if value, ok := cache.Get("account"); !ok || value != "v64" {
		t.Fatalf("value = %q, ok=%v; want v64", value, ok)
	}
}
