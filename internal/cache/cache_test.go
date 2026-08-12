package cache

import (
	"testing"
	"time"
)

func TestCachePutGet(t *testing.T) {
	c := NewCache(10)

	c.Put(
		"hash1",
		"https://example.com",
		time.Now().Add(time.Hour),
		false,
	)

	got, ok := c.Get("hash1")

	if !ok {
		t.Fatal("expected cache hit")
	}

	if got != "https://example.com" {
		t.Errorf("got %q, want %q", got, "https://example.com")
	}
}

func TestCacheGetMissing(t *testing.T) {
	c := NewCache(10)

	_, ok := c.Get("missing")

	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestCacheExpiration(t *testing.T) {
	c := NewCache(10)

	c.Put(
		"hash1",
		"https://example.com",
		time.Now().Add(-time.Second),
		false,
	)

	_, ok := c.Get("hash1")

	if ok {
		t.Fatal("expected expired item to be a cache miss")
	}
}

func TestCacheEvictsLeastRecentlyUsed(t *testing.T) {
	c := NewCache(2)

	expire := time.Now().Add(time.Hour)

	c.Put("hash1", "url1", expire, false)
	c.Put("hash2", "url2", expire, false)

	// Access hash1, making hash2 the least recently used.
	_, ok := c.Get("hash1")
	if !ok {
		t.Fatal("expected hash1 to exist")
	}

	// This should evict hash2.
	c.Put("hash3", "url3", expire, false)

	if _, ok := c.Get("hash2"); ok {
		t.Fatal("expected hash2 to be evicted")
	}

	if _, ok := c.Get("hash1"); !ok {
		t.Fatal("expected hash1 to remain")
	}

	if _, ok := c.Get("hash3"); !ok {
		t.Fatal("expected hash3 to exist")
	}
}

func TestCacheUpdate(t *testing.T) {
	c := NewCache(10)

	expire := time.Now().Add(time.Hour)

	c.Put("hash1", "url1", expire, false)
	c.Put("hash1", "url2", expire, false)

	got, ok := c.Get("hash1")

	if !ok {
		t.Fatal("expected cache hit")
	}

	if got != "url2" {
		t.Errorf("got %q, want %q", got, "url2")
	}

	if c.size != 1 {
		t.Errorf("size = %d, want 1", c.size)
	}
}
func TestCacheDirtyNodes(t *testing.T) {
	c := NewCache(10)

	expire := time.Now().Add(time.Hour)

	c.Put("hash1", "url1", expire, true)
	c.Put("hash2", "url2", expire, false)

	dirty := c.GetDirtyNodes()

	if len(dirty) != 1 {
		t.Fatalf("got %d dirty nodes, want 1", len(dirty))
	}

	if dirty[0].Hash != "hash1" {
		t.Errorf("got hash %q, want %q", dirty[0].Hash, "hash1")
	}
}

func TestCacheMarkClean(t *testing.T) {
	c := NewCache(10)

	expire := time.Now().Add(time.Hour)

	c.Put("hash1", "url1", expire, true)

	dirty := c.GetDirtyNodes()

	if len(dirty) != 1 {
		t.Fatalf("got %d dirty nodes, want 1", len(dirty))
	}

	c.MarkClean(dirty)

	if len(c.GetDirtyNodes()) != 0 {
		t.Fatal("expected no dirty nodes after MarkClean")
	}
}

func TestCacheMarkCleanDoesNotCleanUpdatedNode(t *testing.T) {
	c := NewCache(10)

	expire := time.Now().Add(time.Hour)

	c.Put("hash1", "url1", expire, true)

	dirty := c.GetDirtyNodes()

	// Simulate the node being updated after the dirty snapshot.
	time.Sleep(time.Millisecond)
	c.Put("hash1", "url2", expire, true)

	c.MarkClean(dirty)

	nodes := c.GetDirtyNodes()

	if len(nodes) != 1 {
		t.Fatal("expected updated node to remain dirty")
	}

	if nodes[0].Hash != "hash1" {
		t.Errorf("got hash %q, want %q", nodes[0].Hash, "hash1")
	}
}

func TestCacheTouch(t *testing.T) {
	c := NewCache(10)

	expire := time.Now().Add(time.Hour)

	c.Put("hash1", "url1", expire, false)

	c.Touch("hash1")

	dirty := c.GetDirtyNodes()

	if len(dirty) != 1 {
		t.Fatal("expected touched node to be dirty")
	}

	if dirty[0].Hash != "hash1" {
		t.Errorf("got hash %q, want %q", dirty[0].Hash, "hash1")
	}
}
