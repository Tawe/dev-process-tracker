package resource

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Collector provides batch runtime resource metrics for processes.
type Collector struct{}

// NewCollector creates a new resource Collector.
func NewCollector() *Collector {
	return &Collector{}
}

// CollectMemory fetches RSS memory in KB for each live PID via the Collector.
func (c *Collector) CollectMemory(pids []int) map[int]int64 {
	return CollectMemory(pids)
}

// CollectMemory returns RSS memory in kilobytes for each live PID.
// Uses a single ps invocation for batch efficiency.
// Dead or inaccessible PIDs are silently omitted.
func CollectMemory(pids []int) map[int]int64 {
	if len(pids) == 0 {
		return nil
	}

	result := make(map[int]int64, len(pids))

	// Build comma-separated PID list for a single ps call
	pidStrs := make([]string, 0, len(pids))
	for _, pid := range pids {
		if pid > 0 {
			pidStrs = append(pidStrs, strconv.Itoa(pid))
		}
	}
	if len(pidStrs) == 0 {
		return result
	}

	output, err := psMemoryBatch(pidStrs)
	if err != nil {
		// Fallback: try individual lookups for each PID
		for _, pid := range pids {
			if kb := psMemorySingle(pid); kb > 0 {
				result[pid] = kb
			}
		}
		return result
	}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		kb, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}
		result[pid] = kb
	}

	return result
}

// FormatMemory renders kilobytes as a human-readable string (e.g. "128 MB", "2.4 GB").
func FormatMemory(kb int64) string {
	const (
		mb = 1024
		gb = 1024 * 1024
	)
	switch {
	case kb >= gb:
		return strconv.FormatFloat(float64(kb)/float64(gb), 'f', 1, 64) + " GB"
	case kb >= mb:
		val := kb / mb
		// Show fractional MB for values under 10 MB
		if val < 10 {
			return strconv.FormatFloat(float64(kb)/float64(mb), 'f', 1, 64) + " MB"
		}
		return strconv.FormatInt(val, 10) + " MB"
	default:
		return strconv.FormatInt(kb, 10) + " KB"
	}
}

// MemoryColor returns an ANSI color code for the given memory size in KB.
// Thresholds:
//   - gray ("8") for < 50 MB
//   - default ("") for 50–200 MB
//   - yellow ("11") for 200–500 MB
//   - orange ("208") for 500 MB–1 GB
//   - red ("9") for > 1 GB
func MemoryColor(kb int64) string {
	const (
		mb50 = 50 * 1024
		mb200 = 200 * 1024
		mb500 = 500 * 1024
		gb1   = 1024 * 1024
	)
	switch {
	case kb >= gb1:
		return "9" // red
	case kb >= mb500:
		return "208" // orange
	case kb >= mb200:
		return "11" // yellow
	case kb >= mb50:
		return "" // default
	default:
		return "8" // gray
	}
}

func psMemoryBatch(pidStrs []string) ([]byte, error) {
	pidArg := strings.Join(pidStrs, ",")
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: tasklist /FI "PID eq 123" /FO CSV — batch not practical, fallback
		return nil, exec.ErrNotFound
	}
	cmd = exec.Command("ps", "-p", pidArg, "-o", "pid=,rss=")
	return cmd.Output()
}

func psMemorySingle(pid int) int64 {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		return 0
	}
	cmd = exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "rss=")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	kb, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0
	}
	return kb
}
