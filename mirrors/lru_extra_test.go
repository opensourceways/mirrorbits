// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package mirrors

import (
	"testing"
	"time"
)

type testValue struct {
	size int
}

func (v *testValue) Size() int { return v.size }

func TestLRUCacheSetGet(t *testing.T) {
	c := NewLRUCache(1000)
	c.Set("key1", &testValue{size: 100})

	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("Expected to find key1")
	}
	if v.Size() != 100 {
		t.Fatalf("Expected size 100, got %d", v.Size())
	}
}

func TestLRUCacheGetMissing(t *testing.T) {
	c := NewLRUCache(1000)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("Expected false for missing key")
	}
}

func TestLRUCacheSetUpdate(t *testing.T) {
	c := NewLRUCache(1000)
	c.Set("key1", &testValue{size: 100})
	c.Set("key1", &testValue{size: 200})

	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("Expected to find key1")
	}
	if v.Size() != 200 {
		t.Fatalf("Expected size 200, got %d", v.Size())
	}
}

func TestLRUCacheSetIfAbsent(t *testing.T) {
	c := NewLRUCache(1000)
	c.SetIfAbsent("key1", &testValue{size: 100})
	c.SetIfAbsent("key1", &testValue{size: 200})

	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("Expected to find key1")
	}
	if v.Size() != 100 {
		t.Fatalf("Expected size 100, got %d", v.Size())
	}
}

func TestLRUCacheSetIfAbsentNew(t *testing.T) {
	c := NewLRUCache(1000)
	c.SetIfAbsent("key1", &testValue{size: 100})

	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("Expected to find key1")
	}
	if v.Size() != 100 {
		t.Fatalf("Expected size 100, got %d", v.Size())
	}
}

func TestLRUCacheDelete(t *testing.T) {
	c := NewLRUCache(1000)
	c.Set("key1", &testValue{size: 100})

	if !c.Delete("key1") {
		t.Fatal("Expected true for delete")
	}

	_, ok := c.Get("key1")
	if ok {
		t.Fatal("Expected false after delete")
	}
}

func TestLRUCacheDeleteMissing(t *testing.T) {
	c := NewLRUCache(1000)
	if c.Delete("nonexistent") {
		t.Fatal("Expected false for missing key")
	}
}

func TestLRUCacheClear(t *testing.T) {
	c := NewLRUCache(1000)
	c.Set("key1", &testValue{size: 100})
	c.Set("key2", &testValue{size: 200})
	c.Clear()

	_, ok := c.Get("key1")
	if ok {
		t.Fatal("Expected false after clear")
	}
}

func TestLRUCacheSetCapacity(t *testing.T) {
	c := NewLRUCache(1000)
	c.Set("key1", &testValue{size: 100})
	c.Set("key2", &testValue{size: 200})
	c.SetCapacity(150)

	_, ok := c.Get("key1")
	if ok {
		t.Fatal("Expected key1 to be evicted")
	}
}

func TestLRUCacheStats(t *testing.T) {
	c := NewLRUCache(1000)
	c.Set("key1", &testValue{size: 100})
	c.Set("key2", &testValue{size: 200})

	length, size, capacity, _ := c.Stats()
	if length != 2 {
		t.Fatalf("Expected length 2, got %d", length)
	}
	if size != 300 {
		t.Fatalf("Expected size 300, got %d", size)
	}
	if capacity != 1000 {
		t.Fatalf("Expected capacity 1000, got %d", capacity)
	}
}

func TestLRUCacheStatsJSON(t *testing.T) {
	c := NewLRUCache(1000)
	c.Set("key1", &testValue{size: 100})
	json := c.StatsJSON()
	if json == "" || json == "{}" {
		t.Fatal("Expected non-empty JSON")
	}
}

func TestLRUCacheStatsJSONNil(t *testing.T) {
	var c *LRUCache
	json := c.StatsJSON()
	if json != "{}" {
		t.Fatalf("Expected {}, got %s", json)
	}
}

func TestLRUCacheKeys(t *testing.T) {
	c := NewLRUCache(1000)
	c.Set("key1", &testValue{size: 100})
	c.Set("key2", &testValue{size: 200})
	c.Set("key3", &testValue{size: 300})

	keys := c.Keys()
	if len(keys) != 3 {
		t.Fatalf("Expected 3 keys, got %d", len(keys))
	}
}

func TestLRUCacheItems(t *testing.T) {
	c := NewLRUCache(1000)
	c.Set("key1", &testValue{size: 100})
	c.Set("key2", &testValue{size: 200})

	items := c.Items()
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}
}

func TestLRUCacheEviction(t *testing.T) {
	c := NewLRUCache(150)
	c.Set("key1", &testValue{size: 100})
	c.Set("key2", &testValue{size: 100})

	_, ok := c.Get("key1")
	if ok {
		t.Fatal("Expected key1 to be evicted")
	}
	_, ok = c.Get("key2")
	if !ok {
		t.Fatal("Expected key2 to exist")
	}
}

func TestLRUCacheEmpty(t *testing.T) {
	c := NewLRUCache(1000)
	length, size, _, oldest := c.Stats()
	if length != 0 || size != 0 {
		t.Fatal("Expected empty cache")
	}
	if !oldest.IsZero() {
		t.Fatal("Expected zero time for empty cache")
	}
}

func TestLRUCacheTimeAccessed(t *testing.T) {
	c := NewLRUCache(1000)
	c.Set("key1", &testValue{size: 100})
	time.Sleep(10 * time.Millisecond)
	c.Get("key1")
}
