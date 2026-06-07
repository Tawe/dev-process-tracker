# Architecture

## Overview

This feature is implemented as a presentation-layer enhancement inside the existing TUI table flow.
The architecture moves managed services to a split interaction model: at normal widths, when a managed service is selected, the managed-services area renders as a 50|50 two-pane view with the list on one side and selected-service details on the other side.
The design intentionally avoids introducing a new process-lifecycle subsystem in this phase and instead composes existing service state and log data through the current TUI dependency boundary.

## Pattern

### Pattern Name
Split-pane state enrichment

### Rationale
- The user problem is primarily scanability and first-level triage inside the TUI.
- Keeping the list visible while showing details preserves orientation and speeds up repeated inspection across multiple services.
- A 50|50 split gives predictable spatial behavior and prevents the details pane from feeling like an afterthought.
- A persistent placeholder pane keeps the layout stable even when no service is selected.
- Existing service discovery already exposes enough state for running, stopped, and crashed rendering.
- Existing log access is sufficient for showing a short headline and recent tail context without redesigning process supervision.

## Runtime Flow

### Flow 1: Managed list state recognition
1. The TUI loads managed services and discovered server state.
2. Managed row rendering derives a current service state for each managed service.
3. The row renderer maps state to a symbol-plus-text presentation.
4. The managed list remains the first-level scan surface for state recognition.

### Flow 2: 50|50 split-pane selection rendering
1. Focus moves to the managed services list.
2. A managed service becomes the active selection.
3. At normal widths, the managed-services area renders as a 50|50 split view with the list retained on one side and the selected-service details on the other side.
4. The details pane header includes action buttons: [restart|start] [stop] [edit].
5. The details pane renders content in this order: state line (name + symbol + state), source, service metadata (CWD, port(s), command), then crash-specific context if applicable.
6. Service metadata fields (CWD, command, ports) are rendered conditionally — only when non-empty — and placed after the source line and before any crash-specific context.

### Flow 3: Action button rendering and interaction
1. The details pane header renders action buttons based on the selected service's current state.
2. For stopped services: show [start] [edit].
3. For running services: show [restart] [stop] [edit].
4. For crashed services: show [restart] [edit].
5. Buttons are styled with icons and colors: restart (↻), start (▶), stop (■), edit (✎).
6. Button clicks are detected via mouse position tracking within the header region.
7. Clicking a button triggers the corresponding service action command.

### Flow 4: Independent scrolling
1. The TUI maintains three separate viewport models: runningVP, managedListVP, managedDetailsVP.
2. Mouse scroll events are routed to the viewport under the cursor position.
3. Keyboard scroll events (up/down arrows, page up/down) affect only the currently focused section.
4. Scrolling one viewport does not change the scroll position of other viewports.
5. Selection changes do not reset scroll positions in other sections.

### Flow 5: Empty-selection placeholder rendering
1. The managed-services split view is visible but no managed service is currently selected.
2. The details pane remains visible instead of collapsing.
3. The pane renders a placeholder prompt that tells the user to select a managed service for inspection.
4. Action buttons are hidden or disabled when no service is selected.

### Flow 6: Failed-service context enrichment
1. The TUI loads managed services and discovered server state.
2. Managed row rendering derives a current service state for each managed service.
3. The row renderer maps state to a symbol-plus-text presentation.
4. The managed list remains the first-level scan surface for state recognition.

### Flow 2: 50|50 split-pane selection rendering
1. Focus moves to the managed services list.
2. A managed service becomes the active selection.
3. At normal widths, the managed-services area renders as a 50|50 split view with the list retained on one side and the selected-service details on the other side.
4. The details pane renders content in this order: state line (name + symbol + state), source, service metadata (CWD, port(s), command), then crash-specific context if applicable.
5. Service metadata fields (CWD, command, ports) are rendered conditionally — only when non-empty — and placed after the source line and before any crash-specific context.

### Flow 3: Empty-selection placeholder rendering
1. The managed-services split view is visible but no managed service is currently selected.
2. The details pane remains visible instead of collapsing.
3. The pane renders a placeholder prompt that tells the user to select a managed service for inspection.

### Flow 6: Failed-service context enrichment
1. The selected managed service resolves to failed or crashed state.
2. The details pane asks the TUI dependency boundary for best-available crash reason and latest log path.
3. Recent non-empty log lines are reduced to a compact tail for triage.
4. The crash-specific section shows the failure headline first, then log path, then recent log context. This section appears after service metadata in the details pane.

## Module Boundaries

### `pkg/cli/tui/table.go`
- Owns managed-service presentation in the TUI.
- Owns the split-pane managed-services layout and selected-service details rendering.
- Owns layout decisions for width fitting, 50|50 pane proportions at normal widths, selection highlighting, and visible detail density.

### `pkg/cli/tui/helpers.go`
- Owns reusable state-to-presentation mapping.
- Owns symbol/color selection helpers, placeholder text helpers, and compact log-tail shaping helpers.
- Must remain presentation-oriented rather than becoming a secondary data source.

### `pkg/cli/tui/deps.go`
- Owns the narrow TUI dependency contract.
- Must expose only the data the TUI needs for rendering and first-level diagnostics.

### `pkg/cli/tui_adapter.go`
- Owns adaptation from CLI application services into the TUI dependency contract.
- Bridges existing process-manager log access into TUI-safe methods.
- Must not introduce business logic beyond adaptation.

### `pkg/cli/tui/tui_ui_test.go`
- Owns UI regression coverage for managed-service rendering.
- Validates split-view rendering, placeholder rendering, symbol visibility, crash context visibility, and width-sensitive output behavior.

## Structure

```text
pkg/
├── cli/
│   ├── tui_adapter.go          # CLI-to-TUI runtime bridge
│   └── tui/
│       ├── deps.go             # TUI dependency contract
│       ├── helpers.go          # state/symbol/log/placeholder helpers
│       ├── table.go            # managed list + split details rendering
│       └── tui_ui_test.go      # UI regression coverage
```

## Module Boundaries and Ownership Rule

- The managed-services list remains the single source of selection.
- The details pane is a pure projection of the currently selected managed service including its operational metadata (working directory, port(s), command), or a placeholder when nothing is selected.
- Selection changes must update the details pane without changing the list interaction contract.
- The split-pane layout belongs to the TUI presentation layer and must not leak process-manager concerns into row rendering.

## Layout Rule

- At normal terminal widths, the managed-services area uses a 50|50 split between list and details.
- At narrow widths, the layout may compress or rebalance, but it must preserve:
  - visible state markers in the list
  - a visible details or placeholder pane
  - the primary failure headline when relevant

## Invariants

- Managed-service state must be visible directly in the managed list.
- When a managed service is selected, the list must remain visible beside the selected-service details pane.
- When no managed service is selected, the details pane must remain visible with a placeholder prompt.
- Symbol usage must be paired with text state so meaning does not depend on color alone.
- Service metadata must degrade gracefully when individual fields (CWD, command, ports) are empty or unset.
- Multi-port metadata must render in a compact format consistent with the list-view port presentation.
- Failure context must degrade gracefully when logs are unavailable.
- Existing keyboard-driven managed-service interactions must remain intact.
- Width pressure must remove secondary detail before removing primary state signals.
- Action buttons must be context-sensitive based on the selected service's state.
- Action buttons must be hidden or disabled when no service is selected.
- Action buttons must use icons that are recognizable independent of color (C3 compliance).
- Each section (running, managed list, details) must scroll independently without affecting other sections.
- Scroll position must persist per section across selection changes.

## Runtime vs Test Scaffolding Separation

- Runtime presentation logic lives in `table.go`, `helpers.go`, `deps.go`, and `tui_adapter.go`.
- Output assertions and regression expectations live in `tui_ui_test.go`.
- No test-only helpers should leak into the runtime rendering path.

## E2E Decision

- No browser or API E2E framework exists for this TUI flow.
- Acceptance remains spec-first through BDD scenarios and Go-based TUI rendering tests.
- If the project later adds TUI automation or higher-level acceptance tooling, that layer should reuse the same managed-state and split-pane behavioral contract rather than redefining it.

## Extension Rule

Future deeper diagnostics must extend the selected-service details pane through the TUI dependency boundary.
They should not bypass the current boundary by embedding process-manager calls directly into table rendering or by mixing lifecycle-capture decisions into list-presentation helpers.

## Review Notes

- This architecture keeps ownership concentrated in the TUI presentation layer because the problem is UX-first.
- A later lifecycle/exit-cause improvement can enrich the same details pane without invalidating list semantics.
- Constraint carryover from requirements is handled here through 50|50 split preservation, empty-state stability, keyboard-preservation, and symbol-plus-text readability invariants.
