# UAT Refinement Brief

## Objective

Add interactive action buttons to the details pane header and ensure independent scrolling for all sections. This enables users to perform service actions directly from the details view without navigating away.

## Approved Changes

1. **Action buttons in details header**: The details pane title line now includes a button group for service actions.
   - Format: "Selected service details" → "Details {button_group}"
   - Buttons: `[restart|start] [stop]`
   - Context-sensitive: show "start" when stopped, "restart" when running/crashed
   - Note: Edit button removed from scope - deferred to future phase

2. **Button styling and icons**:
   - Restart: circular arrow icon (↻ or ⟳)
   - Start: play icon (▶)
   - Stop: stop icon (■ or ◼)
   - Colors: distinguishable from regular text, likely using lipgloss styling

3. **TUI library investigation findings**:
   - Charm bubbles v2.1.0 does not include a built-in button component
   - Buttons will be implemented as styled text with click detection
   - Will use lipgloss for styling and custom click handlers

4. **Details pane context scope**:
   - Details show information for the actively selected service
   - If selection is in the process (running) area, show details for that service
   - If selection is in managed services area, show details for the managed service
   - **IMPLEMENTED**: Discovered (non-managed) services show full details including PID, port, command, directory, project, start time, agent info, and health status

5. **Independent scrolling**:
   - All three viewports (running, managed list, managed details) must scroll independently
   - Already implemented via separate viewport models, needs verification and potential refinement

## Changed Requirement IDs

| ID | Action | Summary |
|----|--------|---------|
| BR-8 | additive_change | New: details pane header shall display action buttons for the selected service |
| BR-9 | additive_change | New: action buttons shall be context-sensitive based on service state |
| BR-10 | additive_change | New: each section (running, managed list, details) shall scroll independently |
| BR-11 | additive_change | New: details pane shows information for actively selected service, whether running or managed |
| Edge-7 | additive_change | New: action buttons shall be disabled/hidden when no service is selected |
| Edge-8 | additive_change | New: action buttons shall handle edge cases (service starting, service just stopped) |
| Edge-9 | additive_change | New: placeholder shown when no service selected in currently focused section |

## Affected Downstream Trace

| Stage | Impact |
|-------|--------|
| bdd | New scenarios: `details_header_shows_action_buttons`, `action_buttons_are_context_sensitive`, `sections_scroll_independently` |
| architecture | Flow updates for button rendering, click handling, viewport interaction routing |
| tests | New tests for button rendering, button state transitions, independent scroll behavior |
| tasks | New task: implement action buttons, verify independent scrolling |
| obligations | New: `OBL-action-button-rendering`, `OBL-independent-scroll` |
| artifacts | New: `ART-action-buttons-test-go` |

## Execution Slices

### Slice 1: Add action buttons to details header

**Objective**: Render action buttons in the details pane title line with appropriate styling and icons.

**Direct artifacts/files**:
- `pkg/cli/tui/table.go` — `renderManagedDetails()` function
- `pkg/cli/tui/helpers.go` — button rendering helpers
- `pkg/cli/tui/tui_action_buttons_test.go` — new test file

**Direct GREEN targets**:
- `details_header_shows_action_buttons` (BR-8)
- `action_buttons_are_context_sensitive` (BR-9)
- `TEST-action-buttons-ui`

**Impacted canonical task IDs**: TASK-5 (new)

**Why this slice exists**: Buttons are a new interaction mechanism. Need to render them first with proper styling before adding click handling. The Charm library has no built-in buttons, so we'll use styled text with click regions.

### Slice 2: Wire button click handling

**Objective**: Make buttons clickable/activatable via mouse and keyboard.

**Direct artifacts/files**:
- `pkg/cli/tui/model.go` — click and key handling
- `pkg/cli/tui/commands.go` — service action commands (start/stop/restart)
- `pkg/cli/tui/tui_action_buttons_test.go` — interaction tests

**Direct GREEN targets**:
- Button click triggers correct service action
- Button state updates after action

**Impacted canonical task IDs**: TASK-5

**Why this slice exists**: Rendering alone isn't enough; buttons must trigger actions. This slice connects button clicks to existing service management commands.

### Slice 3: Verify and refine independent scrolling

**Objective**: Ensure all three viewports scroll independently without interfering with each other.

**Direct artifacts/files**:
- `pkg/cli/tui/table.go` — viewport management
- `pkg/cli/tui/tui_viewport_test.go` — existing, may need expansion

**Direct GREEN targets**:
- `sections_scroll_independently` (BR-10)
- `TEST-independent-scroll`

**Impacted canonical task IDs**: TASK-6 (new)

**Why this slice exists**: The current implementation has separate viewports but we need to verify they scroll independently and handle edge cases correctly (e.g., mouse scroll in details shouldn't scroll list).

## Validation

```bash
# New button rendering tests
go test ./pkg/cli/tui -run 'TestActionButtons' -count=1

# Independent scroll tests
go test ./pkg/cli/tui -run 'TestViewport.*Independent' -count=1

# Full managed service test suite
go test ./pkg/cli/tui -run 'TestManagedSplitView|TestView_ManagedCrashContextAndSymbols|TestActionButtons' -count=1

# Integration test with buttons
go test ./pkg/cli/tui -run 'TestTUIKeySequence|TestTUISimpleUpdate' -count=1
```

## Watchlist

- Button styling must work across different terminal color schemes (C3 interaction)
- Button click regions must be accurately detected (no ±1 drift)
- Independent scrolling must not interfere with selection navigation
- Action buttons for discovered services need design decision (see Open Decisions)
- Button state during service transitions (starting, stopping) needs careful handling

## Open Decisions

### Decision 1: Discovered services details and actions

**Question**: What should the details pane show when a discovered (non-managed) service is selected in the running processes area?

**Options**:
1. Show basic info (PID, port, command) with no action buttons
2. Show basic info with "Add to managed" button instead of start/stop/restart
3. Hide details pane or show different placeholder
4. Show same details as managed services but disable action buttons

**Resolution**: Option 1 (enhanced) - Show comprehensive details including PID, port, command, directory, project, start time, agent info, health status, and crash information if applicable. Action buttons for discovered services are deferred to future implementation.

**Status**: **RESOLVED - Implemented in commit 097772f**

**Implementation details**:
- Renamed `managedDetailsVP` to `selectedDetailsVP` for semantic clarity
- Renamed `renderManagedDetails()` to `renderSelectedServiceDetails()`
- Added focus-based logic: running services show their own details, managed services show their details
- Discovered services show: Name, Source, Status, PID, Port/Protocol, Command, Directory, Project, Start time, Agent info, Health, Crash info
- Placeholder shown when no service is selected: "Select a running service to inspect details"

### Decision 2: Button keyboard shortcuts

**Decision**: Keyboard shortcuts already exist in the current keymap. No new shortcuts needed for buttons.

**Status**: **RESOLVED**

### Decision 3: Edit button action

**Decision**: Remove edit button from this phase. Defer to future implementation.

**Status**: **RESOLVED**
