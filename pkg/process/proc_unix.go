//go:build !windows

package process

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func terminateProcess(pid int) error {
	return syscall.Kill(-pid, syscall.SIGTERM)
}

func terminateProcessFallback(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

func killProcess(pid int) error {
	return syscall.Kill(-pid, syscall.SIGKILL)
}

func killProcessFallback(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}

func isProcessAlive(pid int) bool {
	return syscall.Kill(pid, syscall.Signal(0)) == nil
}

func getProcessStartTime(pid int) (time.Time, error) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart=")
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: ps start time for pid %d: %v", ErrProcessStartTimeUnavailable, pid, err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return time.Time{}, fmt.Errorf("%w: empty ps start time for pid %d", ErrProcessStartTimeUnavailable, pid)
	}

	layouts := []string{
		"Mon Jan _2 15:04:05 2006",
		"Mon Jan 2 15:04:05 2006",
	}
	for _, layout := range layouts {
		startTime, parseErr := time.ParseInLocation(layout, raw, time.Local)
		if parseErr == nil {
			return startTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("%w: cannot parse %q for pid %d", ErrProcessStartTimeUnavailable, raw, pid)
}

func getProcessCommand(pid int) (string, error) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ps command for pid %d: %v", pid, err)
	}
	return strings.TrimSpace(string(out)), nil
}
