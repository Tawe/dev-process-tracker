package cli

import tuipkg "github.com/devports/devpt/pkg/cli/tui"

// TopCmd starts the interactive TUI mode (like 'top').
func (a *App) TopCmd() error {
	return tuipkg.Run(NewTUIAdapter(a))
}
