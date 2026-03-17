package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestTUISimpleUpdate tests model updates directly without running the full program
func TestTUISimpleUpdate(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	model := newTopModel(app)

	t.Run("tab switches focus between running and managed", func(t *testing.T) {
		initialFocus := model.focus

		// Send Tab key
		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})

		// Should not return a command
		assert.Nil(t, cmd)

		// Focus should change
		updatedModel := newModel.(*topModel)
		assert.NotEqual(t, initialFocus, updatedModel.focus, "Focus should change after Tab")

		// Focus should toggle between the two modes
		if initialFocus == focusRunning {
			assert.Equal(t, focusManaged, updatedModel.focus)
		} else {
			assert.Equal(t, focusRunning, updatedModel.focus)
		}
	})

	t.Run("escape key in logs mode returns to table", func(t *testing.T) {
		model.mode = viewModeLogs

		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

		assert.Nil(t, cmd)
		updatedModel := newModel.(*topModel)
		assert.Equal(t, viewModeTable, updatedModel.mode, "Should return to table mode")
	})

	t.Run("forward slash enters search mode", func(t *testing.T) {
		model.mode = viewModeTable

		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

		assert.Nil(t, cmd)
		updatedModel := newModel.(*topModel)
		assert.Equal(t, viewModeSearch, updatedModel.mode, "Should enter search mode")
	})

	t.Run("question mark enters help mode", func(t *testing.T) {
		model.mode = viewModeTable

		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

		assert.Nil(t, cmd)
		updatedModel := newModel.(*topModel)
		assert.Equal(t, viewModeHelp, updatedModel.mode, "Should enter help mode")
	})

	t.Run("s key cycles through sort modes", func(t *testing.T) {
		// Ensure we're in table mode for sort to work
		model.mode = viewModeTable
		initialSort := model.sortBy

		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

		assert.Nil(t, cmd)
		updatedModel := newModel.(*topModel)
		assert.NotEqual(t, initialSort, updatedModel.sortBy, "Sort mode should cycle")
	})
}

// TestTUIKeySequence tests a sequence of keypresses
func TestTUIKeySequence(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	t.Run("navigate and return to table", func(t *testing.T) {
		model := newTopModel(app)
		initialMode := model.mode

		// Press '/' to enter search mode
		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		model = newModel.(*topModel)
		assert.Equal(t, viewModeSearch, model.mode)

		// Press Esc to return to table
		newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		model = newModel.(*topModel)
		assert.Equal(t, initialMode, model.mode)
	})

	t.Run("help mode and exit", func(t *testing.T) {
		model := newTopModel(app)

		// Press '?' to enter help
		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		model = newModel.(*topModel)
		assert.Equal(t, viewModeHelp, model.mode)

		// Press Esc to exit help
		newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		model = newModel.(*topModel)
		assert.Equal(t, viewModeTable, model.mode)
	})
}

// TestTUIQuitKey tests that q key produces quit command
func TestTUIQuitKey(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	model := newTopModel(app)

	t.Run("q key returns quit command", func(t *testing.T) {
		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		// Should return a command (quit command)
		assert.NotNil(t, cmd, "q key should return a command")
	})

	t.Run("ctrl+c returns quit command", func(t *testing.T) {
		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

		assert.NotNil(t, cmd, "ctrl+c should return a command")
	})
}

// TestTUIViewRendering tests that View() returns expected content
func TestTUIViewRendering(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	model := newTopModel(app)
	model.width = 100
	model.height = 40

	t.Run("table view contains expected elements", func(t *testing.T) {
		model.mode = viewModeTable
		output := model.View()

		// Check for expected UI elements
		assert.Contains(t, output, "Dev Process Tracker", "Should show title")
		assert.Contains(t, output, "Name", "Should have Name column")
		assert.Contains(t, output, "Port", "Should have Port column")
		assert.Contains(t, output, "PID", "Should have PID column")
	})

	t.Run("help view contains help text", func(t *testing.T) {
		model.mode = viewModeHelp
		output := model.View()

		assert.Contains(t, output, "Keymap", "Should show keymap header")
		assert.Contains(t, output, "q quit", "Should mention quit key")
	})
}

// TestViewportStateTransitions tests state transitions for viewport interactions
// Covers: OBL-highlight-state, OBL-viewport-integration
func TestViewportStateTransitions(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	t.Run("viewport state initialization", func(t *testing.T) {
		model := newTopModel(app)

		// After implementation: model should have viewport, highlightIndex, highlightMatches fields
		_ = model
		t.Skip("TODO: Verify viewport state fields exist - OBL-highlight-state")
	})

	t.Run("highlight index boundary conditions", func(t *testing.T) {
		model := newTopModel(app)
		model.mode = viewModeLogs
		model.highlightMatches = []int{10, 20, 30}

		// Test lower boundary
		model.highlightIndex = 0
		_ = model

		// Test upper boundary
		model.highlightIndex = len(model.highlightMatches) - 1
		_ = model

		t.Skip("TODO: Test boundary conditions - Edge-2")
	})

	t.Run("highlight index with empty matches", func(t *testing.T) {
		model := newTopModel(app)
		model.mode = viewModeLogs
		model.highlightMatches = []int{}
		model.highlightIndex = 0

		// Should handle gracefully without crash
		_ = model
		t.Skip("TODO: Handle empty highlights - Edge case")
	})
}
