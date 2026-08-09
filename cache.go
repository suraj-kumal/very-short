package veryshort

import (
	"sync"
	"time"
)

type Cache struct {
	mu        sync.RWMutex
	capacaity int
	size      int

	items map[string]*Node

	head *Node
	tail *Node
}

type Node struct {
	// hash is key
	hash           string
	url            string
	expire         time.Time
	dirty          bool
	lastAccessTime time.Time

	next *Node
	prev *Node
}

type DirtyNode struct {
	Hash           string
	LastAccessTime time.Time
}

// constructor
func NewCache(capacaity int) *Cache {
	return &Cache{
		capacaity: capacaity,
		items:     make(map[string]*Node),
		head:      nil,
		tail:      nil,
	}
}

func (c *Cache) addToFront(node *Node) {
	node.prev = nil
	node.next = c.head

	if c.head != nil {
		c.head.prev = node
	}

	c.head = node

	if c.tail == nil {
		c.tail = node
	}

	c.size++
}

func (c *Cache) remove(node *Node) {
	if node == nil {
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		// Removing head
		c.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		// Removing tail
		c.tail = node.prev
	}

	node.prev = nil
	node.next = nil

	c.size--
}

func (c *Cache) moveToFront(node *Node) {
	if node == c.head {
		return
	}

	c.remove(node)
	c.addToFront(node)
}

func (c *Cache) Get(hash string) (string, bool) {
	c.mu.Lock()

	defer c.mu.Unlock()
	node, ok := c.items[hash]
	if !ok {
		return "", false
	}

	// Check expiration
	if time.Now().After(node.expire) {
		c.remove(node)
		delete(c.items, hash)
		return "", false
	}

	node.lastAccessTime = time.Now()
	node.dirty = true
	c.moveToFront(node)

	return node.url, true
}

func (c *Cache) Put(hash, url string, expire time.Time, dirty bool) {
	c.mu.Lock()

	defer c.mu.Unlock()

	if node, ok := c.items[hash]; ok {
		node.url = url
		node.expire = expire
		node.lastAccessTime = time.Now()
		node.dirty = dirty

		c.moveToFront(node)
		return
	}

	node := &Node{
		hash:           hash,
		url:            url,
		expire:         expire,
		lastAccessTime: time.Now(),
		dirty:          dirty,
	}

	c.items[hash] = node
	c.addToFront(node)

	if c.size > c.capacaity {
		lru := c.tail
		c.remove(lru)
		delete(c.items, lru.hash)
	}
}

func (c *Cache) GetDirtyNodes() []DirtyNode {
	c.mu.Lock()
	defer c.mu.Unlock()

	dirtyNodes := make([]DirtyNode, 0)

	for _, node := range c.items {
		if node.dirty {
			dirtyNodes = append(dirtyNodes, DirtyNode{
				Hash:           node.hash,
				LastAccessTime: node.lastAccessTime,
			})
		}
	}

	return dirtyNodes
}

func (c *Cache) MarkClean(nodes []DirtyNode) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, dirtyNode := range nodes {
		node, ok := c.items[dirtyNode.Hash]

		if !ok {
			continue
		}

		// Only clean if it was not updated again
		if node.lastAccessTime.Equal(dirtyNode.LastAccessTime) {
			node.dirty = false
		}
	}
}

func (c *Cache) Touch(hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.items[hash]
	if !ok {
		return
	}

	node.lastAccessTime = time.Now()
	node.dirty = true
}
