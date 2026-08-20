# Cache Package

The `cache` package provides a thread-safe, in-memory **LRU (Least Recently Used) cache** for storing URL values indexed by a hash.

In addition to normal cache operations, the cache supports:

* Maximum capacity
* LRU eviction
* Entry expiration
* Dirty-node tracking
* Last-access tracking
* Safe concurrent access using `sync.RWMutex`

## Overview

Each cache entry is identified by a unique `hash` and contains:

* The associated URL
* An expiration time
* A dirty flag
* The last access time

The cache maintains two data structures:

1. A `map[string]*Node` for O(1) lookup.
2. A doubly linked list for maintaining LRU order.

The linked list is organized as:

```text
head
  ↓
[Most Recently Used]
  ↓
[Node]
  ↓
[Node]
  ↓
[Least Recently Used]
  ↓
tail
```

Whenever an item is accessed or inserted, it is moved to the front of the list.

When the cache exceeds its configured capacity, the node at the tail is removed.

---

## Cache Structure

```go
type Cache struct {
    mu        sync.RWMutex
    capacaity int
    size      int

    items map[string]*Node

    head *Node
    tail *Node
}
```

### Fields

| Field       | Description                                          |
| ----------- | ---------------------------------------------------- |
| `mu`        | Mutex used to protect concurrent access to the cache |
| `capacaity` | Maximum number of entries allowed in the cache       |
| `size`      | Current number of entries                            |
| `items`     | Hash-to-node lookup map                              |
| `head`      | Most recently used node                              |
| `tail`      | Least recently used node                             |

> **Note:** `capacaity` is currently misspelled. Consider renaming it to `capacity`.

---

## Node

Each cached value is represented by a `Node`.

```go
type Node struct {
    hash           string
    url            string
    expire         time.Time
    dirty          bool
    lastAccessTime time.Time

    next *Node
    prev *Node
}
```

### Fields

| Field            | Description                                                  |
| ---------------- | ------------------------------------------------------------ |
| `hash`           | Unique identifier used to find the cache entry               |
| `url`            | URL stored in the cache                                      |
| `expire`         | Time at which the entry becomes invalid                      |
| `dirty`          | Indicates whether the entry has changed or needs persistence |
| `lastAccessTime` | Last time the node was accessed or touched                   |
| `next`           | Next node in the LRU list                                    |
| `prev`           | Previous node in the LRU list                                |

---

# Creating a Cache

Use `NewCache` to create a cache with a fixed maximum capacity.

```go
cache := NewCache(1000)
```

### Signature

```go
func NewCache(capacaity int) *Cache
```

### Parameters

* `capacaity` — Maximum number of entries the cache can contain.

When a new entry causes the cache size to exceed this capacity, the least recently used entry is automatically removed.

---

# Put

`Put` inserts or updates a cache entry.

```go
cache.Put(
    hash,
    url,
    expire,
    dirty,
)
```

### Signature

```go
func (c *Cache) Put(
    hash string,
    url string,
    expire time.Time,
    dirty bool,
)
```

### Behavior

If the hash does not already exist:

1. A new node is created.
2. The node is inserted into the lookup map.
3. The node is placed at the front of the LRU list.
4. The cache size is incremented.
5. If the capacity is exceeded, the least recently used node is removed.

If the hash already exists:

1. The existing URL is updated.
2. The expiration time is updated.
3. The last-access time is updated.
4. The dirty state is updated.
5. The node is moved to the front of the LRU list.

### Example

```go
expire := time.Now().Add(10 * time.Minute)

cache.Put(
    "abc123",
    "https://example.com",
    expire,
    false,
)
```

---

# Get

`Get` retrieves a URL using its hash.

```go
url, ok := cache.Get("abc123")
```

### Signature

```go
func (c *Cache) Get(hash string) (string, bool)
```

The second return value indicates whether the entry was successfully found.

### Successful lookup

```go
url, ok := cache.Get("abc123")

if ok {
    fmt.Println(url)
}
```

### Cache miss

`Get` returns:

```go
"", false
```

when:

* The hash does not exist.
* The entry has expired.

### Expired entries

Expiration is checked when an entry is retrieved.

If the entry has expired, it is removed from both:

* The LRU linked list
* The lookup map

and a cache miss is returned.

### Access behavior

A successful `Get`:

* Updates `lastAccessTime`
* Marks the node as dirty
* Moves the node to the front of the LRU list

Therefore, reading an entry also updates its LRU position.

---

# LRU Eviction

The cache uses a doubly linked list to implement LRU behavior.

The list follows this structure:

```text
HEAD                                      TAIL
 ↓                                          ↓
[Most Recent] -> [ ... ] -> [Least Recent]
```

For example:

```text
A -> B -> C
```

means:

* `A` is the most recently used entry.
* `C` is the least recently used entry.

If the cache reaches its capacity and another entry is inserted, `C` is evicted.

After inserting `D`:

```text
D -> A -> B
```

---

# Internal List Operations

## addToFront

```go
func (c *Cache) addToFront(node *Node)
```

Adds a node to the head of the linked list.

The node becomes the most recently used entry.

---

## remove

```go
func (c *Cache) remove(node *Node)
```

Removes a node from the linked list.

It correctly handles:

* Removing the head
* Removing the tail
* Removing a middle node
* Removing the only node

---

## moveToFront

```go
func (c *Cache) moveToFront(node *Node)
```

Moves an existing node to the head of the LRU list.

This is used when an existing cache entry is accessed or updated.

---

# Dirty Nodes

The cache supports tracking entries that need to be persisted or synchronized elsewhere.

A node is considered dirty when:

```go
node.dirty == true
```

The cache exposes two methods for working with dirty entries:

```go
GetDirtyNodes()
MarkClean(...)
```

This allows the cache to be used together with a persistence layer.

For example:

```text
Cache
  │
  ├── dirty entry A
  ├── clean entry B
  ├── dirty entry C
  │
  ▼
GetDirtyNodes()
  │
  ▼
Persistence layer
  │
  ▼
MarkClean()
```

---

# GetDirtyNodes

Returns all cache entries currently marked as dirty.

```go
dirtyNodes := cache.GetDirtyNodes()
```

### Signature

```go
func (c *Cache) GetDirtyNodes() []contracts.DirtyNode
```

Each returned `DirtyNode` contains:

```go
type DirtyNode struct {
    Hash           string
    LastAccessTime time.Time
}
```

The `LastAccessTime` is important because it allows `MarkClean` to determine whether the entry changed after it was read for persistence.

### Example

```go
dirtyNodes := cache.GetDirtyNodes()

for _, node := range dirtyNodes {
    // Persist node to storage.
}
```

---

# MarkClean

Marks previously dirty entries as clean.

```go
cache.MarkClean(dirtyNodes)
```

### Signature

```go
func (c *Cache) MarkClean(nodes []contracts.DirtyNode)
```

The cache only marks an entry clean when its `LastAccessTime` is unchanged.

This prevents a newer update from accidentally being marked clean.

For example:

```text
1. GetDirtyNodes()
       │
       ▼
   Node A
   lastAccess = T1
       │
       ▼
2. Persistence starts
       │
       ▼
3. Node A is updated
   lastAccess = T2
       │
       ▼
4. MarkClean(T1)
       │
       ▼
Node A remains dirty
```

This protects against losing updates that occurred while persistence was running.

---

# Touch

`Touch` marks an existing cache entry as recently accessed.

```go
cache.Touch("abc123")
```

### Signature

```go
func (c *Cache) Touch(hash string)
```

If the node exists, `Touch`:

* Updates `lastAccessTime`
* Marks the node as dirty

If the node does not exist, `Touch` does nothing.

Unlike `Get`, `Touch` does not return the URL.

---

# Concurrency

The cache is designed to be safely accessed by multiple goroutines.

A `sync.RWMutex` protects the cache's:

* Map
* Linked list
* Node state
* Size

Methods that modify the cache acquire the write lock.

```go
c.mu.Lock()
defer c.mu.Unlock()
```

This prevents concurrent operations from corrupting the map or linked-list structure.

---

# Complexity

The cache is designed for constant-time cache operations.

| Operation       | Complexity |
| --------------- | ---------: |
| `Get`           |       O(1) |
| `Put`           |       O(1) |
| LRU eviction    |       O(1) |
| `Touch`         |       O(1) |
| `moveToFront`   |       O(1) |
| `remove`        |       O(1) |
| `addToFront`    |       O(1) |
| `GetDirtyNodes` |       O(n) |
| `MarkClean`     |       O(k) |

Where:

* `n` = number of entries in the cache
* `k` = number of dirty nodes supplied to `MarkClean`

The O(1) behavior of `Get` and `Put` is achieved by combining a hash map with a doubly linked list.

---

# Example

A typical usage might look like:

```go
cache := NewCache(100)

expire := time.Now().Add(5 * time.Minute)

cache.Put(
    "abc123",
    "https://example.com",
    expire,
    true,
)

url, ok := cache.Get("abc123")

if ok {
    fmt.Println("URL:", url)
}

dirtyNodes := cache.GetDirtyNodes()

// Persist dirtyNodes...

cache.MarkClean(dirtyNodes)
```

---

# Cache Lifecycle

A typical cache entry follows this lifecycle:

```text
             Put
              │
              ▼
        ┌─────────────┐
        │ Cache Entry │
        └──────┬──────┘
               │
       ┌───────┴────────┐
       │                │
       ▼                ▼
     Get()            Touch()
       │                │
       ▼                ▼
 lastAccessTime    lastAccessTime
 updated           updated
       │                │
       └───────┬────────┘
               ▼
            Dirty
               │
               ▼
       GetDirtyNodes()
               │
               ▼
          Persistence
               │
               ▼
          MarkClean()
               │
               ▼
            Clean
```

An entry can also expire at any point. When an expired entry is requested through `Get`, it is removed from the cache.

---

# Design Notes

## Why use both a map and a linked list?

The map provides fast lookup:

```text
hash → Node
```

The linked list provides fast LRU ordering:

```text
HEAD → ... → TAIL
```

Using both structures allows the cache to perform lookup, insertion, removal, and LRU promotion in O(1) time.

## Why track `lastAccessTime`?

`lastAccessTime` serves two purposes:

1. It records when the cache entry was accessed.
2. It acts as a version-like value when dirty entries are persisted.

`MarkClean` uses this value to avoid clearing the dirty flag for an entry that changed after `GetDirtyNodes` was called.

---

# Important Implementation Considerations

### Capacity

`NewCache` currently accepts any integer capacity. A capacity of `0` means every inserted item will immediately be evicted.

It may be preferable to validate capacity during construction.

### Expiration

Expiration is checked during `Get`. Expired entries that are never accessed again remain in the cache until another operation removes them through normal eviction.

If automatic expiration is required, a background cleanup mechanism would be needed.

### Mutex

Although the cache uses `sync.RWMutex`, `Get` currently uses `Lock()` rather than `RLock()` because `Get` modifies the node by updating:

* `lastAccessTime`
* `dirty`
* LRU position

Therefore, a read lock would not be sufficient with the current implementation.

---

# Package Contract

The cache provides the following public API:

```go
func NewCache(capacaity int) *Cache

func (c *Cache) Get(hash string) (string, bool)

func (c *Cache) Put(
    hash string,
    url string,
    expire time.Time,
    dirty bool,
)

func (c *Cache) GetDirtyNodes() []contracts.DirtyNode

func (c *Cache) MarkClean(nodes []contracts.DirtyNode)

func (c *Cache) Touch(hash string)
```

The internal linked-list methods are intentionally unexported:

```go
func (c *Cache) addToFront(node *Node)
func (c *Cache) remove(node *Node)
func (c *Cache) moveToFront(node *Node)
```

This keeps LRU implementation details hidden from consumers of the package.

