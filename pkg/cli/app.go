package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/devports/devpt/pkg/health"
	"github.com/devports/devpt/pkg/lifecycle"
	"github.com/devports/devpt/pkg/models"
	"github.com/devports/devpt/pkg/process"
	"github.com/devports/devpt/pkg/registry"
	"github.com/devports/devpt/pkg/scanner"
)

var warnLegacyCommandsOnce sync.Once

// App is the main application handler
type App struct {
	config         models.ConfigPaths
	registry       *registry.Registry
	scanner        *scanner.ProcessScanner
	resolver       *scanner.ProjectResolver
	detector       *scanner.AgentDetector
	processManager *process.Manager
	healthChecker  *health.Checker
	stdout         io.Writer
	stderr         io.Writer
}

// NewApp creates and initializes the application
func NewApp() (*App, error) {
	if err := scanner.CheckPrereqs(); err != nil {
		return nil, err
	}

	config, err := models.GetConfigPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to get config paths: %w", err)
	}

	if err := config.EnsureDirs(); err != nil {
		return nil, fmt.Errorf("failed to create config directories: %w", err)
	}

	reg := registry.NewRegistry(config.RegistryFile)
	if err := reg.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load registry: %v\n", err)
	}
	warnLegacyCommandsOnce.Do(func() {
		warnLegacyManagedCommands(reg, os.Stderr)
	})

	return &App{
		config:         config,
		registry:       reg,
		scanner:        scanner.NewProcessScanner(),
		resolver:       scanner.NewProjectResolver(),
		detector:       scanner.NewAgentDetector(),
		processManager: process.NewManager(config.LogsDir),
		healthChecker:  health.NewChecker(0),
		stdout:         os.Stdout,
		stderr:         os.Stderr,
	}, nil
}

func (a *App) outWriter() io.Writer {
	if a != nil && a.stdout != nil {
		return a.stdout
	}
	return io.Discard
}

func (a *App) errWriter() io.Writer {
	if a != nil && a.stderr != nil {
		return a.stderr
	}
	return io.Discard
}

func (a *App) withOutput(stdout, stderr io.Writer) *App {
	if a == nil {
		return nil
	}
	clone := *a
	clone.stdout = stdout
	clone.stderr = stderr
	return &clone
}

// discoverServers combines scanning and detection into complete server info
func (a *App) discoverServers() ([]*models.ServerInfo, error) {
	processes, err := (&appDeps{app: a}).ScanProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed to scan processes: %w", err)
	}

	managedServices := a.registry.ListServices()
	for _, proc := range processes {
		if proc.CWD != "" {
			proc.ProjectRoot = a.resolver.FindProjectRoot(proc.CWD)
		}
		a.detector.EnrichProcessRecord(proc)
	}

	return a.buildServerInfos(processes, managedServices), nil
}

func (a *App) buildServerInfos(processes []*models.ProcessRecord, managedServices []*models.ManagedService) []*models.ServerInfo {
	commandMap := a.getCommandMap(processes)
	var servers []*models.ServerInfo

	matchedServices := make(map[*models.ManagedService]*models.ProcessRecord, len(managedServices))
	matchedProcesses := make(map[*models.ProcessRecord]*models.ManagedService, len(managedServices))
	reconciledServices := make(map[*models.ManagedService]lifecycle.ReconciledService, len(managedServices))
	for _, svc := range managedServices {
		reconciled := lifecycle.ReconcileWithResolver(svc, processes, managedServices, a.resolver.FindProjectRoot)
		reconciledServices[svc] = reconciled
		if reconciled.Status == string(models.StatusRunning) && reconciled.Verified && reconciled.Process != nil {
			matchedServices[svc] = reconciled.Process
			matchedProcesses[reconciled.Process] = svc
		}
	}

	for _, proc := range processes {
		if proc == nil {
			continue
		}

		matchedSvc := matchedProcesses[proc]
		if matchedSvc == nil && !scanner.IsDevProcess(proc, commandMap[proc.PID]) {
			continue
		}

		source := models.SourceManual
		if proc.AgentTag != nil {
			source = proc.AgentTag.Source
		}

		servers = append(servers, &models.ServerInfo{
			ManagedService: matchedSvc,
			ProcessRecord:  proc,
			Source:         source,
			Status:         "running",
		})
	}

	for _, svc := range managedServices {
		if matchedServices[svc] != nil {
			continue
		}

		reconciled := reconciledServices[svc]
		status := reconciled.Status
		if status == "" {
			status = string(models.StatusStopped)
		}
		crashReason := ""
		crashLogTail := []string(nil)
		if status == string(models.StatusCrashed) {
			crashReason, crashLogTail = a.getCrashReport(svc.Name, 12)
		}
		servers = append(servers, &models.ServerInfo{
			ManagedService: svc,
			Source:         models.SourceManaged,
			Status:         status,
			CrashReason:    crashReason,
			CrashLogTail:   crashLogTail,
		})
	}

	return servers
}

func (a *App) getCrashReport(serviceName string, lines int) (string, []string) {
	if lines <= 0 {
		lines = 12
	}
	logLines, err := a.processManager.Tail(serviceName, lines)
	if err != nil {
		return "No logs captured for last run", nil
	}
	reason := inferCrashReason(logLines)
	if reason == "" {
		reason = "Process exited unexpectedly (no explicit error line detected)"
	}
	return reason, logLines
}

func inferCrashReason(lines []string) string {
	keywords := []string{
		"panic",
		"fatal",
		"exception",
		"traceback",
		"error:",
		"eaddrinuse",
		"address already in use",
		"segmentation fault",
		"killed",
		"exit status",
	}

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				return line
			}
		}
	}

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}

	return ""
}

// getCommandMap creates a map of PID to command string
func (a *App) getCommandMap(processes []*models.ProcessRecord) map[int]string {
	cmdMap := make(map[int]string)
	for _, proc := range processes {
		if proc != nil {
			cmdMap[proc.PID] = proc.Command
		}
	}
	return cmdMap
}

func warnLegacyManagedCommands(reg *registry.Registry, out io.Writer) {
	if reg == nil || out == nil {
		return
	}
	services := reg.ListServices()
	if len(services) == 0 {
		return
	}

	var warnings []string
	for _, svc := range services {
		if svc == nil {
			continue
		}
		if p, ok := firstBlockedShellPattern(svc.Command); ok {
			warnings = append(warnings, fmt.Sprintf("  - %s (pattern %q)", svc.Name, p))
		}
	}
	if len(warnings) == 0 {
		return
	}
	sort.Strings(warnings)
	fmt.Fprintln(out, "Warning: legacy managed commands detected that rely on shell patterns.")
	fmt.Fprintln(out, "These commands may fail under strict execution. Update them to direct executable form.")
	for _, w := range warnings {
		fmt.Fprintln(out, w)
	}
}
