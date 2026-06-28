package tui

import (
	"strings"
	"testing"
)

func TestRenderDetailsActionsRunningShowsAllThree(t *testing.T) {
	t.Parallel()
	out := renderDetailsActions("running")
	for _, want := range []string{restartIcon, "restart", stopIcon, "stop", editIcon, "edit"} {
		if !strings.Contains(out, want) {
			t.Errorf("running row missing %q in: %q", want, out)
		}
	}
}

func TestRenderDetailsActionsNonRunningShowsEditOnly(t *testing.T) {
	t.Parallel()
	for _, state := range []string{"stopped", "crashed", "starting", ""} {
		out := renderDetailsActions(state)
		if !strings.Contains(out, editIcon) {
			t.Errorf("state %q should still show edit: %q", state, out)
		}
		if strings.Contains(out, restartIcon) || strings.Contains(out, stopIcon) {
			t.Errorf("state %q should not show restart/stop: %q", state, out)
		}
	}
}
