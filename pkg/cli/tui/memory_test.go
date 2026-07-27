package tui

import (
	"strings"
	"testing"
)

func TestDetailsPane_MemoryDisplay(t *testing.T) {
	t.Parallel()

	t.Run("memory shown in running service details", func(t *testing.T) {
		m := newTestModel()
		m.width = 120
		m.height = 40
		m.focus = focusRunning
		m.selected = 0
		// Simulate memory data for PID 1001 (the test server)
		m.memory[1001] = 128 * 1024 // 128 MB

		visible := m.visibleServers()
		managed := m.managedServices()
		details := m.renderSelectedServiceDetails(60, visible, managed)

		if !strings.Contains(details, "Memory:") {
			t.Fatal("details pane should contain 'Memory:' line")
		}
		if !strings.Contains(details, "128 MB") {
			t.Fatalf("details pane should show '128 MB', got:\n%s", details)
		}
		if !strings.Contains(details, "Memory:") {
			t.Fatal("details pane should contain Memory line")
		}
	})

	t.Run("memory omitted when no data available", func(t *testing.T) {
		m := newTestModel()
		m.width = 120
		m.height = 40
		m.focus = focusRunning
		m.selected = 0
		// No memory data — should not show Memory line

		visible := m.visibleServers()
		managed := m.managedServices()
		details := m.renderSelectedServiceDetails(60, visible, managed)

		if strings.Contains(details, "Memory:") {
			t.Fatal("details pane should not show Memory when no data available")
		}
	})

	t.Run("memory color thresholds", func(t *testing.T) {
		tests := []struct {
			kb       int64
			contains string
		}{
			{8 * 1024, "8.0 MB"},       // gray (under 50 MB, fractional)
			{128 * 1024, "128 MB"},     // default
			{312 * 1024, "312 MB"},     // yellow (200-500 MB)
			{780 * 1024, "780 MB"},     // orange (500 MB - 1 GB)
			{2 * 1024 * 1024, "2.0 GB"}, // red (>1 GB)
		}

		for _, tt := range tests {
			m := newTestModel()
			m.width = 120
			m.height = 40
			m.focus = focusRunning
			m.selected = 0
			m.memory[1001] = tt.kb

			visible := m.visibleServers()
			managed := m.managedServices()
			details := m.renderSelectedServiceDetails(60, visible, managed)

			if !strings.Contains(details, tt.contains) {
				t.Errorf("for %d KB, expected details to contain %q, got:\n%s", tt.kb, tt.contains, details)
			}
		}
	})
}

func TestMemoryMsg_UpdatesMemoryMap(t *testing.T) {
	m := newTestModel()
	m.memoryBusy = true

	msg := memoryMsg{memory: map[int]int64{
		1001: 256 * 1024,
		2002: 512 * 1024,
	}}

	model, cmd := m.Update(msg)
	_ = model

	if m.memoryBusy {
		t.Fatal("memoryBusy should be false after memoryMsg")
	}
	// Async handlers return nil — the main tick loop drives the heartbeat
	if cmd != nil {
		t.Fatalf("memoryMsg should return nil cmd, got %v", cmd)
	}
	if m.memory[1001] != 256*1024 {
		t.Fatalf("memory[1001] = %d, want %d", m.memory[1001], 256*1024)
	}
	if m.memory[2002] != 512*1024 {
		t.Fatalf("memory[2002] = %d, want %d", m.memory[2002], 512*1024)
	}
}
