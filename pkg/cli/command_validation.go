package cli

import (
	"github.com/devports/devpt/pkg/models"
)

// validateManagedCommand rejects empty or shell-injection-prone commands.
// Thin wrapper over models.ValidateManagedCommand (shared with the TUI form).
func validateManagedCommand(command string) error {
	return models.ValidateManagedCommand(command)
}

// firstBlockedShellPattern returns the first disallowed shell metacharacter
// sequence in command, if any. Thin wrapper over models.FirstBlockedShellPattern.
func firstBlockedShellPattern(command string) (string, bool) {
	return models.FirstBlockedShellPattern(command)
}

// validateManagedServiceFields validates editable fields shared by the CLI
// `add` command and the TUI add/edit form: non-empty name + valid command.
func validateManagedServiceFields(name, command string) error {
	return models.ValidateManagedServiceFields(name, command)
}
