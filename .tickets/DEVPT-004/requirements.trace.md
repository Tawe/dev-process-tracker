# Requirements

Ticket: `DEVPT-004`

## Behavioral Requirements

### BR-1

- `BR-1` [bdd] WHEN the managed services list is shown in the TUI, the system shall display a distinct visual state marker for each managed service state: running (using a play marker `▶`), starting, stopped, and crashed.

### BR-2

- `BR-2` [bdd] WHEN users view the managed services list in the TUI, the system shall make running, starting, stopped, and crashed services distinguishable without requiring navigation to another screen.

### BR-3

- `BR-3` [bdd] WHEN a managed service is selected in the TUI, the system shall present the managed-services area as a split view with a 50|50 list pane and details pane for the selected service.

### BR-4

- `BR-4` [bdd] IF the selected managed service is crashed, THEN the system shall display a concise failure headline using the best available reason.

### BR-5

- `BR-5` [bdd] IF the selected managed service is crashed and recent logs are available, THEN the system shall display recent log context that helps users decide whether to restart the service or inspect logs further.

### BR-6

- `BR-6` [bdd] WHEN no managed service is selected in the split view, the system shall show a placeholder details pane that prompts the user to select a managed service for inspection.

### BR-7

- `BR-7` [bdd] WHEN a managed service is selected in the split view, the system shall display the service's working directory, port(s), and command in the details pane so users can inspect service configuration at a glance.

### BR-8

- `BR-8` [bdd] WHEN a managed service is selected in the details pane, the system shall display action buttons (start/restart, stop) in the details pane header for managing the service directly.

### BR-9

- `BR-9` [bdd] WHEN action buttons are displayed in the details pane header, the system shall show appropriate buttons based on the service state: start for stopped services, restart for running/crashed services, stop for running services.

### BR-10

- `BR-10` [bdd] WHEN users scroll in the TUI, each section (running processes, managed list, details pane) shall scroll independently without affecting the scroll position of other sections.

### BR-11

- `BR-11` [bdd] WHEN a service is selected in the TUI (either running or managed), the system shall display relevant details for that service in the details pane, with the content tailored to the service type: running services show PID, port, command, directory, project, start time, agent info, and health status; managed services show their existing details format.

## Constraints

- `C1` [tests] The managed services area shall remain readable within standard terminal sizes, using a 50|50 split at normal widths and degrading gracefully at narrow widths while preserving list scanability and selected-service context without obscuring primary state signals.
- `C2` [tests] The TUI shall preserve existing keyboard-driven managed service interactions while adding status markers and state details.
- `C3` [tests] Status markers shall be understandable using both symbol shape and text state so that service state remains recognizable across common terminal themes; the running state shall use a play marker (`▶`).

## Edge Cases

- `Edge-1` [tests] IF a crashed managed service has no captured logs, THEN the system shall still show its crashed state and present the best available failure context without leaving the details area blank.
- `Edge-2` [tests] IF a managed service was intentionally stopped by the user, THEN the system shall present it as stopped rather than crashed in the managed service view.
- `Edge-3` [tests] IF a managed service exits immediately after start, THEN the system shall surface a non-running state and the best available context in the managed service details area.
- `Edge-4` [tests] IF terminal width is insufficient for full managed service details, THEN the system shall truncate or compress details while preserving the service state marker and primary failure headline.
- `Edge-5` [tests] IF a managed service has empty or unset metadata fields (CWD, command, ports), THEN the system shall still render the details pane gracefully without displaying blank lines for missing fields.
- `Edge-6` [tests] IF a managed service has multiple ports, THEN the details pane shall render all ports in a compact format without duplicating the list-view port summary.
- `Edge-7` [tests] IF no service is selected in the details pane, the system shall hide or disable action buttons to prevent invalid actions.
- `Edge-8` [tests] IF a service is in a transition state (starting, stopping), the system shall handle action button state appropriately to prevent conflicting actions.

## Route Policy Summary

| Route | Count | IDs |
|---|---:|---|
| bdd | 10 | `BR-1`, `BR-2`, `BR-3`, `BR-4`, `BR-5`, `BR-6`, `BR-7`, `BR-8`, `BR-9`, `BR-10` |
| tests | 11 | `C1`, `C2`, `C3`, `Edge-1`, `Edge-2`, `Edge-3`, `Edge-4`, `Edge-5`, `Edge-6`, `Edge-7`, `Edge-8` |
| clarification | 0 | - |
| not_applicable | 0 | - |
