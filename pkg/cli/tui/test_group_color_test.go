package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/devports/devpt/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestGroupConfirmSelectedRowColor(t *testing.T) {
	t.Parallel()

	t.Run("managed list: selected row amber, other group members dimmed orange", func(t *testing.T) {
		deps := &fakeAppDeps{
			servers: []*models.ServerInfo{
				makeRunningServer("api-gateway", 1001, 3000),
				makeRunningServer("api-auth", 1002, 3001),
			},
			services: []*models.ManagedService{
				{Name: "api-gateway", CWD: "/tmp/api-gateway", Command: "node server.js", Ports: []int{3000}},
				{Name: "api-auth", CWD: "/tmp/api-auth", Command: "node auth.js", Ports: []int{3001}},
			},
		}
		m := newTopModel(deps)
		m.mode = viewModeTable
		m.focus = focusManaged
		m.managedSel = 0
		m.width = 120
		m.height = 30

		m.Update(tea.KeyPressMsg{Code: 'g'})
		m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
		assert.NotNil(t, m.confirm)

		content := m.renderManagedList(60, m.managedServices())
		lines := strings.Split(content, "\n")

		// Selected row (managedSel=0, alphabetically api-auth) → amber
		assert.Contains(t, lines[0], "48;5;178", "selected row in group confirm should be amber")
		// Other group member → dimmed orange
		assert.Contains(t, lines[1], "48;5;94", "other group member should be dimmed orange")
	})

	t.Run("running table: selected row amber, other group members dimmed orange", func(t *testing.T) {
		deps := &fakeAppDeps{
			servers: []*models.ServerInfo{
				makeRunningServer("api-gateway", 1001, 3000),
				makeRunningServer("api-auth", 1002, 3001),
			},
			services: []*models.ManagedService{
				{Name: "api-gateway", CWD: "/tmp/api-gateway", Command: "node server.js", Ports: []int{3000}},
				{Name: "api-auth", CWD: "/tmp/api-auth", Command: "node auth.js", Ports: []int{3001}},
			},
		}
		m := newTopModel(deps)
		m.mode = viewModeTable
		m.focus = focusRunning
		m.selected = 0
		m.width = 120
		m.height = 30

		m.Update(tea.KeyPressMsg{Code: 'g'})
		// Switch to managed to trigger group stop (group actions require managed focus)
		m.focus = focusManaged
		m.managedSel = 0
		m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
		assert.NotNil(t, m.confirm)

		// Now switch back to running and render
		m.focus = focusRunning
		visible := m.visibleServers()
		displayNames := m.displayNames(visible)
		content := m.renderRunningTable(120, visible, displayNames)
		lines := strings.Split(content, "\n")

		// Find the data lines (skip header + divider)
		for _, line := range lines[2:] {
			stripped := ansi.Strip(line)
			// Selected row should be amber
			if m.selected >= 0 && strings.Contains(stripped, "api-auth") {
				assert.Contains(t, line, "48;5;178", "selected running row in group confirm should be amber")
			}
			// Non-selected group member should be dimmed orange
			if strings.Contains(stripped, "api-gateway") && !strings.Contains(stripped, "api-auth") {
				assert.Contains(t, line, "48;5;94", "non-selected group member should be dimmed orange")
			}
		}
	})
}
