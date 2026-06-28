package models

import "testing"

func TestValidateManagedServiceFields(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, svc, cmd string
		wantErr        bool
	}{
		{"valid", "api", "npm run dev", false},
		{"empty name", "  ", "npm run dev", true},
		{"empty command", "api", "   ", true},
		{"shell pattern ampersand", "api", "a && b", true},
		{"shell pattern pipe", "api", "cat x | grep y", true},
		{"subshell", "api", "$(whoami)", true},
	}
	for _, c := range cases {
		err := ValidateManagedServiceFields(c.svc, c.cmd)
		if c.wantErr && err == nil {
			t.Errorf("%s: expected error, got nil", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: unexpected error: %v", c.name, err)
		}
	}
}

func TestFirstBlockedShellPattern(t *testing.T) {
	t.Parallel()
	if p, ok := FirstBlockedShellPattern("npm run dev && echo hi"); !ok || p != "&&" {
		t.Fatalf("expected && pattern, got %q ok=%v", p, ok)
	}
	if _, ok := FirstBlockedShellPattern("npm run dev"); ok {
		t.Fatal("clean command should report no pattern")
	}
}
