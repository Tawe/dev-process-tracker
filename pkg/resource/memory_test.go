package resource

import (
	"os"
	"testing"
)

func TestFormatMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kb   int64
		want string
	}{
		{0, "0 KB"},
		{512, "512 KB"},
		{1024, "1.0 MB"},
		{5 * 1024, "5.0 MB"},
		{50 * 1024, "50 MB"},
		{128 * 1024, "128 MB"},
		{200 * 1024, "200 MB"},
		{500 * 1024, "500 MB"},
		{1024 * 1024, "1.0 GB"},
		{2560 * 1024, "2.5 GB"},
	}

	for _, tt := range tests {
		got := FormatMemory(tt.kb)
		if got != tt.want {
			t.Errorf("FormatMemory(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}

func TestMemoryColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kb   int64
		want string
	}{
		{0, "8"},           // gray
		{10 * 1024, "8"},   // gray (under 50 MB)
		{50 * 1024, ""},    // default (50 MB exactly)
		{100 * 1024, ""},   // default (under 200 MB)
		{200 * 1024, "11"}, // yellow (200 MB exactly)
		{300 * 1024, "11"}, // yellow
		{500 * 1024, "208"}, // orange (500 MB exactly)
		{750 * 1024, "208"}, // orange
		{1024 * 1024, "9"},  // red (1 GB exactly)
		{5 * 1024 * 1024, "9"}, // red
	}

	for _, tt := range tests {
		got := MemoryColor(tt.kb)
		if got != tt.want {
			t.Errorf("MemoryColor(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}

func TestCollectMemoryEmpty(t *testing.T) {
	t.Parallel()

	result := CollectMemory(nil)
	if result != nil {
		t.Errorf("CollectMemory(nil) = %v, want nil", result)
	}

	result = CollectMemory([]int{})
	if result != nil {
		t.Errorf("CollectMemory([]) = %v, want nil", result)
	}
}

func TestCollectMemoryInvalidPIDs(t *testing.T) {
	t.Parallel()

	result := CollectMemory([]int{0, -1})
	if len(result) != 0 {
		t.Errorf("CollectMemory with invalid PIDs = %v, want empty", result)
	}
}

func TestCollectMemoryCurrentProcess(t *testing.T) {
	pid := os.Getpid()
	result := CollectMemory([]int{pid})
	if len(result) == 0 {
		t.Fatal("CollectMemory should return memory for current process")
	}
	kb, ok := result[pid]
	if !ok {
		t.Fatal("CollectMemory should include current PID")
	}
	if kb <= 0 {
		t.Fatalf("memory for current process should be positive, got %d", kb)
	}
}

func TestCollectMemoryBatchMultiple(t *testing.T) {
	pid := os.Getpid()
	// Use current PID twice — should still work
	result := CollectMemory([]int{pid, pid})
	if len(result) == 0 {
		t.Fatal("CollectMemory should return results")
	}
	if _, ok := result[pid]; !ok {
		t.Fatal("CollectMemory should include current PID")
	}
}
