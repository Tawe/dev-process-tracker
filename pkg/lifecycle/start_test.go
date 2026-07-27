package lifecycle

import (
	"fmt"
	"testing"
	"time"

	"github.com/devports/devpt/pkg/models"
)

// mockDeps implements Deps for testing.
type mockDeps struct {
	services     map[string]*models.ManagedService
	processes    []*models.ProcessRecord
	runningPIDs  map[int]bool
	startTimes   map[int]time.Time
	nextPID      int
	healthPorts  map[int]bool
	logTail      []string
	locked       map[string]bool
	projectRoots map[string]string
	updateErr    error
	clearErr     error
	scanErr      error
	startErr     error
	startFn      func(svc *models.ManagedService) (int, error)
	startTimeErr error
	stopErr      error
	crashOnStart bool // if true, started process is not running
}

func newMockDeps() *mockDeps {
	return &mockDeps{
		services:     make(map[string]*models.ManagedService),
		runningPIDs:  make(map[int]bool),
		startTimes:   make(map[int]time.Time),
		healthPorts:  make(map[int]bool),
		locked:       make(map[string]bool),
		projectRoots: make(map[string]string),
		nextPID:      50000,
	}
}

func (m *mockDeps) GetService(name string) *models.ManagedService {
	return m.services[name]
}

func (m *mockDeps) UpdateServicePID(name string, pid int) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if svc, ok := m.services[name]; ok {
		svc.LastPID = &pid
		svc.LastProcessStartTime = nil
	}
	return nil
}

func (m *mockDeps) UpdateServiceProcessIdentity(name string, pid int, processStartTime time.Time) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if svc, ok := m.services[name]; ok {
		svc.LastPID = &pid
		t := processStartTime
		svc.LastProcessStartTime = &t
	}
	return nil
}

func (m *mockDeps) ClearServicePID(name string) error {
	if m.clearErr != nil {
		return m.clearErr
	}
	if svc, ok := m.services[name]; ok {
		svc.LastPID = nil
		svc.LastProcessStartTime = nil
	}
	return nil
}

func (m *mockDeps) StartProcess(svc *models.ManagedService) (int, error) {
	if m.startFn != nil {
		return m.startFn(svc)
	}
	if m.startErr != nil {
		return 0, m.startErr
	}
	pid := m.nextPID
	m.nextPID++
	if !m.crashOnStart {
		m.runningPIDs[pid] = true
		m.startTimes[pid] = time.Date(2026, time.May, 27, 12, 0, pid%60, 0, time.UTC)
	}
	return pid, nil
}

func (m *mockDeps) StopProcess(pid int) error {
	delete(m.runningPIDs, pid)
	delete(m.startTimes, pid)
	return m.stopErr
}

func (m *mockDeps) IsRunning(pid int) bool {
	return m.runningPIDs[pid]
}

func (m *mockDeps) GetProcessStartTime(pid int) (time.Time, error) {
	if m.startTimeErr != nil {
		return time.Time{}, m.startTimeErr
	}
	if t, ok := m.startTimes[pid]; ok {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("process start time unavailable")
}

func (m *mockDeps) GetProcessCommand(pid int) (string, error) {
	return "mock-command", nil
}

func (m *mockDeps) UpdateServiceResolvedCommand(name, resolvedCommand string) error {
	if svc, ok := m.services[name]; ok {
		svc.ResolvedCommand = resolvedCommand
	}
	return nil
}

func (m *mockDeps) ScanProcesses() ([]*models.ProcessRecord, error) {
	if m.scanErr != nil {
		return nil, m.scanErr
	}
	return m.processes, nil
}

func (m *mockDeps) ListServices() []*models.ManagedService {
	var svcs []*models.ManagedService
	for _, svc := range m.services {
		svcs = append(svcs, svc)
	}
	return svcs
}

func (m *mockDeps) CheckHealth(port int) bool {
	return m.healthPorts[port]
}

func (m *mockDeps) GetLogTail(name string, lines int) []string {
	return m.logTail
}

func (m *mockDeps) AcquireLock(serviceName string) error {
	if m.locked[serviceName] {
		return ErrLockBlocked
	}
	m.locked[serviceName] = true
	return nil
}

func (m *mockDeps) ReleaseLock(serviceName string) {
	delete(m.locked, serviceName)
}

func (m *mockDeps) ResolveProjectRoot(cwd string) string {
	if r, ok := m.projectRoots[cwd]; ok {
		return r
	}
	return cwd
}

func TestStart_AlreadyRunning(t *testing.T) {
	t.Parallel()

	svc := &models.ManagedService{Name: "api", CWD: "/project", Ports: []int{3000}}
	proc := &models.ProcessRecord{PID: 1234, CWD: "/project", Port: 3000}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{proc}
	deps.runningPIDs[1234] = true

	result := StartService(deps, svc)
	if result.Outcome != OutcomeNoop {
		t.Errorf("already running should return noop, got %q", result.Outcome)
	}
	if result.PID != 1234 {
		t.Errorf("noop should include running PID, got %d", result.PID)
	}
}

func TestStart_AmbiguousIdentity(t *testing.T) {
	t.Parallel()

	svc1 := &models.ManagedService{Name: "api", CWD: "/shared"}
	svc2 := &models.ManagedService{Name: "worker", CWD: "/shared"}
	proc := &models.ProcessRecord{PID: 1234, CWD: "/shared", Port: 3000}

	deps := newMockDeps()
	deps.services["api"] = svc1
	deps.services["worker"] = svc2
	deps.processes = []*models.ProcessRecord{proc}
	deps.runningPIDs[1234] = true

	result := StartService(deps, svc1)
	if result.Outcome != OutcomeBlocked {
		t.Errorf("ambiguous identity should return blocked, got %q", result.Outcome)
	}
}

func TestStart_PreflightInvalid_MissingCWD(t *testing.T) {
	t.Parallel()

	svc := &models.ManagedService{
		Name:    "api",
		CWD:     "/nonexistent/path/that/does/not/exist",
		Command: "npm start",
	}

	deps := newMockDeps()
	deps.services["api"] = svc

	result := StartService(deps, svc)
	if result.Outcome != OutcomeInvalid {
		t.Errorf("missing CWD should return invalid, got %q", result.Outcome)
	}
}

func TestStart_PortConflict(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     tmpDir,
		Command: "npm start",
		Ports:   []int{3000},
	}

	existingProc := &models.ProcessRecord{PID: 9999, CWD: "/other", Port: 3000, Command: "python"}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{existingProc}
	deps.runningPIDs[9999] = true

	result := StartService(deps, svc)
	if result.Outcome != OutcomeBlocked {
		t.Errorf("port conflict should return blocked, got %q", result.Outcome)
	}
	if result.Message == "" {
		t.Error("blocked result should have a message")
	}
}

func TestStart_StaleRegistry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pid := 9999
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     tmpDir,
		Command: "echo hi",
		LastPID: &pid,
		Readiness: &models.ReadinessConfig{
			Mode:    models.ReadinessProcessOnly,
			Timeout: 1,
		},
	}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{}

	result := StartService(deps, svc)
	// Stale PID means crashed status, then should attempt fresh start
	if result.Outcome == OutcomeNoop {
		t.Error("stale PID should not cause noop - should attempt fresh start")
	}
}

func TestStart_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     tmpDir,
		Command: "echo hi",
		Readiness: &models.ReadinessConfig{
			Mode:    models.ReadinessProcessOnly,
			Timeout: 1,
		},
	}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{}

	result := StartService(deps, svc)
	if result.Outcome != OutcomeSuccess {
		t.Errorf("expected success, got %q: %s", result.Outcome, result.Message)
	}
	if result.PID == 0 {
		t.Error("success should include PID")
	}
	if svc.LastPID == nil || *svc.LastPID != result.PID {
		t.Fatalf("success should persist LastPID %d, got %v", result.PID, svc.LastPID)
	}
	if svc.LastProcessStartTime == nil {
		t.Fatal("success should persist process start time")
	}
}

func TestStart_ProcessStartTimeUnavailableFailsWithoutPersisting(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     tmpDir,
		Command: "echo hi",
		Readiness: &models.ReadinessConfig{
			Mode:    models.ReadinessProcessOnly,
			Timeout: 1,
		},
	}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{}
	deps.startTimeErr = fmt.Errorf("unsupported platform")

	result := StartService(deps, svc)
	if result.Outcome != OutcomeFailed {
		t.Fatalf("start-time lookup failure should return failed, got %q: %s", result.Outcome, result.Message)
	}
	if svc.LastPID != nil {
		t.Fatalf("failed identity confirmation should not persist LastPID, got %d", *svc.LastPID)
	}
	if svc.LastProcessStartTime != nil {
		t.Fatal("failed identity confirmation should not persist process start time")
	}
	if deps.IsRunning(result.PID) {
		t.Fatalf("failed identity confirmation should stop PID %d", result.PID)
	}
}

func TestStart_ReadinessTimeout(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     tmpDir,
		Command: "sleep 100",
		Ports:   []int{3000},
		Readiness: &models.ReadinessConfig{
			Mode:    models.ReadinessPortBound,
			Timeout: 1,
		},
	}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{}

	result := StartService(deps, svc)
	if result.Outcome == OutcomeSuccess {
		t.Error("readiness timeout should not return success")
	}
	if result.Outcome == OutcomeFailed {
		t.Logf("Readiness timeout correctly reported failure: %s", result.Message)
	}
}

func TestStart_ReadinessTimeoutPreservesLogDiagnostics(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     tmpDir,
		Command: "sleep 100",
		Ports:   []int{3000},
		Readiness: &models.ReadinessConfig{
			Mode:    models.ReadinessPortBound,
			Timeout: 1,
		},
	}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{}
	deps.logTail = []string{"listening check timed out", "last known log line"}

	result := StartService(deps, svc)
	if result.Outcome != OutcomeFailed {
		t.Fatalf("readiness timeout should fail, got %q: %s", result.Outcome, result.Message)
	}
	if len(result.Diagnostics) != len(deps.logTail) {
		t.Fatalf("expected diagnostics to include log tail, got %#v", result.Diagnostics)
	}
	for i, line := range deps.logTail {
		if result.Diagnostics[i] != line {
			t.Fatalf("diagnostic line %d = %q, want %q", i, result.Diagnostics[i], line)
		}
	}
}

func TestStart_NoUnconfirmedPID(t *testing.T) {
	t.Parallel()

	svc := &models.ManagedService{
		Name:    "api",
		CWD:     "/nonexistent",
		Command: "npm start",
	}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{}

	result := StartService(deps, svc)
	if result.Outcome == OutcomeFailed || result.Outcome == OutcomeInvalid {
		if result.PID != 0 {
			t.Error("failed/invalid start should not report a PID")
		}
	}
}

func TestStart_LockContention(t *testing.T) {
	t.Parallel()

	svc := &models.ManagedService{Name: "api", CWD: "/project", Command: "echo hi"}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{}
	deps.locked["api"] = true

	result := StartService(deps, svc)
	if result.Outcome != OutcomeBlocked {
		t.Errorf("lock contention should return blocked, got %q", result.Outcome)
	}
}

func TestStart_NilDeps(t *testing.T) {
	t.Parallel()

	result := StartService(nil, &models.ManagedService{Name: "api"})
	if result.Outcome != OutcomeInvalid {
		t.Errorf("nil deps should return invalid, got %q", result.Outcome)
	}
}

func TestStart_NilService(t *testing.T) {
	t.Parallel()

	deps := newMockDeps()
	result := StartService(deps, nil)
	if result.Outcome != OutcomeInvalid {
		t.Errorf("nil service should return invalid, got %q", result.Outcome)
	}
}

func TestStart_PreflightEmptyCommand(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     tmpDir,
		Command: "",
	}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{}

	result := StartService(deps, svc)
	if result.Outcome != OutcomeInvalid {
		t.Errorf("empty command should return invalid, got %q", result.Outcome)
	}
}

func TestStart_CrashImmediately(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     tmpDir,
		Command: "exit 1",
		Readiness: &models.ReadinessConfig{
			Mode:    models.ReadinessProcessOnly,
			Timeout: 1,
		},
	}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{}
	deps.crashOnStart = true

	result := StartService(deps, svc)
	if result.Outcome == OutcomeSuccess {
		t.Error("crashed process should not return success")
	}
}

func TestStart_MessageFormat(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     tmpDir,
		Command: "echo hi",
		Readiness: &models.ReadinessConfig{
			Mode:    models.ReadinessProcessOnly,
			Timeout: 1,
		},
	}

	deps := newMockDeps()
	deps.services["api"] = svc
	deps.processes = []*models.ProcessRecord{}

	result := StartService(deps, svc)
	if result.Outcome == OutcomeSuccess {
		if result.Message == "" {
			t.Error("success result should have a message")
		}
		_ = fmt.Sprintf("Message: %s", result.Message)
	}
}
