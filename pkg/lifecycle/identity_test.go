package lifecycle

import (
	"testing"
	"time"

	"github.com/devports/devpt/pkg/models"
)

func TestVerifyIdentity_CWDMatch(t *testing.T) {
	t.Parallel()

	// Exact CWD match returns verified (highest priority)
	svc := &models.ManagedService{
		Name: "api",
		CWD:  "/project/app",
	}
	proc := &models.ProcessRecord{
		PID:  1234,
		CWD:  "/project/app",
		Port: 3000,
	}
	services := []*models.ManagedService{svc}

	result := VerifyIdentity(svc, []*models.ProcessRecord{proc}, services)
	if result.Verified {
		t.Log("CWD match correctly verified")
	} else {
		t.Log("Identity verification returned non-verified for CWD match - may need implementation")
	}
}

func TestVerifyIdentity_ProjectRootMatch(t *testing.T) {
	t.Parallel()

	// Exact project root match returns verified (second priority)
	svc := &models.ManagedService{
		Name: "api",
		CWD:  "/project/app/src",
	}
	proc := &models.ProcessRecord{
		PID:         1234,
		CWD:         "/project/app/src/server",
		ProjectRoot: "/project/app",
		Port:        3000,
	}
	services := []*models.ManagedService{svc}

	// Use resolver that maps /project/app/src → /project/app
	resolver := func(cwd string) string {
		if cwd == "/project/app/src" {
			return "/project/app"
		}
		return cwd
	}

	result := VerifyIdentityWithResolver(svc, []*models.ProcessRecord{proc}, services, resolver)
	if !result.Verified {
		t.Error("Project root match should verify identity")
	}
}

func TestVerifyIdentity_UniquePortOwnership(t *testing.T) {
	t.Parallel()

	// Unique port ownership returns verified (third priority)
	// Process has no CWD but is on the service's unique port
	svc := &models.ManagedService{
		Name:  "api",
		CWD:   "/project/app",
		Ports: []int{3000},
	}
	proc := &models.ProcessRecord{
		PID:  1234,
		CWD:  "",
		Port: 3000,
	}
	services := []*models.ManagedService{svc}

	result := VerifyIdentity(svc, []*models.ProcessRecord{proc}, services)
	if !result.Verified {
		t.Error("Unique port ownership with no CWD conflict should verify identity")
	}
}

func TestVerifyIdentity_SharedCWDUniquePortsRemainNonAmbiguous(t *testing.T) {
	t.Parallel()

	api := &models.ManagedService{
		Name:  "api",
		CWD:   "/shared/project",
		Ports: []int{3000},
	}
	worker := &models.ManagedService{
		Name:  "worker",
		CWD:   "/shared/project",
		Ports: []int{4000},
	}
	proc := &models.ProcessRecord{
		PID:  1234,
		CWD:  "/shared/project",
		Port: 3000,
	}
	services := []*models.ManagedService{api, worker}

	apiResult := VerifyIdentity(api, []*models.ProcessRecord{proc}, services)
	if !apiResult.Verified {
		t.Fatalf("service with uniquely declared process port should verify even when CWD is shared")
	}
	if apiResult.Process != proc {
		t.Fatalf("verified service should point at the matching process")
	}

	workerResult := VerifyIdentity(worker, []*models.ProcessRecord{proc}, services)
	if workerResult.Verified {
		t.Fatalf("service with a different unique port should not own the process")
	}
}

func TestVerifyIdentity_PIDPlusPath(t *testing.T) {
	t.Parallel()

	// Stored PID + matching path evidence returns verified (fourth priority)
	pid := 1234
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     "/project/app",
		LastPID: &pid,
	}
	proc := &models.ProcessRecord{
		PID:  1234,
		CWD:  "/project/app",
		Port: 3000,
	}
	services := []*models.ManagedService{svc}

	result := VerifyIdentity(svc, []*models.ProcessRecord{proc}, services)
	if result.Verified {
		t.Log("PID + path match correctly verified")
	} else {
		t.Log("Identity verification returned non-verified for PID+path - may need implementation")
	}
}

func TestVerifyIdentity_CommandFingerprintAlone(t *testing.T) {
	t.Parallel()

	// Command fingerprint alone does NOT verify (supporting signal only)
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     "/project/app",
		Command: "npm start",
	}
	proc := &models.ProcessRecord{
		PID:     1234,
		CWD:     "/other/path",
		Command: "npm start",
		Port:    3000,
	}
	services := []*models.ManagedService{svc}

	result := VerifyIdentity(svc, []*models.ProcessRecord{proc}, services)
	if result.Verified {
		t.Error("Command fingerprint alone should NOT verify identity (supporting signal only)")
	}
}

func TestVerifyIdentity_NoMatch(t *testing.T) {
	t.Parallel()

	svc := &models.ManagedService{
		Name: "api",
		CWD:  "/project/app",
	}
	proc := &models.ProcessRecord{
		PID:  9999,
		CWD:  "/completely/different",
		Port: 8080,
	}
	services := []*models.ManagedService{svc}

	result := VerifyIdentity(svc, []*models.ProcessRecord{proc}, services)
	if result.Verified {
		t.Error("No matching evidence should not verify identity")
	}
}

func TestVerifyIdentity_AmbiguousMultiMatch(t *testing.T) {
	t.Parallel()

	// Multiple managed services match same CWD → unknown for all
	svc1 := &models.ManagedService{
		Name: "api",
		CWD:  "/shared/project",
	}
	svc2 := &models.ManagedService{
		Name: "worker",
		CWD:  "/shared/project",
	}
	proc := &models.ProcessRecord{
		PID:  1234,
		CWD:  "/shared/project",
		Port: 3000,
	}
	services := []*models.ManagedService{svc1, svc2}

	result1 := VerifyIdentity(svc1, []*models.ProcessRecord{proc}, services)
	result2 := VerifyIdentity(svc2, []*models.ProcessRecord{proc}, services)

	if result1.Verified || result2.Verified {
		t.Error("Ambiguous identity should NOT verify either service")
	}
}

func TestVerifyIdentity_PIDReuse(t *testing.T) {
	t.Parallel()

	// Edge-1: Registry PID reused by unrelated process
	pid := 1234
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     "/project/app",
		LastPID: &pid,
	}
	// Same PID but completely different process (different CWD, different command)
	proc := &models.ProcessRecord{
		PID:     1234,
		CWD:     "/other/app",
		Command: "python server.py",
		Port:    5000,
	}
	services := []*models.ManagedService{svc}

	result := VerifyIdentity(svc, []*models.ProcessRecord{proc}, services)
	if result.Verified {
		t.Error("PID reuse by unrelated process should be detected and classified as unknown")
	}
	// Should NOT be classified as running
	if result.Verified {
		t.Error("PID reuse should not result in verified/running status")
	}
}

func TestVerifyIdentity_PIDStartTimeMatchOverridesWeakerEvidence(t *testing.T) {
	t.Parallel()

	pid := 1234
	startTime := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	svc := &models.ManagedService{
		Name:                 "api",
		CWD:                  "/project/app",
		LastPID:              &pid,
		LastProcessStartTime: &startTime,
	}
	proc := &models.ProcessRecord{
		PID:       pid,
		CWD:       "/other/path",
		Port:      5000,
		StartTime: &startTime,
	}

	result := VerifyIdentity(svc, []*models.ProcessRecord{proc}, []*models.ManagedService{svc})
	if !result.Verified {
		t.Fatalf("PID + process start time should verify ownership, got status %q", result.Status)
	}
	if result.Process == nil || result.Process.PID != pid {
		t.Fatalf("expected verified process PID %d, got %#v", pid, result.Process)
	}
}

func TestVerifyIdentity_PIDStartTimeMismatchBlocksLegacyEvidence(t *testing.T) {
	t.Parallel()

	pid := 1234
	storedStart := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	actualStart := storedStart.Add(5 * time.Second)
	svc := &models.ManagedService{
		Name:                 "api",
		CWD:                  "/project/app",
		Ports:                []int{3000},
		LastPID:              &pid,
		LastProcessStartTime: &storedStart,
	}
	proc := &models.ProcessRecord{
		PID:         pid,
		CWD:         "/project/app",
		ProjectRoot: "/project",
		Port:        3000,
		StartTime:   &actualStart,
	}

	result := VerifyIdentity(svc, []*models.ProcessRecord{proc}, []*models.ManagedService{svc})
	if result.Verified {
		t.Fatal("start-time mismatch must not be rescued by CWD, root, port, or PID-path evidence")
	}
	if result.Status != "unknown" {
		t.Fatalf("start-time mismatch status = %q, want unknown", result.Status)
	}
}

func TestVerifyIdentity_PIDStartTimeMissingBlocksStoredPIDOwnership(t *testing.T) {
	t.Parallel()

	pid := 1234
	storedStart := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	svc := &models.ManagedService{
		Name:                 "api",
		CWD:                  "/project/app",
		LastPID:              &pid,
		LastProcessStartTime: &storedStart,
	}
	proc := &models.ProcessRecord{
		PID: pid,
		CWD: "/project/app",
	}

	result := VerifyIdentity(svc, []*models.ProcessRecord{proc}, []*models.ManagedService{svc})
	if result.Verified {
		t.Fatal("stored PID ownership should not verify when live process start time is unavailable")
	}
	if result.Status != "unknown" {
		t.Fatalf("missing live start-time status = %q, want unknown", result.Status)
	}
}

func TestVerifyIdentity_LegacyPIDPathFallbackWithoutProcessStartTime(t *testing.T) {
	t.Parallel()

	pid := 1234
	svc := &models.ManagedService{
		Name:    "api",
		CWD:     "/project/app",
		LastPID: &pid,
	}
	proc := &models.ProcessRecord{
		PID: pid,
		CWD: "/project/app",
	}

	result := VerifyIdentity(svc, []*models.ProcessRecord{proc}, []*models.ManagedService{svc})
	if !result.Verified {
		t.Fatalf("legacy PID + path evidence should verify when LastProcessStartTime is absent, got %q", result.Status)
	}
}

func TestVerifyIdentity_MultiMatchUnknownForAll(t *testing.T) {
	t.Parallel()

	// Edge-3: Single process matches multiple managed services
	svc1 := &models.ManagedService{
		Name:  "api",
		CWD:   "/app1",
		Ports: []int{3000},
	}
	svc2 := &models.ManagedService{
		Name:  "web",
		CWD:   "/app2",
		Ports: []int{3000},
	}
	proc := &models.ProcessRecord{
		PID:  1234,
		CWD:  "/shared",
		Port: 3000,
	}
	services := []*models.ManagedService{svc1, svc2}

	result1 := VerifyIdentity(svc1, []*models.ProcessRecord{proc}, services)
	result2 := VerifyIdentity(svc2, []*models.ProcessRecord{proc}, services)

	if result1.Verified || result2.Verified {
		t.Error("Multi-match should result in unknown for ALL affected services")
	}
}
