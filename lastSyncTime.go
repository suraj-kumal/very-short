package main

import (
	"sync"
	"time"
)

type SyncState struct {
	mu           sync.RWMutex
	lastSyncTime time.Time
	running      bool
}

func NewSyncState() *SyncState {
	return &SyncState{}
}

func (s *SyncState) Set(t time.Time) {
	s.mu.Lock()

	defer s.mu.Unlock()

	s.lastSyncTime = t
}

func (s *SyncState) Get() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lastSyncTime
}

func (s *SyncState) CheckInterval(t time.Time) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int(t.Sub(s.lastSyncTime).Minutes())
}

func (s *SyncState) TryStartSync() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return false
	}

	if time.Since(s.lastSyncTime) < 10*time.Minute {
		return false
	}

	s.running = true
	return true
}

func (s *SyncState) FinishSync() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastSyncTime = time.Now()
	s.running = false
}
