# Dev Process Tracker

![Dev Process Tracker hero](devpttitle.png)

Dev Process Tracker (`devpt`) tracks and controls local dev services.

## What it does

- Opens an interactive TUI by default (`devpt`)
- Shows running services with name, port, pid, project, command, and health
- Tracks managed services you register with `devpt add`
- Lets you start, restart, stop, remove, and inspect services
- Provides logs for managed services and best-effort logs for unmanaged processes
- Marks managed services as `crashed` when they exit unexpectedly and shows an inferred crash reason

## Install

```bash
go build -o devpt ./cmd/devpt
```

## Run tests

```bash
go test ./...
```

## Challenge smoke test

Run a smoke flow in an isolated temp home:

```bash
./scripts/challenge_smoke_test.sh
```

This runs `build`, `test`, and core command flow: `add`, `start`, `status`, `logs`, `restart`, `ls`, and `stop`.

## Quick start

```bash
# Open the TUI (default)
devpt

# Register a service
devpt add my-app ~/projects/my-app "npm run dev" 3000

# Start / stop / restart
devpt start my-app
devpt stop my-app
devpt restart my-app

# Logs
devpt logs my-app --lines 100

# Batch operations
devpt start api frontend worker
devpt restart 'web-*'
devpt stop web-api:3000
```

## CLI commands

### Default

```bash
devpt
```

Opens the TUI.

### Manage services

```bash
devpt add <name> <cwd> "<cmd>" [ports...]
devpt start <name> [<name>...]          # Start one or more services
devpt stop <name> [<name>...]           # Stop one or more services
devpt stop --port <port>
devpt restart <name> [<name>...]        # Restart one or more services
devpt logs <name> [--lines N]
```

### Batch operations

Start, stop, or restart multiple services at once:

```bash
# Start multiple specific services
devpt start api frontend worker

# Use glob patterns to match service names
devpt start 'web-*'        # Starts all services matching 'web-*'
devpt stop '*-test'        # Stops all services ending with '-test'
devpt restart 'claude-*'   # Restarts all services starting with 'claude-*'

# Target specific service by name:port
devpt start web-api:3000   # Start web-api on port 3000 only
devpt stop "some:thing"    # Service with colon in literal name

# Mix patterns and specific names
devpt start api 'web-*' worker
```

Batch operations run sequentially, print per-service status, continue on failure, and return exit code `1` if any service fails.

### Inspect

```bash
devpt ls [--details]
devpt status <name|port>
```

`devpt status <name>` includes `CRASH DETAILS` for crashed managed services with an inferred reason and recent log lines.

### Meta

```bash
devpt help
devpt --version
```

## TUI keymap

- `Tab`: switch focus between running and managed lists
- `Enter`:
  - running list: open logs
  - managed list: start selected service
- mouse click: select rows in either list
- mouse wheel / page keys: scroll the active viewport
- `Ctrl+E`: stop selected running service (with confirm)
- `Ctrl+R`: restart selected running managed service
- `Ctrl+A`: open command input (`add ...` prefilled)
- `x` / `Delete` / `Ctrl+D`: remove selected managed service (with confirm)
- `/`: edit the inline filter in the footer
- `Ctrl+L`: clear filter
- `s`: cycle sort mode
- `h`: toggle health detail
- `?`: open help modal
- `b`: back from logs/command
- `f`: toggle log follow mode (in logs view)
- `q`: quit

## TUI layout

- Running services are shown in the top table. The active sort column header is bold.
- Managed services are shown in a separate section below with the total count in the section title.
- Filter state lives in the footer help row:
  - default: `/ filter`
  - editing: `/ >query`
  - applied: `/ query`
- Help and confirmation are rendered as centered modals over the table.
- Logs view header is rendered as `Logs: <service> | Port: <port> | PID: <pid>`.

## TUI command input

TUI command mode (`:` or `Ctrl+A`) supports:

```text
add <name> <cwd> "<cmd>" [ports...]
start <name>
stop <name|--port PORT>
remove <name>
restore <name>
list
help
```

## AI Agent Detection

Detected AI-started servers show `agent:name` in the source column instead of `manual`.

### Detection methods

1. **Parent process name**: `claude`, `cursor`, `copilot`, and similar names
2. **Environment variables**: `CLAUDE_*`, `CURSOR_*`, `COPILOT_*` prefixes on platforms where available

### Naming convention

Use a naming prefix if you want ownership to be obvious in the registry:

```bash
# Services started by Claude
devpt add claude-frontend ~/projects/frontend "npm run dev" 3000
devpt add claude-api ~/projects/backend "go run main.go" 8000

# Services started by Cursor
devpt add cursor-worker ~/projects/worker "npm start" 4000

# Services started by Copilot
devpt add copilot-service ~/projects/service "python app.py" 5000
```

### Example with built-in test servers

```bash
# From repo root, register test servers with AI owner names
devpt add claude-node ./sandbox/servers/node-basic "npm run dev" 3100
devpt add claude-python ./sandbox/servers/python-basic "python3 server.py" 3300
devpt add cursor-go ./sandbox/servers/go-basic "go run main.go" 3400
devpt add copilot-node-fallback ./sandbox/servers/node-port-fallback "npm run dev" 3200
devpt add claude-node-crash ./sandbox/servers/node-crash "npm run dev" 3500
devpt add cursor-node-warnings ./sandbox/servers/node-warnings "npm run dev" 3600

# Start them
devpt start claude-node
devpt start claude-python
devpt start cursor-go
devpt start copilot-node-fallback
devpt start claude-node-crash
devpt start cursor-node-warnings

# View in devpt TUI
devpt
```

Each test server exposes `/health` and `/`.

## Notes

- Managed services are registry entries you control via `devpt`.
- Running list is process-driven. Managed services can appear even before a port is bound.
- `name:port` is supported for CLI targeting where multiple services share a base name.
- Quote glob patterns like `'web-*'` so your shell does not expand them first.
- If stop needs elevated permissions, TUI asks for confirmation to run `sudo kill -9 <pid>`.
- Service names can include a prefix (e.g., `claude-`, `cursor-`, `copilot-`) to indicate AI agent ownership in your registry.
- No login or API credentials are required for judges to run this project locally.

## Troubleshooting

### Service not appearing

Check running listeners:

```bash
lsof -nP -iTCP -sTCP:LISTEN
```

Check registry entry:

```bash
devpt ls --details
```

### Process won’t stop

Try from TUI first (`Ctrl+E`). If escalation is required, run:

```bash
sudo kill -9 <pid>
```

### Logs unavailable for unmanaged process

Some processes only write to attached terminal output. In that case there may be nothing tail-able from files/unified logs.
