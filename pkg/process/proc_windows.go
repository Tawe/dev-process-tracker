//go:build windows

package process

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func setProcessGroup(cmd *exec.Cmd) {
	// Windows: no special process group setup needed for basic use
	// The process will be managed by its PID
}

func terminateProcess(pid int) error {
	return terminateProcessFallback(pid)
}

func terminateProcessFallback(pid int) error {
	// On Windows, use taskkill for graceful termination
	return exec.Command("taskkill", "/PID", strconv.Itoa(pid)).Run()
}

func killProcess(pid int) error {
	return killProcessFallback(pid)
}

func killProcessFallback(pid int) error {
	// On Windows, use taskkill /F for forceful termination
	return exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid)).Run()
}

func isProcessAlive(pid int) bool {
	// Check if process exists using tasklist
	err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid)).Run()
	return err == nil
}

func getProcessStartTime(pid int) (time.Time, error) {
	return time.Time{}, ErrProcessStartTimeUnavailable
}

// getProcessCommand returns the command line for a live process via wmic.
// wmic is deprecated on newer Windows but still ships on GitHub Actions
// runners; if it is missing or the lookup fails we surface a clear error so
// callers (which only use this for best-effort display) can degrade gracefully.
func getProcessCommand(pid int) (string, error) {
	out, err := exec.Command("wmic", "process", "where", "ProcessId="+strconv.Itoa(pid), "get", "CommandLine", "/value").Output()
	if err != nil {
		return "", fmt.Errorf("wmic command for pid %d: %v", pid, err)
	}
	// wmic /value emits "CommandLine=<...>\r\n\r\n"; strip the key and whitespace.
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "CommandLine=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "CommandLine=")), nil
		}
	}
	return "", fmt.Errorf("wmic returned no CommandLine for pid %d", pid)
}
