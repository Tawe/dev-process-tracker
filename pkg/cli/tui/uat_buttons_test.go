package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/devports/devpt/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UAT for the details-pane action buttons being functional (DEVPT-020):
// each button routes to its action. Tested via dispatchDetailsAction (the click
// handler delegates to it after hit-testing), so the assertion is on the action
// side effects, not fragile mouse coordinates.

func newActionButtonTestModel() *topModel {
	m := newTopModel(&fakeAppDeps{
		services: []*models.ManagedService{
			{Name: "api", CWD: "/srv", Command: "npm run dev", Ports: []int{3000}},
		},
	})
	m.width = 120
	m.height = 30
	m.mode = viewModeTable
	m.focus = focusManaged
	m.managedSel = 0
	return m
}

func TestUAT_DetailsButton_EditOpensForm(t *testing.T) {
	m := newActionButtonTestModel()
	require.Nil(t, m.form)

	m.dispatchDetailsAction("edit")

	require.NotNil(t, m.form, "edit button must open the edit form")
	assert.True(t, m.form.isEdit)
	assert.Equal(t, "api", m.form.fields[formFieldName].Value(), "edit form must pre-fill the selected service")
}

func TestUAT_DetailsButton_StartStartsService(t *testing.T) {
	m := newActionButtonTestModel()
	m.dispatchDetailsAction("start")
	assert.True(t, strings.HasPrefix(m.cmdStatus, "Started"), "start button should report Started, got %q", m.cmdStatus)
}

func TestUAT_DetailsButton_RestartRestartsService(t *testing.T) {
	m := newActionButtonTestModel()
	m.dispatchDetailsAction("restart")
	assert.Contains(t, m.cmdStatus, "Restarted", "restart button should report Restarted, got %q", m.cmdStatus)
}

func TestUAT_DetailsButton_StopOpensConfirm(t *testing.T) {
	m := newActionButtonTestModel()
	require.Nil(t, m.confirm)
	m.dispatchDetailsAction("stop")
	require.NotNil(t, m.confirm, "stop button must open the confirm modal")
	assert.Equal(t, confirmStopPID, m.confirm.kind)
}

// Unknown label is a safe no-op (defensive; shouldn't happen via render).
func TestUAT_DetailsButton_UnknownLabelNoOp(t *testing.T) {
	m := newActionButtonTestModel()
	before := m.cmdStatus
	_, cmd := m.dispatchDetailsAction("bogus")
	assert.Equal(t, before, m.cmdStatus)
	assert.Nil(t, cmd)
}

// actionRowScreenY returns the screen Y of the details action row, computed
// by inverting handleTableMouseClick's coordinate math from the model's own
// geometry fields (so it tracks the handler if geometry changes).
func actionRowScreenY(m *topModel) int {
	headerOffset := m.tableTopLines(m.width)
	yoffset := m.table.selectedDetailsVP.YOffset()
	// mouse.Y = detailsActionLine + headerOffset + lastRunningHeight - YOffset
	return m.detailsActionLine + headerOffset + m.table.lastRunningHeight - yoffset
}

// Click handler routes an action-row click to dispatchDetailsAction end to end.
// Verifies the hit-test wiring (m.detailsActionLine + btn x-ranges) is populated
// by render and consumed by handleTableMouseClick.
func TestUAT_ClickOnEditButtonOpensForm(t *testing.T) {
	m := newActionButtonTestModel()
	// Render once so detailsActionLine / detailsActionBtns are populated.
	_ = m.View().Content
	require.NotEqual(t, -1, m.detailsActionLine, "render must populate the action line")
	require.NotEmpty(t, m.detailsActionBtns)

	var edit detailsAction
	for _, b := range m.detailsActionBtns {
		if b.label == "edit" {
			edit = b
		}
	}
	require.Equal(t, "edit", edit.label)

	click := tea.MouseClickMsg{Button: tea.MouseLeft, X: m.table.lastListWidth + edit.x0, Y: actionRowScreenY(m)}
	m.Update(click)

	assert.NotNil(t, m.form, "clicking the edit button must open the edit form")
}
