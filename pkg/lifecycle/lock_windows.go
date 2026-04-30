//go:build windows

package lifecycle

import (
	"os/exec"
	"strconv"
)

func lockProcessAlive(pid int) bool {
	err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid)).Run()
	return err == nil
}
