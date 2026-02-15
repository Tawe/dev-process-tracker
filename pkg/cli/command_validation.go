package cli

import (
	"fmt"
	"strings"
)

var blockedShellPatterns = []string{
	"&&", "||", ";", "|", ">", "<", "`", "$(", "${",
}

func firstBlockedShellPattern(command string) (string, bool) {
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

func validateManagedCommand(command string) error {
	cmd := strings.TrimSpace(command)
	if cmd == "" {
		return fmt.Errorf("command cannot be empty")
	}
	if p, ok := firstBlockedShellPattern(cmd); ok {
		return fmt.Errorf("command contains disallowed shell pattern %q; use a direct executable command (e.g. \"npm run dev\")", p)
	}
	return nil
}
