package scribe

import (
	"testing"
	"time"
)

func TestUpdateNotificationsDisabled_DefaultFalse(t *testing.T) {
	setupTempHome(t)
	_ = EnsureScribeDirs()

	disabled, err := IsUpdateNotificationsDisabled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if disabled {
		t.Error("expected default to be false")
	}
}

func TestUpdateNotificationsDisabled_RoundTrip(t *testing.T) {
	setupTempHome(t)
	_ = EnsureScribeDirs()

	if err := SetUpdateNotificationsDisabled(true); err != nil {
		t.Fatalf("set true: %v", err)
	}
	disabled, err := IsUpdateNotificationsDisabled()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !disabled {
		t.Error("expected true after setting true")
	}

	if err := SetUpdateNotificationsDisabled(false); err != nil {
		t.Fatalf("set false: %v", err)
	}
	disabled, err = IsUpdateNotificationsDisabled()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if disabled {
		t.Error("expected false after setting false")
	}
}

func TestLastUpdateCheck_DefaultZero(t *testing.T) {
	setupTempHome(t)
	_ = EnsureScribeDirs()

	ts, err := GetLastUpdateCheck()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ts.IsZero() {
		t.Errorf("expected zero time, got %v", ts)
	}
}

func TestLastUpdateCheck_RoundTrip(t *testing.T) {
	setupTempHome(t)
	_ = EnsureScribeDirs()

	before := time.Now()
	if err := SetLastUpdateCheck(); err != nil {
		t.Fatalf("set: %v", err)
	}

	ts, err := GetLastUpdateCheck()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if ts.IsZero() {
		t.Fatal("expected non-zero time after setting")
	}
	if ts.Before(before.Add(-time.Second)) {
		t.Errorf("timestamp %v is before test start %v", ts, before)
	}
}

func TestShouldCheckForUpdate_NeverChecked(t *testing.T) {
	setupTempHome(t)
	_ = EnsureScribeDirs()

	if !ShouldCheckForUpdate(24 * time.Hour) {
		t.Error("should return true when never checked")
	}
}

func TestShouldCheckForUpdate_RecentlyChecked(t *testing.T) {
	setupTempHome(t)
	_ = EnsureScribeDirs()

	_ = SetLastUpdateCheck()
	if ShouldCheckForUpdate(24 * time.Hour) {
		t.Error("should return false immediately after checking")
	}
}

func TestShouldCheckForUpdate_ExpiredInterval(t *testing.T) {
	setupTempHome(t)
	_ = EnsureScribeDirs()

	// Write a timestamp 25 hours ago
	config, _ := LoadConfig()
	config.LastUpdateCheck = time.Now().Add(-25 * time.Hour).Format(time.RFC3339)
	_ = SaveConfig(config)

	if !ShouldCheckForUpdate(24 * time.Hour) {
		t.Error("should return true when interval has elapsed")
	}
}
