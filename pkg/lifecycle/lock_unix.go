//go:build !windows

package lifecycle

import "syscall"

func lockProcessAlive(pid int) bool {
	return syscall.Kill(pid, syscall.Signal(0)) == nil
}
