# Changelog

## 0.6.0

- Added in-TUI editing for managed services so you can add or edit service configs without touching config files; includes `--force` to overwrite existing entries
- Added colored, clickable action buttons in the details pane and paste-into-form support so service edits can be triggered directly from the running list
- Fixed Windows release build so the cross-compile step no longer fails with `undefined: getProcessCommand` (the v0.5.0 and v0.5.1 releases were broken on this step)
- Updated CI to Node 24-compatible GitHub Actions so release runs no longer emit deprecation warnings

## 0.5.1

- Fixed `remove`/`rm` command so `devpt remove <name>` works from the CLI instead of failing as unknown; also added to help text

## 0.5.0

- Added resolved command capture at spawn time so the system learns the OS-interpreted command (e.g., `bunx vite` → `node .../vite`) for reliable identity matching
- Added process start time to the identity evidence chain so PID reuse is detected safely
- Added universal details pane so selecting a running service shows its full details alongside managed services
- Added per-process memory display in the details pane with color-coded thresholds so resource usage is visible at a glance
- Added copy-to-clipboard icon next to command text in logs and details pane
- Fixed TUI managed stop to route through the lifecycle layer instead of raw PID calls so "invalid pid: 0" no longer occurs on stale registry entries
- Fixed TUI restart/stop routing so keyboard shortcuts target managed services when the managed list has focus
- Fixed scanner to recognize versioned Python binaries (e.g., `python3.12`) in runtime command checks
- Refactored identity matching to use an ordered evidence chain (PID+time → port → CWD+command → CWD → root) for shared-CWD correctness
- Refactored TUI row color logic into a shared source of truth across running table and managed list
- Updated PROCESS_MANAGEMENT.md with identity architecture, resolved command capture, restart preflight rules, and non-negotiable rules

## 0.4.2

- Fixed port-bound readiness timeout so services like Open WebUI that take 10–15s to bind their port are no longer falsely marked unhealthy
- Fixed false ambiguity warnings so processes already uniquely claimed by another service via their port binding are skipped
- Fixed managed details pane click routing so clicking the right-side details pane no longer selects items in the left-side service list
- Fixed Windows cross-compilation so the lock file compiles without missing `syscall.Kill`
- Refactored package internals to remove ~330 lines of dead code, unreachable paths, and duplicated logic

## 0.4.1

- Fixed Linux crash when running as non-root by adding /proc/net/tcp fallback so lsof is no longer required
- Refactored TUI render-path to reduce recomputation overhead
- Aligned process lifecycle with behavioral contract for consistent start/stop/restart behavior
- Refactored TUI commands module into focused files for maintainability

## 0.4.0

- Added namespace-based process grouping so related managed services can be controlled together
- Added OSC 8 clickable hyperlinks to the TUI so service names and commands are directly actionable from the terminal
- Added wildcard pattern support to the status command so multiple services can be queried at once
- Added service metadata to the managed details pane so context like namespace and tags are visible alongside process info
- Fixed namespace extraction so leading non-alphanumeric characters are handled correctly
- Fixed ^C in command mode so it properly cancels without side effects and managed list/details scrolling is independent

## 0.3.0

- Added a managed-services split view in the TUI so selection and navigation stay clear when browsing running and registered services
- Fixed TUI selection behavior so focus, row targeting, and split-pane navigation stay aligned while moving between running and managed services

## 0.2.2

- Added a Shift+S sort direction toggle in the TUI so sort order can be reversed without changing the active column
- Fixed managed service PID validation so stop and restart only act on processes that still match the registered service
- Fixed cross-platform builds by separating Unix and Windows process control paths

## 0.2.1

- Added table sorting controls with mouse support and reverse sort in the TUI

## 0.2.0

- Added multi-service `start`, `stop`, and `restart` commands with quoted glob pattern support so multiple managed services can be controlled in one invocation
- Added `name:port` targeting for managed services so ambiguous service names can be disambiguated from the CLI
- Extracted the Bubble Tea UI into `pkg/cli/tui` so the TUI logic is isolated from the main CLI package
- Added mouse row selection, mouse wheel scrolling, and viewport-focused navigation so table and log interaction works without keyboard-only control
- Added centered modal overlays for help and confirmation dialogs so help and destructive actions no longer replace the main table view
- Replaced the ad hoc search field with Bubbles text input so filter editing behaves like a real input control and updates inline in the footer
- Simplified the table chrome by moving counts into headers, bolding the active sort column, and removing redundant status text from the top of the screen
- Fixed `Enter` handling so the top section opens logs and the bottom section starts the selected managed service without being swallowed by confirm bindings
- Fixed log rendering so the header is separated from the first log line and the viewport uses the actual remaining terminal height
- Fixed stale table layout offsets so footer spacing, viewport sizing, and mouse hit-testing stay aligned after the filter moved into the footer
- Added shared keymap-driven help text with Bubble components so visible shortcuts and actual bindings stay in sync
- Added clearer TUI and quickstart documentation so the current footer filter, modal help, mouse controls, batch commands, and logs header behavior are documented
- Bumped the application version to `0.2.0` and rendered the version in the TUI header in muted gray
