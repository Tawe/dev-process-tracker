package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/devports/devpt/pkg/models"
	"github.com/devports/devpt/pkg/registry"
)

func TestWarnLegacyManagedCommands(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	regPath := filepath.Join(tmp, "registry.json")
	reg := registry.NewRegistry(regPath)
	if err := reg.Load(); err != nil {
		t.Fatalf("load registry: %v", err)
	}

	now := time.Now()
	if err := reg.AddService(&models.ManagedService{
		Name:      "safe",
		CWD:       tmp,
		Command:   "npm run dev",
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("add safe service: %v", err)
	}
	if err := reg.AddService(&models.ManagedService{
		Name:      "legacy",
		CWD:       tmp,
		Command:   "npm run dev && echo ok",
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("add legacy service: %v", err)
	}

	var out bytes.Buffer
	warnLegacyManagedCommands(reg, &out)
	s := out.String()
	if !strings.Contains(s, "legacy") {
		t.Fatalf("expected warning to include legacy service, got: %q", s)
	}
	if !strings.Contains(s, "pattern \"&&\"") {
		t.Fatalf("expected warning to include pattern info, got: %q", s)
	}
}
