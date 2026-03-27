# Dev Process Tracker - Quick Start Guide

## Installation

Build from source:
```bash
cd ~/path/to/dev-process-tracker
go build -o devpt ./cmd/devpt/main.go
```

Optionally install to PATH:
```bash
sudo mv devpt /usr/local/bin/devpt
```

Then use from anywhere:
```bash
devpt ls
```
## First steps

### See running services

```bash
devpt ls
```

Shows listening ports with PID, project, and source.

### Register a managed service

```bash
devpt add myapp ~/myapp "npm start" 3000
```

### List with details

```bash
devpt ls --details
```

### Check your registered services

```bash
cat ~/.config/devpt/registry.json
```

## Common Workflows

### Start a managed service

```bash
devpt start myapp
```

Logs are written to `~/.config/devpt/logs/myapp/<timestamp>.log`

### Start multiple services at once

```bash
# Start multiple specific services
devpt start api frontend worker

# Use glob patterns to match services (quote to prevent shell expansion)
devpt start 'web-*'        # Starts all services matching 'web-*'
devpt start '*-test'       # Starts all services ending with '-test'

# Target a specific service by name:port
devpt start web-api:3000   # Start web-api on port 3000 only
devpt stop "some:thing"    # Literal service name containing a colon

# Mix patterns and specific names
devpt start api 'web-*' worker
```

### Stop a service by name

```bash
devpt stop myapp
```

### Stop multiple services at once

```bash
# Stop multiple specific services
devpt stop api frontend

# Use glob patterns (quote to prevent shell expansion)
devpt stop 'web-*'        # Stops all services matching 'web-*'

# Target a specific service by name:port
devpt stop web-api:3000   # Stop web-api on port 3000 only
devpt stop *-test         # Stops all services ending with '-test'
```

### Stop a service by port

```bash
devpt stop --port 3000
```

### Restart a service

```bash
devpt restart myapp
```

### Restart multiple services at once

```bash
# Restart multiple specific services
devpt restart api frontend worker

# Use glob patterns
devpt restart web-*       # Restarts all services matching 'web-*'
devpt restart claude-*    # Restarts all services starting with 'claude-'
```

### View logs

```bash
devpt logs myapp
devpt logs myapp --lines 100
```

### Use the TUI

```bash
devpt
```

Key interactions:
- `Tab` switches between the running-services table and the managed-services list
- `Enter` opens logs from the top table and starts the selected service from the bottom list
- `/` opens inline filter editing in the footer
- `?` opens the help modal
- mouse click selects rows and mouse wheel scrolls the active pane
- logs header shows `Logs: <service> | Port: <port> | PID: <pid>`

## File Locations

```
~/.config/devpt/
├── registry.json          # Your managed services
└── logs/
    ├── myapp/
    │   ├── 2026-02-09T16-00-01.log
    │   └── 2026-02-09T16-05-30.log
    └── otherapp/
        └── 2026-02-09T16-10-00.log
```

## Notes

1. **Edit registry manually** - `~/.config/devpt/registry.json` is just JSON
2. **Check what's using a port** - `devpt ls --details | grep :3000`
3. **Find projects** - `devpt ls | grep "my-project"`
4. **See processes without names** - `devpt ls --details | grep -v "^-"`
5. **Quote glob patterns** - use `'web-*'` instead of `web-*` to avoid shell expansion

## Troubleshooting

**"lsof: command not found"**
```bash
brew install lsof
```

**Registry file seems broken**
```bash
rm ~/.config/devpt/registry.json
# It will be recreated next time you add a service
```

**Process won't stop**
```bash
# Find the PID
devpt ls | grep myapp

# Force kill it (use carefully!)
kill -9 <PID>
```

## Help

```bash
devpt help
```
