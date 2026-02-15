package cli

import "testing"

func TestValidateManagedCommand(t *testing.T) {
	t.Parallel()

	valid := []string{
		"npm run dev",
		"node server.js",
		"python3 -m uvicorn app:app --reload",
		"go run ./cmd/api",
	}
	for _, c := range valid {
		if err := validateManagedCommand(c); err != nil {
			t.Fatalf("expected valid command %q, got error: %v", c, err)
		}
	}

	invalid := []string{
		"",
		"npm run dev && echo ok",
		"npm run dev | tee out.log",
		"`whoami`",
		"echo ${HOME}",
	}
	for _, c := range invalid {
		if err := validateManagedCommand(c); err == nil {
			t.Fatalf("expected invalid command %q to fail", c)
		}
	}
}

func TestFirstBlockedShellPattern(t *testing.T) {
	t.Parallel()

	p, ok := firstBlockedShellPattern("npm run dev && echo ok")
	if !ok || p != "&&" {
		t.Fatalf("expected && pattern, got %q (ok=%v)", p, ok)
	}

	if p, ok = firstBlockedShellPattern("npm run dev"); ok || p != "" {
		t.Fatalf("expected no pattern, got %q (ok=%v)", p, ok)
	}
}
