package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/devports/devpt/pkg/models"
	"github.com/stretchr/testify/assert"
)

func managedSplitTestModel() *topModel {
	stoppedAt := time.Date(2026, 3, 27, 21, 54, 25, 0, time.UTC)
	deps := &fakeAppDeps{
		services: []*models.ManagedService{
			{
				Name:     "test-go-basic-fake",
				CWD:      "/Users/kirby/.config/dev-process-tracker/sandbox/servers/go-basic",
				Command:  "go run .",
				Ports:    []int{3401},
				LastStop: &stoppedAt,
			},
			{
				Name:    "docs-preview",
				CWD:     "/tmp/docs-preview",
				Command: "npm run dev",
				Ports:   []int{3001},
			},
		},
		servers: []*models.ServerInfo{
			{
				ManagedService: &models.ManagedService{Name: "test-go-basic-fake", CWD: "/Users/kirby/.config/dev-process-tracker/sandbox/servers/go-basic", Command: "go run .", Ports: []int{3401}},
				Status:         "crashed",
				Source:         models.SourceManaged,
				CrashReason:    "exit status 1",
				CrashLogTail: []string{
					"2026/03/27 21:54:25 [go-basic] listening on http://localhost:3400",
					"2026/03/27 21:54:25 listen tcp :3400: bind: address already in use",
					"exit status 1",
				},
			},
		},
		logPaths: map[string]string{
			"test-go-basic-fake": "~/.config/devpt/logs/test-go-basic-fake/2026-03-12T22-14-37.log",
		},
	}

	model := newTopModel(deps)
	model.width = 120
	model.height = 30
	model.mode = viewModeTable
	model.focus = focusManaged
	model.managedSel = 0
	return model
}

func TestManagedSplitView_SelectedServiceShowsDedicatedDetailsPane(t *testing.T) {
	model := managedSplitTestModel()
	// Services are sorted alphabetically, so test-go-basic-fake is at index 1
	model.managedSel = 1

	output := model.View().Content
	assert.Contains(t, output, "Managed Services")
	assert.Contains(t, output, "Selected service details")
	assert.Contains(t, output, "Headline: exit status 1")
	assert.Contains(t, output, "test-go-basic-fake")
}

func TestManagedSplitView_NoSelectionShowsPlaceholderPane(t *testing.T) {
	model := managedSplitTestModel()
	model.managedSel = -1

	output := model.View().Content
	assert.Contains(t, output, "Selected service details")
	assert.Contains(t, output, "Select a managed service to inspect status")
}

func TestManagedSplitView_StoppedServiceRemainsStopped(t *testing.T) {
	model := managedSplitTestModel()
	model.managedSel = 0

	output := model.View().Content
	assert.Contains(t, output, "docs-preview [stopped]")
	assert.NotContains(t, output, "docs-preview crashed")
}

func TestManagedSplitView_NarrowWidthPreservesPrimarySignals(t *testing.T) {
	model := managedSplitTestModel()
	model.width = 72
	model.managedSel = 1

	output := model.View().Content
	assert.Contains(t, output, "✘")
	assert.Contains(t, output, "exit status 1")
}

func TestManagedSplitView_SelectedManagedRowHighlightsWholeLine(t *testing.T) {
	model := managedSplitTestModel()
	model.managedSel = 0
	_ = model.View()

	var selectedLine string
	for _, line := range strings.Split(model.table.managedVP.View(), "\n") {
		if strings.Contains(ansi.Strip(line), "docs-preview [stopped]") {
			selectedLine = line
			break
		}
	}

	assert.NotEmpty(t, selectedLine)
	assert.Contains(t, selectedLine, "48;5;57")
	assert.NotContains(t, selectedLine, "\x1b[m docs-preview")
}
