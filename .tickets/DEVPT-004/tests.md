# Tests: DEVPT-004

**Source**: canonical requirements, BDD, and architecture trace state
**Generated**: 2026-04-02

## Module → Test Mapping

| Module | Test File | Purpose |
|--------|-----------|---------|
| `pkg/cli/tui/table.go` + `pkg/cli/tui/helpers.go` | `pkg/cli/tui/tui_managed_split_test.go` | Split view rendering, placeholder pane, stopped vs crashed semantics, narrow-width signal preservation, selected managed-row full-line highlight coverage, and service metadata display |
| `pkg/cli/tui/table.go` + `pkg/cli/tui/helpers.go` | `pkg/cli/tui/tui_ui_test.go` | Existing UI regression coverage for status markers and crash context rendering, including the running-state play marker |
| `pkg/cli/tui/update.go` + `pkg/cli/tui/helpers.go` | `pkg/cli/tui/tui_viewport_test.go` | Mouse interaction regression coverage for exact row selection in running and managed sections, including viewport-offset cases |
| `pkg/cli/tui/update.go` interaction contract | `pkg/cli/tui/tui_state_test.go` | Keyboard interaction regression coverage for managed-service navigation |
| `pkg/cli/tui_adapter.go` + `pkg/cli/tui/deps.go` | `pkg/cli/tui_adapter_test.go` | Real integration coverage for latest managed log path bridging |

## Data Mechanism Tests

| Pattern | Module | Tests |
|---------|--------|-------|
| 50|50 split at normal width | `pkg/cli/tui/table.go` | selected service renders list pane + details pane with service metadata |
| Service metadata display | `pkg/cli/tui/table.go` | working directory, port(s), and command are visible in the details pane for the selected service |
| Service metadata with crash context | `pkg/cli/tui/table.go` | metadata appears before failure headline and log context when a crashed service is selected |
| Missing metadata degradation | `pkg/cli/tui/table.go` | details pane remains readable when individual metadata fields (CWD, command, ports) are empty or unset |
| Multi-port metadata display | `pkg/cli/tui/table.go` | multiple ports render in a compact format without duplicating the list-view port summary |
| Empty selection placeholder | `pkg/cli/tui/table.go` | placeholder remains visible when no managed service is selected |
| Failure context compaction | `pkg/cli/tui/helpers.go` | crash headline + recent tail remain visible without full log expansion |
| Narrow-width degradation | `pkg/cli/tui/table.go` | state marker and primary headline remain visible when width is constrained |
| Mouse row mapping | `pkg/cli/tui/helpers.go` + `pkg/cli/tui/update.go` | mouse clicks select the exact rendered row in both running and managed sections |
| Selected managed-row highlight | `pkg/cli/tui/table.go` | selected managed service row applies full-line highlight instead of symbol-only highlight |
| Running-state marker shape | `pkg/cli/tui/helpers.go` | running/active processes use a play marker (`▶`) instead of a tick |

## External Dependency Tests

| Dependency | Real Test | Behavior When Absent |
|------------|-----------|----------------------|
| Managed log files via process manager | `TestTUIAdapterLatestServiceLogPath_ReturnsManagedLogFile` | details pane must fall back to best available non-log context |

## Constraint Coverage

| Constraint ID | Test File | Tests |
|---------------|-----------|-------|
| C1 | `pkg/cli/tui/tui_managed_split_test.go` | split layout visibility, placeholder stability, narrow-width preservation, service metadata display, multi-port metadata format |
| C2 | `pkg/cli/tui/tui_state_test.go`, `pkg/cli/tui/tui_viewport_test.go` | managed keyboard interaction regression plus mouse row-selection preservation |
| C3 | `pkg/cli/tui/tui_ui_test.go`, `pkg/cli/tui/tui_managed_split_test.go` | symbol + text readability for managed state plus full-line selected-row highlight |

## Verification

```bash
# split-view and selected-row rendering coverage
go test ./pkg/cli/tui -run 'TestManagedSplitView|TestManagedSplitView_SelectedManagedRowHighlightsWholeLine' -count=1

# service metadata coverage (new)
go test ./pkg/cli/tui -run 'TestManagedSplitView.*Metadata|TestManagedSplitView.*ServiceMetadata' -count=1

# mouse row-mapping coverage
go test ./pkg/cli/tui -run 'TestTableMouseClickSelection' -count=1

# integration coverage for adapter bridge
go test ./pkg/cli -run TestTUIAdapterLatestServiceLogPath -count=1
```

## Review Notes

- `pkg/cli/tui/tui_managed_split_test.go` started RED-first and now also covers selected managed-row full-line highlight behavior.
- `pkg/cli/tui/tui_viewport_test.go` now verifies that mouse clicks map to the exact rendered row in both running and managed sections, including viewport-offset cases.
- Existing keyboard and UI tests remain part of the coverage story; together they protect status-marker readability and managed-service interaction continuity across both keyboard and mouse input.
- Running-state presentation now uses a play marker (`▶`) to better communicate active process state while preserving symbol-plus-text readability.
- The ticket docs now reflect the implemented regression coverage; canonical `spec-trace` state should be refreshed separately if it must include the new mouse/highlight evidence.
