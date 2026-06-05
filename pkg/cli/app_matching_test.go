package cli

import (
	"testing"

	"github.com/devports/devpt/pkg/models"
)

func TestBuildServerInfosKeepsManagedNonDevProcess(t *testing.T) {
	t.Parallel()

	app, _, _ := newTestApp(t)
	lastPID := 1234
	svc := &models.ManagedService{
		Name:    "postgres",
		CWD:     "/workspace/db",
		Ports:   []int{5432},
		LastPID: &lastPID,
	}
	proc := &models.ProcessRecord{
		PID:         1234,
		Port:        5432,
		Command:     "/usr/local/bin/postgres",
		CWD:         "/workspace/db",
		ProjectRoot: "/workspace/db",
	}

	servers := app.buildServerInfos([]*models.ProcessRecord{proc}, []*models.ManagedService{svc})
	got := findServerForManagedService(servers, svc)

	if got == nil {
		t.Fatal("expected managed service to be listed")
	}
	if got.ProcessRecord != proc {
		t.Fatalf("expected managed process match, got %#v", got.ProcessRecord)
	}
	if got.Status != string(models.StatusRunning) {
		t.Fatalf("expected running managed status, got %q", got.Status)
	}
}

func TestBuildServerInfosRejectsPIDOnlyMatch(t *testing.T) {
	t.Parallel()

	app, _, _ := newTestApp(t)
	lastPID := 4242
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     "/workspace/api",
		Ports:   []int{3000},
		LastPID: &lastPID,
	}
	proc := &models.ProcessRecord{
		PID:         4242,
		Port:        9999,
		Command:     "/usr/sbin/unrelated",
		CWD:         "/tmp/other",
		ProjectRoot: "/tmp/other",
	}

	servers := app.buildServerInfos([]*models.ProcessRecord{proc}, []*models.ManagedService{svc})
	got := findServerForManagedService(servers, svc)

	if got == nil {
		t.Fatal("expected managed service to be listed")
	}
	if got.ProcessRecord != nil {
		t.Fatalf("expected PID-only candidate to be rejected, got %#v", got.ProcessRecord)
	}
	if got.Status != string(models.StatusCrashed) {
		t.Fatalf("expected stale PID to be reported as crashed, got %q", got.Status)
	}
}

func findServerForManagedService(servers []*models.ServerInfo, svc *models.ManagedService) *models.ServerInfo {
	for _, srv := range servers {
		if srv.ManagedService == svc {
			return srv
		}
	}
	return nil
}
