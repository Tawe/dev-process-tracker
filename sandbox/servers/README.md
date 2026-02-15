# DevPT Server Testbed

This folder provides different local server implementations to test `devpt` discovery, start/stop, logs, naming, and port binding behavior.

## Included servers

- `node-basic` (`npm run dev`) default port `3100`
- `node-port-fallback` (`npm run dev`) default start port `3200`, auto-increments if busy
- `python-basic` (`python3 server.py`) default port `3300`
- `go-basic` (`go run main.go`) default port `3400`
- `node-crash` (`npm run dev`) default port `3500`, crashes intentionally after startup
- `node-warnings` (`npm run dev`) default port `3600`, logs periodic warnings while running

All servers expose:
- `/` plain text response
- `/health` JSON response

## Register with devpt

Run from repo root:

```bash
devpt add node-basic /Users/johnmunn/Documents/projects/dev-process-tracker/sandbox/servers/node-basic "npm run dev" 3100
devpt add node-fallback /Users/johnmunn/Documents/projects/dev-process-tracker/sandbox/servers/node-port-fallback "npm run dev" 3200
devpt add python-basic /Users/johnmunn/Documents/projects/dev-process-tracker/sandbox/servers/python-basic "python3 server.py" 3300
devpt add go-basic /Users/johnmunn/Documents/projects/dev-process-tracker/sandbox/servers/go-basic "go run main.go" 3400
devpt add node-crash /Users/johnmunn/Documents/projects/dev-process-tracker/sandbox/servers/node-crash "npm run dev" 3500
devpt add node-warnings /Users/johnmunn/Documents/projects/dev-process-tracker/sandbox/servers/node-warnings "npm run dev" 3600
```

Then launch TUI:

```bash
devpt
```

## Useful tests

1. Start all services from managed list and verify they appear in running list.
2. Start `node-fallback` twice and verify name suffixing + changing bound ports.
3. Start `node-crash` and confirm it transitions to `crashed`; inspect `devpt status node-crash`.
4. Start `node-warnings` and open logs to confirm warning lines.
5. Stop rows with `Ctrl+E`; verify PID exits.
6. Resize terminal to verify row and footer wrapping.

## Cleanup

```bash
devpt stop node-basic
devpt stop node-fallback
devpt stop python-basic
devpt stop go-basic
devpt stop node-crash
devpt stop node-warnings

devpt ls --details
```
