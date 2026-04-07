# DevPortTrack Debug Protocol

> Runtime coverage index: 2 runtimes (devpt-cli, sandbox fixtures)

---

## Runtime: `devpt-cli`

| Field      | Value                                      |
|------------|--------------------------------------------|
| `id`       | devpt-cli                                  |
| `class`    | backend / CLI                              |
| `entry`    | `cmd/devpt/main.go`                        |
| `owner`    | root                                       |
| `observe`  | stdout/stderr, `~/.config/devpt/logs/`     |
| `control`  | `./devpt {start\|stop\|restart} <name...>` |
| `inject`   | `go run ./cmd/devpt`                       |
| `rollout`  | `go build && ./devpt <cmd>`                |
| `test`     | `go test ./...`                            |

---

### devpt-cli / OBSERVE / VERIFIED

- Action: `./devpt ls`
- Signal: Tabular output showing Name, Port, PID, Project, Source, Status
- Constraints: Requires `lsof` and `ps` system utilities (macOS only)

### devpt-cli / CONTROL / VERIFIED

- Action:
  ```bash
  ./devpt add test-svc /path/to/cwd "command" 3400
  ./devpt start test-svc
  ./devpt stop test-svc
  ./devpt restart test-svc
  ./devpt start 'test-*'
  ./devpt stop test-svc:3400
  ```
- Signal:
  - `start`: start/status lines for each targeted service
  - `stop`: stop/status lines for each targeted service
  - `restart`: restart/status lines for each targeted service
- Constraints:
  - Registry stored at `~/.config/devpt/registry.json`
  - Logs written to `~/.config/devpt/logs/<service-name>/<timestamp>.log`
  - Processes spawn in separate process groups (setpgid)
  - Quote glob patterns to avoid shell expansion before `devpt` sees them
  - `name:port` can be used to target a specific managed service identifier

### devpt-cli / ROLLOUT / VERIFIED

- Action: Build and verify version output
- Signal: `devpt version 0.2.2` (via `./devpt --version`)
- Constraints: No hot reload; requires full rebuild
- See: `.github/copilot-instructions.md` → Quick Reference for build commands

### devpt-cli / TEST / VERIFIED

- Action: Run test suite
- Signal: `ok` for each package; coverage 39.3% (cli), 59.1% (tui)
- Constraints: Tests in `pkg/cli/*_test.go`, `pkg/cli/tui/*_test.go`, `pkg/process/*_test.go`
  - `tui_state_test.go`: Model state transitions (5 tests)
  - `tui_ui_test.go`: UI rendering verification (23 tests, 51 subtests)
  - `tui_key_input_test.go`: Key input handling
  - `tui_viewport_test.go`: Viewport scrolling tests
  - `app_batch_test.go`: Batch operations
  - `app_matching_test.go`: Pattern matching
  - `command_validation_test.go`: Command validation
  - `manager_parse_test.go`: Process command parsing (2 tests)
- See: `.github/copilot-instructions.md` → Testing section for commands

### devpt-cli / TEST / UI VERIFICATION

- Action: Run UI rendering tests
- Signal: `PASS` for all 23 tests covering:
  - Escape sequences (screen clear, ANSI codes)
  - Layout structure (table headers, columns, dividers, footer-based filter state)
  - Responsive design (widths 40-200 chars, heights 10-100 lines)
  - All view modes (table, logs, command, search, help, confirm)
  - Footer content (keybindings, live filter rendering, status)
- Constraints:
  - Tests verify rendered content, not specific ANSI colors
  - Footer assertions tolerate wrapping
  - No external deps beyond `testify/assert`
  - Focused command for current UI work: `go test -mod=mod ./pkg/cli/tui ./pkg/cli`

### devpt-cli / OBSERVE / TUI INTERACTIONS / VERIFIED

- Action: `./devpt`
- Signal:
  - top table shows running services
  - lower section shows `Managed Services (<count>)`
  - `/` activates inline footer filter editing
  - `?` opens a centered help modal
  - logs view header is `Logs: <service> | Port: <port> | PID: <pid>`
- Constraints:
  - mouse click selects rows
  - mouse wheel and page keys scroll the active viewport
  - help and confirmation dialogs are overlay modals, not separate screens

### devpt-cli / INJECT / VERIFIED

- Action: `go run ./cmd/devpt <command>`
- Signal: Immediate execution without explicit build step
- Constraints: Slower than compiled binary

### devpt-cli / EGRESS / N/A

- Rationale: CLI outputs directly to stdout/stderr; no sandboxed context

### devpt-cli / STATE / VERIFIED

- Action:
  ```bash
  # Add managed service to registry
  ./devpt add my-app /path/to/project "npm run dev" 3000

  # Verify registry state
  cat ~/.config/devpt/registry.json | jq '.services["my-app"]'
  ```
- Signal: JSON entry created in registry with name, cwd, command, ports, timestamps
- Constraints: Registry is file-based JSON; thread-safe via RWMutex

---

## Runtime: `sandbox/servers/*` (Test Fixtures)

| Field      | Value                                                                       |
|------------|-----------------------------------------------------------------------------|
| `id`       | go-basic, node-basic, node-crash, node-warnings, node-port-fallback, python-basic |
| `class`    | test fixtures                                                               |
| `entry`    | `sandbox/servers/<name>/main.go` or `server.js` or `dev.js`                 |
| `owner`    | devpt-cli (managed)                                                         |
| `observe`  | `~/.config/devpt/logs/<name>/*.log`                                         |
| `control`  | Via devpt-cli: `./devpt {start\|stop} <name>`                               |
| `inject`   | `go run .` (Go) or `node server.js` (Node)                                  |
| `rollout`  | Rebuild + restart via devpt                                                 |
| `test`     | No dedicated tests (fixtures for manual testing)                            |

### go-basic / OBSERVE / VERIFIED

- Action: `./devpt logs test-go-basic --lines 5`
- Signal: `2026/03/12 14:59:04 [go-basic] listening on http://localhost:3400`
- Constraints: Logs captured only for managed services started via `devpt start`

### go-basic / INJECT / VERIFIED

- Action:
  ```bash
  cd sandbox/servers/go-basic
  go run .
  ```
- Signal: `[go-basic] listening on http://localhost:3400`
- Constraints: Runs in foreground; use with `&` for background execution

---

## Debug Helper Commands

```bash
# Quick rebuild and test
go build -o devpt ./cmd/devpt && ./devpt ls

# Run all CLI tests with coverage
go test ./pkg/cli/... -cover

# Run the focused TUI and CLI package suite used for current UI work
go test -mod=mod ./pkg/cli/tui ./pkg/cli

# Run specific test with verbose output
go test -v ./pkg/cli -run TestWarnLegacyManagedCommands

# Run UI rendering tests (visual regression checks)
go test -v ./pkg/cli/tui -run TestView

# Run state transition tests
go test -v ./pkg/cli/tui -run TestTUI

# View registry state
cat ~/.config/devpt/registry.json | jq '.'

# Check logs for a service
ls ~/.config/devpt/logs/<service-name>/
cat ~/.config/devpt/logs/<service-name>/*.log | tail -20

# Quick health check on a running service
curl -s http://localhost:<port>/health
```
