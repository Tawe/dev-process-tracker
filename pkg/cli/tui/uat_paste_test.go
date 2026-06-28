package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/devports/devpt/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UAT for bracketed-paste into the add/edit form (DEVPT-020).
// Each test maps 1:1 to an acceptance criterion in uat-paste.md.
//
// Bubbletea emits tea.PasteMsg{Content} on bracketed paste (macOS Cmd+V /
// terminal Ctrl+Shift+V). textinput inserts the content at the cursor; these
// tests verify the form routes the message to the focused field.

// AC-P1: a paste inserts text into the focused field.
func TestUAT_PasteInsertsIntoFocusedField(t *testing.T) {
	m := newTopModel(&fakeAppDeps{})
	m.mode = viewModeTable
	m.openAddForm()
	require.Equal(t, formFieldName, m.form.focus)

	m.Update(tea.PasteMsg{Content: "api-gateway"})

	assert.Equal(t, "api-gateway", m.form.fields[formFieldName].Value(),
		"pasted text must appear in the focused (Name) field")
}

// AC-P2: paste routes to the focused field only; others stay empty.
func TestUAT_PasteTargetsOnlyFocusedField(t *testing.T) {
	m := newTopModel(&fakeAppDeps{})
	m.mode = viewModeTable
	m.openAddForm()
	// Move focus to the Command field (index 2).
	m.form.focusField(formFieldCommand)

	m.Update(tea.PasteMsg{Content: "npm run dev"})

	assert.Equal(t, "npm run dev", m.form.fields[formFieldCommand].Value())
	for _, i := range []int{formFieldName, formFieldDir, formFieldPorts} {
		assert.Equal(t, "", m.form.fields[i].Value(),
			"field %d must be untouched by a paste into Command", i)
	}
}

// AC-P3: paste works in every field; also covers the Edit form (AC-P6).
func TestUAT_PasteWorksInAllFields(t *testing.T) {
	m := newTopModel(&fakeAppDeps{})
	m.mode = viewModeTable
	// Edit form (pre-filled) to cover AC-P6.
	svc := &models.ManagedService{Name: "old", CWD: "/", Command: "x", Ports: []int{1}}
	fs := formForEdit(svc)
	m.form = &fs

	want := map[int]string{
		formFieldName:    "renamed-svc",
		formFieldDir:     "/srv/app",
		formFieldCommand: "go run ./cmd",
		formFieldPorts:   "3000, 8080",
	}
	for field, content := range want {
		m.form.focusField(field)
		m.Update(tea.PasteMsg{Content: content})
		// Pre-filled values exist; assert the pasted text is present in that field.
		assert.Contains(t, m.form.fields[field].Value(), content,
			"field %d should contain pasted %q", field, content)
	}
}

// AC-P4: paste with no form open is a no-op (no panic, no search mutation).
func TestUAT_PasteNoOpWhenFormClosed(t *testing.T) {
	m := newTopModel(&fakeAppDeps{})
	m.mode = viewModeTable
	require.Nil(t, m.form)
	beforeSearch := m.searchInput.Value()

	model, _ := m.Update(tea.PasteMsg{Content: "should go nowhere"})

	updated, ok := model.(*topModel)
	require.True(t, ok)
	assert.Nil(t, updated.form)
	assert.Equal(t, beforeSearch, updated.searchInput.Value(),
		"paste must not mutate the search input when no form is open")
}

// AC-P5: pasted content is validated on submit. Pasting a shell-pattern command
// then submitting keeps the form open with an error.
func TestUAT_PastedContentIsValidatedOnSubmit(t *testing.T) {
	m := newTopModel(&fakeAppDeps{})
	m.mode = viewModeTable
	m.openAddForm()
	// Paste a valid name and a shell-injection command.
	m.Update(tea.PasteMsg{Content: "api"})
	m.form.focusField(formFieldCommand)
	m.Update(tea.PasteMsg{Content: "rm -rf x && y"})

	// Submit.
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	require.NotNil(t, m.form, "form must stay open on invalid pasted command")
	assert.NotEmpty(t, m.form.err, "an error must be shown for the disallowed pattern")
}
