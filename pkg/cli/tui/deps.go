package tui

import (
	"time"

	"github.com/devports/devpt/pkg/models"
)

// AppDeps is the narrow surface the TUI needs from the CLI application layer.
type AppDeps interface {
	DiscoverServers() ([]*models.ServerInfo, error)
	ListServices() []*models.ManagedService
	GetService(name string) *models.ManagedService
	ClearServicePID(name string) error
	RegisterService(name, cwd, command string, ports []int) error
	UpdateServiceFields(name, cwd, command string, ports []int) error
	RenameService(oldName, newName string) error
	RemoveService(name string) error
	StartService(name string) error
	StopService(identifier string) error
	RestartService(name string) error
	StopProcess(pid int, timeout time.Duration) error
	TailServiceLogs(name string, lines int) ([]string, error)
	TailProcessLogs(pid int, lines int) ([]string, error)
	LatestServiceLogPath(name string) (string, error)

	// GetProcessMemory returns RSS memory in KB for each live PID.
	// Dead or inaccessible PIDs are silently omitted.
	GetProcessMemory(pids []int) map[int]int64
}
