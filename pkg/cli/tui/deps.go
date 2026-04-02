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
	AddCmd(name, cwd, command string, ports []int) error
	RemoveCmd(name string) error
	StartCmd(name string) error
	StopCmd(identifier string) error
	RestartCmd(name string) error
	StopProcess(pid int, timeout time.Duration) error
	TailServiceLogs(name string, lines int) ([]string, error)
	TailProcessLogs(pid int, lines int) ([]string, error)
	LatestServiceLogPath(name string) (string, error)
}
