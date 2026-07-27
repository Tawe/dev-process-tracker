package models

import (
	"fmt"
	"strings"
)

// blockedShellPatterns are disallowed shell metacharacter sequences in managed
// service commands. Shared by the CLI `add` command and the TUI add/edit form
// so validation cannot drift between them.
var blockedShellPatterns = []string{
	"&&", "||", ";", "|", ">", "<", "`", "$(", "${",
}

// FirstBlockedShellPattern returns the first disallowed shell metacharacter
// sequence found in command, if any.
func FirstBlockedShellPattern(command string) (string, bool) {
	cmd := strings.TrimSpace(command)
	if cmd == "" {
		return "", false
	}
	for _, p := range blockedShellPatterns {
		if strings.Contains(cmd, p) {
			return p, true
		}
	}
	return "", false
}

// ValidateManagedCommand rejects empty or shell-injection-prone commands.
func ValidateManagedCommand(command string) error {
	cmd := strings.TrimSpace(command)
	if cmd == "" {
		return fmt.Errorf("command cannot be empty")
	}
	if p, ok := FirstBlockedShellPattern(cmd); ok {
		return fmt.Errorf("command contains disallowed shell pattern %q; use a direct executable command (e.g. \"npm run dev\")", p)
	}
	return nil
}

// ValidateManagedServiceFields validates the editable fields shared by the CLI
// `add` command and the TUI add/edit form: non-empty name + valid command.
func ValidateManagedServiceFields(name, command string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return ValidateManagedCommand(command)
}
