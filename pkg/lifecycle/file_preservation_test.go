package lifecycle

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLifecycleFilePreservation(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve lifecycle package directory")
	}
	dir := filepath.Dir(currentFile)

	requiredFiles := []string{
		"identity.go",
		"identity_test.go",
		"start.go",
		"start_test.go",
		"stop.go",
		"stop_test.go",
		"restart.go",
		"restart_test.go",
		"reconciler.go",
		"reconciler_test.go",
		"manager.go",
		"manager_test.go",
	}

	for _, name := range requiredFiles {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			info, err := os.Stat(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("required lifecycle file is missing: %s", name)
			}
			if info.IsDir() {
				t.Fatalf("required lifecycle file path is a directory: %s", name)
			}
		})
	}
}
