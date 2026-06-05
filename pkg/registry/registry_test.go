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
