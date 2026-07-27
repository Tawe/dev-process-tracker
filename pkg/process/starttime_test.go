package process

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestGetProcessStartTimeCurrentProcess(t *testing.T) {
	t.Parallel()

	m := NewManager(t.TempDir())
	startTime, err := m.GetProcessStartTime(os.Getpid())
	if errors.Is(err, ErrProcessStartTimeUnavailable) {
		t.Skipf("process start-time lookup unsupported on this platform: %v", err)
	}
	if err != nil {
		t.Fatalf("GetProcessStartTime current pid: %v", err)
	}
	if startTime.IsZero() {
		t.Fatal("start time should not be zero")
	}
	if startTime.After(time.Now().Add(1 * time.Second)) {
		t.Fatalf("start time %v should not be in the future", startTime)
	}

	startTime2, err := m.GetProcessStartTime(os.Getpid())
	if err != nil {
		t.Fatalf("GetProcessStartTime second call: %v", err)
	}
	if !startTime.Equal(startTime2) {
		t.Fatalf("start time should be stable, got %v then %v", startTime, startTime2)
	}
}

func TestGetProcessStartTimeInvalidPID(t *testing.T) {
	t.Parallel()

	m := NewManager(t.TempDir())
	if _, err := m.GetProcessStartTime(0); err == nil {
		t.Fatal("expected invalid PID error")
	}
	if _, err := m.GetProcessStartTime(-1); err == nil {
		t.Fatal("expected invalid PID error")
	}
}
