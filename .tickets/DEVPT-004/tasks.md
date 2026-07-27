# Tasks: DEVPT-004

**Source**: canonical architecture/tests/bdd state + `tasks.trace.md` for trace cross-checking

## Scope Boundaries

- Managed services split view: keep the managed list visible while rendering selected-service details beside it.
- Crash diagnostics in TUI: use existing best-available reason and latest-log bridge; do not redesign process supervision.
- Interaction preservation: keep current keyboard-driven managed-service actions intact while adding split-view behavior.
- Pointer interaction preservation: mouse clicks in both running and managed sections must resolve to the exact rendered row without ±1 drift.

## Ownership Guardrails

| Critical Behavior | Owner Module | Merge/Refactor Task if Overlap |
|-------------------|--------------|--------------------------------|
| latest managed log path bridge | `pkg/cli/tui_adapter.go` + `pkg/cli/tui/deps.go` | Task 1 |
| split managed-services layout | `pkg/cli/tui/table.go` | Task 2 |
| managed status presentation + crash details | `pkg/cli/tui/helpers.go` + `pkg/cli/tui/table.go` | Task 3 |

## Constraint Coverage

| Constraint ID | Tasks |
|---------------|-------|
| C1 | Task 2, Task 3, Task 4, Task 7 |
| C2 | Task 2, Task 3, Task 7 |
| C3 | Task 3, Task 7 |

## Milestones

| Milestone | BDD Scenarios | Tasks | Checkpoint |
|-----------|---------------|-------|------------|
| M1: diagnostics bridge | — | Task 1 | `TEST-tui-adapter-log-path-bridge` GREEN |
| M2: split selection UX | `selected_service_shows_5050_split_pane`, `split_view_without_selection_shows_placeholder` | Task 2 | split pane + placeholder behavior GREEN |
| M3: crash context and final polish | `managed_list_shows_state_markers`, `crashed_service_shows_failure_headline`, `crashed_service_shows_recent_log_context` | Task 3 | status markers + crash context GREEN |
| M4: service metadata | `selected_service_shows_service_metadata`, `crashed_service_shows_metadata_alongside_crash_context` | Task 4 | service metadata (CWD, ports, command) GREEN |
| M5: action buttons | `details_header_shows_action_buttons`, `action_buttons_are_context_sensitive` | Task 5 | action buttons in details header GREEN |
| M6: independent scrolling | `sections_scroll_independently` | Task 6 | independent scroll verification GREEN |
| M7: universal details pane | `running_service_shows_details_in_details_pane`, `managed_service_shows_details_in_details_pane`, `no_selection_shows_placeholder` | Task 7 | universal details pane GREEN |

## Architecture Coverage

| Layer | Arch Files | In Tasks | Gap | Status |
|-------|-----------:|---------:|----:|--------|
| `pkg/cli/` | 1 | 1 | 0 | ✅ |
| `pkg/cli/tui/` | 4 | 4 | 0 | ✅ |

## Tasks

### Task 1: Wire managed diagnostics bridge

**Milestone**: M1 — diagnostics bridge

**Structure**: `pkg/cli/tui_adapter.go`, `pkg/cli/tui/deps.go`

**Makes GREEN (Automated Tests)**:
- `TEST-tui-adapter-log-path-bridge` → `pkg/cli/tui_adapter_test.go`: latest managed log path is exposed to the TUI

**Scope**: Finalize the TUI-facing dependency bridge for managed log-path access and keep the boundary narrow.
**Boundary**: Adapter and dependency interface only; no table layout work in this task.

**Creates**:
- none

**Modifies**:
- `pkg/cli/tui_adapter.go`
- `pkg/cli/tui/deps.go`

**Must Not Touch**:
- `pkg/cli/tui/table.go`
- `pkg/cli/tui/helpers.go`
- `pkg/cli/tui/tui_ui_test.go`

**Create/Move**:
- Ensure the TUI dependency contract exposes latest managed log path access
- Keep CLI-to-TUI adaptation free of presentation logic

**Exclude**: No split-layout rendering, no placeholder copy, no status-marker changes.

**Anti-duplication**: Reuse existing process-manager log-path APIs through the adapter — do NOT create a parallel log lookup path.

**Duplication Guard**:
- Check whether log-path lookup already exists in the CLI application layer before adding helpers
- If adaptation logic duplicates process-manager behavior, merge back into the adapter instead of creating another runtime owner

**Verify**:
```bash
go test ./pkg/cli -run TestTUIAdapterLatestServiceLogPath -count=1
```

**Done when**:
- [x] `TEST-tui-adapter-log-path-bridge` is GREEN
- [x] Adapter boundary stays presentation-free
- [x] No duplicate log lookup path is introduced

### Task 2: Build split selection pane

**Milestone**: M2 — split selection UX (BR-3, BR-6)

**Structure**: `pkg/cli/tui/table.go`, `pkg/cli/tui/helpers.go`

**Makes GREEN (Automated Tests)**:
- `TEST-managed-keyboard-interactions` → `pkg/cli/tui/tui_state_test.go`: managed keyboard interaction regression remains GREEN

**Makes GREEN (Behavior)**:
- `selected_service_shows_5050_split_pane` → `pkg/cli/tui/tui_managed_split_test.go` (BR-3)
- `split_view_without_selection_shows_placeholder` → `pkg/cli/tui/tui_managed_split_test.go` (BR-6)

**Scope**: Implement the managed-services 50|50 split behavior and stable placeholder details pane.
**Boundary**: Split layout, selected-service projection, and placeholder rendering only.

**Creates**:
- none

**Modifies**:
- `pkg/cli/tui/table.go`
- `pkg/cli/tui/helpers.go`

**Must Not Touch**:
- `pkg/cli/tui_adapter.go`
- `pkg/cli/tui/deps.go`
- `pkg/cli/tui/tui_ui_test.go`

**Create/Move**:
- Render the managed-services area as a two-pane 50|50 layout at normal widths
- Keep the list visible while switching the details side based on selection state
- Show a placeholder prompt when no managed service is selected
- Preserve exact mouse row-to-selection mapping while introducing the split layout

**Exclude**: No process-lifecycle redesign, no new CLI status output work, no deeper crash-cause model.

**Anti-duplication**: Use shared helper functions for selection-state rendering and placeholder copy — do NOT duplicate pane logic in multiple render branches.

**Duplication Guard**:
- Check existing managed-section rendering paths before adding pane branches
- If similar placeholder logic exists elsewhere, merge into shared helpers rather than creating a second placeholder owner

**Verify**:
```bash
go test ./pkg/cli/tui -run 'TestManagedSplitView|TestTUISimpleUpdate|TestTUIKeySequence' -count=1
```

**Done when**:
- [x] `selected_service_shows_5050_split_pane` is GREEN
- [x] `split_view_without_selection_shows_placeholder` is GREEN
- [x] `TEST-managed-keyboard-interactions` remains GREEN
- [x] Mouse clicks select the exact rendered managed row without ±1 drift
- [x] The list stays visible beside the details pane
- [x] No duplicate selection-rendering path is introduced

### Task 3: Finish crash context and UI regression coverage

**Milestone**: M3 — crash context and final polish (BR-1, BR-2, BR-4, BR-5)

**Structure**: `pkg/cli/tui/table.go`, `pkg/cli/tui/helpers.go`, `pkg/cli/tui/tui_ui_test.go`

**Makes GREEN (Automated Tests)**:
- `TEST-managed-status-markers-ui` → `pkg/cli/tui/tui_ui_test.go`: state markers and crash context rendering
- `TEST-managed-split-view-ui` → `pkg/cli/tui/tui_managed_split_test.go`: split-view crash details and narrow-width preservation

**Makes GREEN (Behavior)**:
- `managed_list_shows_state_markers` → `pkg/cli/tui/tui_ui_test.go` (BR-1, BR-2)
- `crashed_service_shows_failure_headline` → `pkg/cli/tui/tui_managed_split_test.go` (BR-4)
- `crashed_service_shows_recent_log_context` → `pkg/cli/tui/tui_managed_split_test.go` (BR-5)

**Scope**: Complete status-symbol presentation, crash headline rendering, recent log context rendering, and final regression coverage.
**Boundary**: Final managed-service presentation and tests only.

**Creates**:
- none

**Modifies**:
- `pkg/cli/tui/table.go`
- `pkg/cli/tui/helpers.go`
- `pkg/cli/tui/tui_ui_test.go`

**Must Not Touch**:
- `pkg/cli/tui_adapter.go`
- `pkg/cli/tui/deps.go`

**Create/Move**:
- Ensure symbols and text state remain visible together in the managed list
- Use a play marker (`▶`) for running/active state presentation
- Render crash headline, log path, and compact recent tail in the details pane
- Expand UI regression coverage for standard and narrow widths
- Keep selected managed-row highlight applied to the full row, not only the state symbol

**Exclude**: No changes to process manager, registry schema, or CLI `status` output semantics.

**Anti-duplication**: Reuse existing crash-reason and log-tail helpers — do NOT add a second failure-summary formatter.

**Duplication Guard**:
- Check for existing crash-summary formatting before adding new presentation helpers
- If row rendering and details rendering diverge in state mapping, consolidate through shared helpers immediately

**Verify**:
```bash
go test ./pkg/cli/tui -run 'TestManagedSplitView|TestView_ManagedCrashContextAndSymbols|TestView_ManagedServicesSection' -count=1
```

**Done when**:
- [x] `managed_list_shows_state_markers` is GREEN
- [x] `crashed_service_shows_failure_headline` is GREEN
- [x] `crashed_service_shows_recent_log_context` is GREEN
- [x] `TEST-managed-status-markers-ui` is GREEN
- [x] `TEST-managed-split-view-ui` is GREEN
- [x] Selected managed-row highlight covers the full row
- [x] No duplicate crash-summary path is introduced

### Task 4: Add service metadata to details pane

**Milestone**: M4 — service metadata (BR-7, Edge-5, Edge-6)

**Structure**: `pkg/cli/tui/table.go`, `pkg/cli/tui/helpers.go`

**Makes GREEN (Automated Tests)**:
- `TEST-managed-service-metadata-ui` → `pkg/cli/tui/tui_managed_split_test.go`: working directory, port(s), and command are visible in the details pane

**Makes GREEN (Behavior)**:
- `selected_service_shows_service_metadata` → `pkg/cli/tui/tui_managed_split_test.go` (BR-7)

**Scope**: Display operational metadata (working directory, port(s), command) in the details pane for the selected managed service.
**Boundary**: Details pane content only; no layout or interaction changes.

**Creates**:
- none

**Modifies**:
- `pkg/cli/tui/table.go`
- `pkg/cli/tui/helpers.go`

**Must Not Touch**:
- `pkg/cli/tui_adapter.go`
- `pkg/cli/tui/deps.go`
- `pkg/cli/tui/tui_ui_test.go`

**Create/Move**:
- Render working directory (`CWD`), port(s), and command from `ManagedService` in the details pane
- Place metadata after the state line and before any crash-specific context
- Omit individual fields gracefully when empty or unset (no blank lines)

**Exclude**: No changes to split layout proportions, no new dependency bridge methods, no process-lifecycle changes.

**Anti-duplication**: Reuse existing `fitLine` and `ManagedService` data access — do NOT create a separate metadata formatter.

**Duplication Guard**:
- Check whether metadata fields are already accessible through `selectedManagedService()` or `serverInfoForService()` before adding new access patterns
- If similar field rendering exists elsewhere (e.g., running service details), align the format

**Verify**:
```bash
go test ./pkg/cli/tui -run 'TestManagedSplitView.*Metadata|TestManagedSplitView.*ServiceMetadata' -count=1
```

**Done when**:
- [x] `selected_service_shows_service_metadata` is GREEN
- [x] `crashed_service_shows_metadata_alongside_crash_context` is GREEN
- [x] `TEST-managed-service-metadata-ui` is GREEN
- [x] `TEST-managed-split-view-ui` remains GREEN
- [x] Empty CWD, command, or ports do not produce blank lines (Edge-5)
- [x] Multi-port metadata renders compactly (Edge-6)
- [x] Metadata appears after source and before crash context
- [x] No duplicate metadata access pattern is introduced

### Task 5: Add action buttons to details pane header

**Milestone**: M5 — action buttons (BR-8, BR-9, Edge-7, Edge-8)

**Structure**: `pkg/cli/tui/table.go`, `pkg/cli/tui/helpers.go`, `pkg/cli/tui/model.go`, `pkg/cli/tui/commands.go`

**Makes GREEN (Automated Tests)**:
- `TEST-action-buttons-ui` → `pkg/cli/tui/tui_action_buttons_test.go`: action buttons render correctly with proper styling and icons

**Makes GREEN (Behavior)**:
- `details_header_shows_action_buttons` → `pkg/cli/tui/tui_action_buttons_test.go` (BR-8)
- `action_buttons_are_context_sensitive` → `pkg/cli/tui/tui_action_buttons_test.go` (BR-9)

**Scope**: Render action buttons in the details pane header and wire them to service actions.
**Boundary**: Details pane header only; button click handling and action triggering.

**Creates**:
- `pkg/cli/tui/tui_action_buttons_test.go`

**Modifies**:
- `pkg/cli/tui/table.go`
- `pkg/cli/tui/helpers.go`
- `pkg/cli/tui/model.go`
- `pkg/cli/tui/commands.go`

**Must Not Touch**:
- `pkg/cli/tui_adapter.go`
- `pkg/cli/tui/deps.go`

**Create/Move**:
- Add button rendering logic to `renderManagedDetails()` header
- Create button style helpers with icons and colors
- Add button click detection via mouse position tracking
- Wire button clicks to existing service action commands (start/stop/restart)
- Handle context-sensitive button visibility (start vs restart)

**Exclude**: No changes to process manager, no new service actions, no keyboard shortcut implementation (unless specified).

**Anti-duplication**: Reuse existing service action commands — do NOT create parallel action paths.

**Duplication Guard**:
- Check existing command structure before adding new action handlers
- If similar button logic exists elsewhere, consolidate into shared helpers

**Verify**:
```bash
go test ./pkg/cli/tui -run 'TestActionButtons' -count=1
```

**Done when**:
- [ ] `details_header_shows_action_buttons` is GREEN
- [ ] `action_buttons_are_context_sensitive` is GREEN
- [ ] `TEST-action-buttons-ui` is GREEN
- [ ] Buttons show correct icons and colors
- [ ] Button clicks trigger correct service actions
- [ ] Buttons are hidden/disabled when no service is selected (Edge-7)
- [ ] Buttons handle transition states correctly (Edge-8)

### Task 6: Verify and refine independent scrolling

**Milestone**: M6 — independent scrolling (BR-10)

**Structure**: `pkg/cli/tui/table.go`, `pkg/cli/tui/tui_viewport_test.go`

**Makes GREEN (Automated Tests)**:
- `TEST-independent-scroll` → `pkg/cli/tui/tui_viewport_test.go`: each section scrolls independently

**Makes GREEN (Behavior)**:
- `sections_scroll_independently` → `pkg/cli/tui/tui_viewport_test.go` (BR-10)

**Scope**: Verify that all three viewports scroll independently and refine if necessary.
**Boundary**: Viewport interaction routing only.

**Creates**:
- none

**Modifies**:
- `pkg/cli/tui/table.go`
- `pkg/cli/tui/tui_viewport_test.go`

**Must Not Touch**:
- `pkg/cli/tui_adapter.go`
- `pkg/cli/tui/deps.go`
- `pkg/cli/tui/helpers.go`

**Create/Move**:
- Verify mouse scroll routing to correct viewport
- Verify keyboard scroll affects only focused section
- Add tests for independent scroll behavior
- Fix any cross-viewport scroll interference

**Exclude**: No layout changes, no selection changes, no content changes.

**Anti-duplication**: Use existing viewport models — do NOT create new scroll mechanisms.

**Duplication Guard**:
- Check existing viewport update logic before adding new handlers
- Ensure scroll routing doesn't duplicate selection routing

**Verify**:
```bash
go test ./pkg/cli/tui -run 'TestViewport.*Independent' -count=1
```

**Done when**:
- [ ] `sections_scroll_independently` is GREEN
- [ ] `TEST-independent-scroll` is GREEN
- [ ] Scrolling one section does not affect others
- [ ] Scroll positions persist across selection changes
- [ ] Mouse scroll routed to correct viewport
- [ ] Keyboard scroll affects only focused section

### Task 7: Make details pane universal for running and managed services

**Milestone**: M7 — universal details pane (BR-11, Edge-9)

**Structure**: `pkg/cli/tui/table.go`, `pkg/cli/tui/helpers.go`

**Makes GREEN (Automated Tests)**:
- `TEST-universal-details-pane-ui` → `pkg/cli/tui/tui_managed_split_test.go`: details pane shows appropriate content based on focus and selection

**Makes GREEN (Behavior)**:
- `running_service_shows_details_in_details_pane` → `pkg/cli/tui/tui_managed_split_test.go` (BR-11)
- `managed_service_shows_details_in_details_pane` → `pkg/cli/tui/tui_managed_split_test.go` (BR-11)
- `no_selection_shows_placeholder` → `pkg/cli/tui/tui_managed_split_test.go` (Edge-9)

**Scope**: Make the details pane work for both running and managed services based on which section is currently focused.
**Boundary**: Details pane content selection only; no layout changes.

**Creates**:
- none

**Modifies**:
- `pkg/cli/tui/table.go`

**Must Not Touch**:
- `pkg/cli/tui_adapter.go`
- `pkg/cli/tui/deps.go`
- `pkg/cli/tui/helpers.go` (selection logic only)

**Create/Move**:
- Rename `managedDetailsVP` to `selectedDetailsVP` for semantic clarity
- Rename `renderManagedDetails()` to `renderSelectedServiceDetails()`
- Add focus-based logic to determine which service details to show:
  - When `focusRunning` and a running service is selected, show running service details (PID, port, command, directory, project, start time, agent info, health, crash info)
  - When `focusManaged` and a managed service is selected, show managed service details (existing behavior)
  - When no service is selected in the focused section, show appropriate placeholder

**Exclude**: No layout changes, no viewport structure changes, no new data models.

**Anti-duplication**: Reuse existing service information access methods — do NOT create parallel lookup paths.

**Duplication Guard**:
- Check existing service data access patterns before adding new helpers
- Ensure running service details don't duplicate managed service details rendering
- Maintain single source of truth for service information display

**Verify**:
```bash
go test ./pkg/cli/tui -run 'TestManagedSplitView|TestUniversalDetailsPane' -count=1
```

**Done when**:
- [x] `running_service_shows_details_in_details_pane` is GREEN
- [x] `managed_service_shows_details_in_details_pane` is GREEN
- [x] `no_selection_shows_placeholder` is GREEN
- [x] `TEST-universal-details-pane-ui` is GREEN
- [x] Running services show PID, port, command, directory, project, start time, agent info, health, crash info
- [x] Managed services maintain existing details display
- [x] Placeholder shown when no service selected in focused section (Edge-9)
- [x] No duplicate service data access patterns introduced
- [x] Semantic naming matches universal behavior

## Post-Implementation

- [x] No duplication (grep check)
- [x] Scope boundaries respected
- [x] All unit/integration tests GREEN
- [x] All BDD scenarios GREEN
- [ ] Smoke test passes through the TUI flow
- [x] Fallback and narrow-width behavior matches requirements
- [x] Universal details pane implemented (commit 097772f)
- [x] Running services show comprehensive details
- [x] Managed services maintain existing details behavior
- [x] Placeholder shows appropriate message based on focus
