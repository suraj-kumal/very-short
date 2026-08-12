package timesync

import (
	"testing"
	"time"
)

func TestSyncStateSetGet(t *testing.T) {
	state := NewSyncState()

	now := time.Now()

	state.Set(now)

	got := state.Get()

	if !got.Equal(now) {
		t.Errorf("Get() = %v, want %v", got, now)
	}
}

func TestSyncStateCheckInterval(t *testing.T) {
	state := NewSyncState()

	lastSync := time.Now().Add(-30 * time.Minute)
	state.Set(lastSync)

	got := state.CheckInterval(time.Now())

	if got != 30 {
		t.Errorf("CheckInterval() = %d, want 30", got)
	}
}

func TestSyncStateTryStartSync(t *testing.T) {
	state := NewSyncState()

	// Last sync was more than 10 minutes ago.
	state.Set(time.Now().Add(-11 * time.Minute))

	if !state.TryStartSync() {
		t.Fatal("expected TryStartSync() to return true")
	}

	// A second sync should not start while one is already running.
	if state.TryStartSync() {
		t.Fatal("expected second TryStartSync() to return false")
	}
}

func TestSyncStateTryStartSyncTooSoon(t *testing.T) {
	state := NewSyncState()

	// Last sync was less than 10 minutes ago.
	state.Set(time.Now().Add(-5 * time.Minute))

	if state.TryStartSync() {
		t.Fatal("expected TryStartSync() to return false")
	}
}

func TestSyncStateFinishSync(t *testing.T) {
	state := NewSyncState()

	state.Set(time.Now().Add(-20 * time.Minute))

	if !state.TryStartSync() {
		t.Fatal("expected sync to start")
	}

	state.FinishSync()

	// FinishSync should make the state recently synced,
	// so another sync should not start immediately.
	if state.TryStartSync() {
		t.Fatal("expected sync not to start immediately after FinishSync")
	}

	if state.Get().IsZero() {
		t.Fatal("expected lastSyncTime to be updated")
	}
}
