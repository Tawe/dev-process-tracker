package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// renderRow is a tiny helper so tests read the styled string only.
func renderRow(state string) string {
	row, _ := renderDetailsActions(state)
	return row
}

func TestRenderDetailsActionsRunningShowsAllThree(t *testing.T) {
	t.Parallel()
	plain := ansi.Strip(renderRow("running"))
	for _, want := range []string{restartIcon, "restart", stopIcon, "stop", editIcon, "edit"} {
		assert.True(t, strings.Contains(plain, want), "running row missing %q in: %q", want, plain)
	}
	assert.NotContains(t, plain, startIcon, "running row should not show start")
}

func TestRenderDetailsActionsStoppedShowsStartAndEdit(t *testing.T) {
	t.Parallel()
	for _, state := range []string{"stopped", "crashed"} {
		plain := ansi.Strip(renderRow(state))
		assert.Contains(t, plain, startIcon, "state %q should show start: %q", state, plain)
		assert.Contains(t, plain, "edit", "state %q should show edit: %q", state, plain)
		assert.NotContains(t, plain, restartIcon, "state %q should not show restart", state)
		assert.NotContains(t, plain, stopIcon, "state %q should not show stop", state)
	}
}

func TestRenderDetailsActionsStartingShowsEditOnly(t *testing.T) {
	t.Parallel()
	plain := ansi.Strip(renderRow("starting"))
	assert.Contains(t, plain, editIcon)
	assert.NotContains(t, plain, startIcon)
	assert.NotContains(t, plain, restartIcon)
	assert.NotContains(t, plain, stopIcon)
}

// edit is stable on the left; then re/start; then stop. Order is part of the
// spec (edit must not jump position between states).
func TestUAT_DetailsActionRowEditIsLeftmost(t *testing.T) {
	t.Parallel()
	running := ansi.Strip(renderRow("running"))
	assert.Less(t, strings.Index(running, "edit"), strings.Index(running, "restart"))
	assert.Less(t, strings.Index(running, "restart"), strings.Index(running, "stop"))

	stopped := ansi.Strip(renderRow("stopped"))
	assert.Less(t, strings.Index(stopped, "edit"), strings.Index(stopped, "start"))
}

// Whole buttons are colored by action (not gray labels): edit=cyan, start/
// restart=green, stop=red. The label text carries the action's color code.
func TestUAT_DetailsActionButtonsAreColored(t *testing.T) {
	t.Parallel()
	stopped := renderRow("stopped")
	assert.Regexp(t, "\x1b\\[1;36m.*edit", stopped, "edit should be cyan")
	assert.Regexp(t, "\x1b\\[1;32m.*start", stopped, "start should be green")

	running := renderRow("running")
	assert.Regexp(t, "\x1b\\[1;32m.*restart", running, "restart should be green")
	assert.Regexp(t, "\x1b\\[1;31m.*stop", running, "stop should be red")
}

// Labels survive the pane-width fit (regression guard: fitLine used to count
// escape bytes as width and truncate the trailing label).
func TestUAT_DetailsActionRowLabelsSurviveFit(t *testing.T) {
	t.Parallel()
	for _, state := range []string{"running", "stopped", "crashed"} {
		plain := ansi.Strip(fitAnsiLine(renderRow(state), 56))
		assert.Contains(t, plain, "edit", "state %q: edit label truncated at width 56", state)
		if state == "running" {
			assert.Contains(t, plain, "restart")
			assert.Contains(t, plain, "stop")
		} else {
			assert.Contains(t, plain, "start")
		}
	}
}

// Hit regions are reported and ordered left-to-right matching the rendered row.
func TestRenderDetailsActionsReportsHitRegions(t *testing.T) {
	t.Parallel()
	_, hits := renderDetailsActions("running")
	assert.Equal(t, "edit", hits[0].label)
	assert.Equal(t, "restart", hits[1].label)
	assert.Equal(t, "stop", hits[2].label)
	// non-overlapping, ascending, each button wide enough for icon+space+label
	for _, h := range hits {
		assert.Greater(t, h.x1, h.x0, "%s has empty range", h.label)
	}
	assert.LessOrEqual(t, hits[0].x1, hits[1].x0, "edit must end before restart starts")
	assert.LessOrEqual(t, hits[1].x1, hits[2].x0, "restart must end before stop starts")
}
