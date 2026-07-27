package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/devports/devpt/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormNewIsEmpty(t *testing.T) {
	t.Parallel()
	fs := formForAdd()
	assert.False(t, fs.isEdit)
	assert.Equal(t, formFieldName, fs.focus)
	for i := range fs.fields {
		assert.Equal(t, "", fs.fields[i].Value(), "field %d should start empty", i)
	}
}

func TestFormPrefillFromService(t *testing.T) {
	t.Parallel()
	svc := &models.ManagedService{Name: "api", CWD: "/srv", Command: "npm run dev", Ports: []int{3000, 8080}}
	fs := formForEdit(svc)
	assert.True(t, fs.isEdit)
	assert.Equal(t, "api", fs.origName)
	assert.Equal(t, "api", fs.fields[formFieldName].Value())
	assert.Equal(t, "/srv", fs.fields[formFieldDir].Value())
	assert.Equal(t, "npm run dev", fs.fields[formFieldCommand].Value())
	assert.Equal(t, "3000, 8080", fs.fields[formFieldPorts].Value())
}

func TestFormRunningHintReflectsLastPID(t *testing.T) {
	t.Parallel()
	pid := 5
	running := formForEdit(&models.ManagedService{Name: "api", Command: "x", LastPID: &pid})
	assert.Equal(t, "Applies on next start/restart", running.runningHint())
	stopped := formForEdit(&models.ManagedService{Name: "api", Command: "x"})
	assert.Equal(t, "", stopped.runningHint())
}

func TestFormTabCyclesFocus(t *testing.T) {
	t.Parallel()
	fs := formForAdd()
	fs.update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, formFieldDir, fs.focus)
	// advance to the last field, then wrap back to the first
	for i := 0; i < formFieldCount-1; i++ {
		fs.update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	assert.Equal(t, formFieldName, fs.focus)
	// shift+tab reverses (from Name wraps back to Ports)
	fs.update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assert.Equal(t, formFieldPorts, fs.focus)
}

func TestFormTypingUpdatesFocusedField(t *testing.T) {
	t.Parallel()
	fs := formForAdd()
	fs.update(tea.KeyPressMsg{Text: "a", Code: 'a'})
	fs.update(tea.KeyPressMsg{Text: "b", Code: 'b'})
	assert.Equal(t, "ab", fs.fields[formFieldName].Value())
}

func TestFormCancelOnEsc(t *testing.T) {
	t.Parallel()
	fs := formForAdd()
	submit, quit, _ := fs.update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.True(t, quit)
	assert.Nil(t, submit)
}

func TestFormSubmitValidation(t *testing.T) {
	t.Parallel()
	// empty name
	fs := formForAdd()
	_, errMsg := fs.submit()
	assert.NotEmpty(t, errMsg, "empty name should error")

	// shell-pattern command
	fs = formForAdd()
	fs.fields[formFieldName].SetValue("api")
	fs.fields[formFieldCommand].SetValue("a && b")
	_, errMsg = fs.submit()
	assert.NotEmpty(t, errMsg, "shell-pattern command should error")

	// non-integer port
	fs = formForAdd()
	fs.fields[formFieldName].SetValue("api")
	fs.fields[formFieldCommand].SetValue("npm run dev")
	fs.fields[formFieldPorts].SetValue("3000, oops")
	_, errMsg = fs.submit()
	assert.NotEmpty(t, errMsg, "non-integer port should error")

	// valid input parses ports
	fs = formForAdd()
	fs.fields[formFieldName].SetValue("api")
	fs.fields[formFieldCommand].SetValue("npm run dev")
	fs.fields[formFieldPorts].SetValue("3000, 8080")
	intent, errMsg := fs.submit()
	assert.Empty(t, errMsg)
	require.NotNil(t, intent)
	assert.Equal(t, []int{3000, 8080}, intent.ports)
}

func TestFormEnterSubmitsValid(t *testing.T) {
	t.Parallel()
	fs := formForAdd()
	fs.fields[formFieldName].SetValue("api")
	fs.fields[formFieldCommand].SetValue("npm run dev")
	submit, quit, _ := fs.update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.False(t, quit)
	require.NotNil(t, submit)
	assert.False(t, submit.isEdit)
}

func TestFormEnterKeepsModalOpenOnInvalid(t *testing.T) {
	t.Parallel()
	fs := formForAdd()
	submit, quit, _ := fs.update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.False(t, quit)
	assert.Nil(t, submit)
	assert.NotEmpty(t, fs.err)
}

func TestFormSubmitEditRename(t *testing.T) {
	t.Parallel()
	fs := formForEdit(&models.ManagedService{Name: "api", CWD: "/s", Command: "x", Ports: []int{1}})
	fs.fields[formFieldName].SetValue("api-v2")
	intent, errMsg := fs.submit()
	assert.Empty(t, errMsg)
	require.NotNil(t, intent)
	assert.True(t, intent.isEdit)
	assert.True(t, intent.rename)
	assert.Equal(t, "api", intent.oldName)
	assert.Equal(t, "api-v2", intent.name)
}

func TestFormSubmitEditNoRename(t *testing.T) {
	t.Parallel()
	fs := formForEdit(&models.ManagedService{Name: "api", CWD: "/s", Command: "x", Ports: []int{1}})
	intent, errMsg := fs.submit()
	assert.Empty(t, errMsg)
	require.NotNil(t, intent)
	assert.True(t, intent.isEdit)
	assert.False(t, intent.rename)
}

// TopModel integration: submitting the edit form applies rename + field update
// through AppDeps and closes the form.
func TestTopModelEditFormAppliesRenameAndUpdate(t *testing.T) {
	deps := &fakeAppDeps{services: []*models.ManagedService{
		{Name: "api", CWD: "/old", Command: "sleep 1", Ports: []int{3000}},
	}}
	m := newTopModel(deps)
	m.mode = viewModeTable
	m.focus = focusManaged
	m.managedSel = 0
	m.openEditForm()
	require.NotNil(t, m.form)

	m.form.fields[formFieldName].SetValue("api-v2")
	m.form.fields[formFieldCommand].SetValue("npm run dev")
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Nil(t, m.form, "form should close after successful submit")
	assert.Nil(t, deps.GetService("api"), "old name should be gone")
	got := deps.GetService("api-v2")
	require.NotNil(t, got)
	assert.Equal(t, "npm run dev", got.Command)
}

// TopModel integration: invalid submit keeps the form open with an error.
func TestTopModelAddFormValidationKeepsFormOpen(t *testing.T) {
	deps := &fakeAppDeps{}
	m := newTopModel(deps)
	m.mode = viewModeTable
	m.openAddForm()
	require.NotNil(t, m.form)
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.NotNil(t, m.form, "form should stay open on invalid submit")
	assert.NotEmpty(t, m.form.err)
}
