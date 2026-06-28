package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/devports/devpt/pkg/models"
)

func TestRegistryUpdateServiceProcessIdentityPersistsPIDAndStartTime(t *testing.T) {
	t.Parallel()

	reg := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	svc := &models.ManagedService{Name: "api", CWD: "/tmp", Command: "sleep 10"}
	if err := reg.AddService(svc); err != nil {
		t.Fatalf("AddService: %v", err)
	}

	processStart := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	if err := reg.UpdateServiceProcessIdentity("api", 1234, processStart); err != nil {
		t.Fatalf("UpdateServiceProcessIdentity: %v", err)
	}

	got := reg.GetService("api")
	if got.LastPID == nil || *got.LastPID != 1234 {
		t.Fatalf("LastPID = %#v, want 1234", got.LastPID)
	}
	if got.LastProcessStartTime == nil || !got.LastProcessStartTime.Equal(processStart) {
		t.Fatalf("LastProcessStartTime = %#v, want %v", got.LastProcessStartTime, processStart)
	}
	if got.LastStart == nil {
		t.Fatal("LastStart should still record lifecycle event time")
	}
	if got.LastStop != nil {
		t.Fatalf("LastStop = %#v, want nil", got.LastStop)
	}

	raw, err := os.ReadFile(reg.FilePath())
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	var persisted models.Registry
	if err := json.Unmarshal(raw, &persisted); err != nil {
		t.Fatalf("unmarshal registry: %v", err)
	}
	persistedSvc := persisted.Services["api"]
	if persistedSvc.LastPID == nil || *persistedSvc.LastPID != 1234 {
		t.Fatalf("persisted LastPID = %#v, want 1234", persistedSvc.LastPID)
	}
	if persistedSvc.LastProcessStartTime == nil || !persistedSvc.LastProcessStartTime.Equal(processStart) {
		t.Fatalf("persisted LastProcessStartTime = %#v, want %v", persistedSvc.LastProcessStartTime, processStart)
	}
}

func TestRegistryClearServicePIDClearsProcessStartTime(t *testing.T) {
	t.Parallel()

	reg := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	svc := &models.ManagedService{Name: "api", CWD: "/tmp", Command: "sleep 10"}
	if err := reg.AddService(svc); err != nil {
		t.Fatalf("AddService: %v", err)
	}
	processStart := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	if err := reg.UpdateServiceProcessIdentity("api", 1234, processStart); err != nil {
		t.Fatalf("UpdateServiceProcessIdentity: %v", err)
	}

	if err := reg.ClearServicePID("api"); err != nil {
		t.Fatalf("ClearServicePID: %v", err)
	}

	got := reg.GetService("api")
	if got.LastPID != nil {
		t.Fatalf("LastPID = %#v, want nil", got.LastPID)
	}
	if got.LastProcessStartTime != nil {
		t.Fatalf("LastProcessStartTime = %#v, want nil", got.LastProcessStartTime)
	}
	if got.LastStop == nil {
		t.Fatal("LastStop should be set")
	}
}

func TestRegistryUpdateServicePIDClearsExistingProcessStartTime(t *testing.T) {
	t.Parallel()

	reg := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	svc := &models.ManagedService{Name: "api", CWD: "/tmp", Command: "sleep 10"}
	if err := reg.AddService(svc); err != nil {
		t.Fatalf("AddService: %v", err)
	}
	processStart := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	if err := reg.UpdateServiceProcessIdentity("api", 1234, processStart); err != nil {
		t.Fatalf("UpdateServiceProcessIdentity: %v", err)
	}

	if err := reg.UpdateServicePID("api", 5678); err != nil {
		t.Fatalf("UpdateServicePID: %v", err)
	}

	got := reg.GetService("api")
	if got.LastPID == nil || *got.LastPID != 5678 {
		t.Fatalf("LastPID = %#v, want 5678", got.LastPID)
	}
	if got.LastProcessStartTime != nil {
		t.Fatalf("legacy UpdateServicePID should clear LastProcessStartTime, got %#v", got.LastProcessStartTime)
	}
}

// addServiceWithRuntime seeds a registry with a service that has runtime/
// identity fields populated, so preservation can be asserted after rename/upsert.
func addServiceWithRuntime(t *testing.T, reg *Registry, name string) {
	t.Helper()
	svc := &models.ManagedService{Name: name, CWD: "/old", Command: "sleep 10", Ports: []int{3000}}
	if err := reg.AddService(svc); err != nil {
		t.Fatalf("AddService: %v", err)
	}
	start := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	if err := reg.UpdateServiceProcessIdentity(name, 4242, start); err != nil {
		t.Fatalf("UpdateServiceProcessIdentity: %v", err)
	}
	if err := reg.UpdateServiceResolvedCommand(name, "/usr/bin/sleep"); err != nil {
		t.Fatalf("UpdateServiceResolvedCommand: %v", err)
	}
}

func TestRenameServicePreservesRuntimeFields(t *testing.T) {
	t.Parallel()

	reg := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	addServiceWithRuntime(t, reg, "api")
	before := reg.GetService("api")
	beforeStart := before.LastProcessStartTime
	beforeCreated := before.CreatedAt
	beforeUpdated := before.UpdatedAt

	if err := reg.RenameService("api", "api-v2"); err != nil {
		t.Fatalf("RenameService: %v", err)
	}

	if reg.GetService("api") != nil {
		t.Fatal("old key should be gone after rename")
	}
	got := reg.GetService("api-v2")
	if got == nil {
		t.Fatal("renamed service missing")
	}
	if got.LastPID == nil || *got.LastPID != 4242 {
		t.Fatalf("LastPID = %#v, want 4242", got.LastPID)
	}
	if got.ResolvedCommand != "/usr/bin/sleep" {
		t.Fatalf("ResolvedCommand = %q, want /usr/bin/sleep", got.ResolvedCommand)
	}
	if got.LastProcessStartTime == nil || !got.LastProcessStartTime.Equal(*beforeStart) {
		t.Fatalf("LastProcessStartTime not preserved")
	}
	if got.CreatedAt != beforeCreated {
		t.Fatalf("CreatedAt changed: got %v want %v", got.CreatedAt, beforeCreated)
	}
	if !got.UpdatedAt.After(beforeUpdated) {
		t.Fatal("UpdatedAt should advance")
	}

	// Persists to disk.
	raw, err := os.ReadFile(reg.FilePath())
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	var persisted models.Registry
	if err := json.Unmarshal(raw, &persisted); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := persisted.Services["api"]; ok {
		t.Fatal("old key persisted")
	}
	if persisted.Services["api-v2"] == nil {
		t.Fatal("renamed key not persisted")
	}
}

func TestRenameServiceRejectsSameName(t *testing.T) {
	t.Parallel()
	reg := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	addServiceWithRuntime(t, reg, "api")
	if err := reg.RenameService("api", "api"); err == nil {
		t.Fatal("rename to same name should error")
	}
}

func TestRenameServiceRejectsExistingTarget(t *testing.T) {
	t.Parallel()
	reg := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	addServiceWithRuntime(t, reg, "api")
	addServiceWithRuntime(t, reg, "web")
	if err := reg.RenameService("api", "web"); err == nil {
		t.Fatal("rename to existing target should error")
	}
	if reg.GetService("api") == nil {
		t.Fatal("source should remain after rejected rename")
	}
}

func TestRenameServiceMissingSource(t *testing.T) {
	t.Parallel()
	reg := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	if err := reg.RenameService("nope", "yep"); err == nil {
		t.Fatal("rename of missing source should error")
	}
}

func TestUpsertServiceForceOverwritesPreservingRuntime(t *testing.T) {
	t.Parallel()
	reg := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	addServiceWithRuntime(t, reg, "api")
	before := reg.GetService("api")

	if err := reg.UpsertService(&models.ManagedService{Name: "api", CWD: "/new", Command: "npm run dev", Ports: []int{8080}}); err != nil {
		t.Fatalf("UpsertService: %v", err)
	}
	got := reg.GetService("api")
	if got.CWD != "/new" || got.Command != "npm run dev" || len(got.Ports) != 1 || got.Ports[0] != 8080 {
		t.Fatalf("editable fields not overwritten: %+v", got)
	}
	if got.LastPID == nil || *got.LastPID != 4242 {
		t.Fatalf("LastPID not preserved: %#v", got.LastPID)
	}
	if got.ResolvedCommand != "/usr/bin/sleep" {
		t.Fatalf("ResolvedCommand not preserved: %q", got.ResolvedCommand)
	}
	if got.CreatedAt != before.CreatedAt {
		t.Fatalf("CreatedAt should be preserved on overwrite")
	}
}

func TestUpsertServiceForceCreatesWhenAbsent(t *testing.T) {
	t.Parallel()
	reg := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	if err := reg.UpsertService(&models.ManagedService{Name: "brand", CWD: "/p", Command: "go run .", Ports: []int{9000}}); err != nil {
		t.Fatalf("UpsertService: %v", err)
	}
	got := reg.GetService("brand")
	if got == nil {
		t.Fatal("absent service should be created")
	}
	if got.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should be set on create")
	}
}

func TestAddServiceStillRejectsExisting(t *testing.T) {
	t.Parallel()
	reg := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	addServiceWithRuntime(t, reg, "api")
	if err := reg.AddService(&models.ManagedService{Name: "api", CWD: "/x", Command: "x"}); err == nil {
		t.Fatal("AddService should still reject an existing name")
	}
}
