# Test Plan

## Test Plans By Kind

### unit

- Action buttons render and respond correctly in details header (`TEST-action-buttons-ui`)
  Covers: `BR-8`, `BR-9`, `Edge-7`, `Edge-8`
  File: `pkg/cli/tui/tui_action_buttons_test.go`
- Services that exit immediately show proper context (`TEST-immediate-exit-service`)
  Covers: `Edge-3`
  File: `pkg/cli/tui/tui_managed_split_test.go`
- Sections scroll independently (`TEST-independent-scroll`)
  Covers: `BR-10`
  File: `pkg/cli/tui/tui_viewport_test.go`
- Managed keyboard interactions remain intact (`TEST-managed-keyboard-interactions`)
  Covers: `C2`
  File: `pkg/cli/tui/tui_state_test.go`
- Service metadata (CWD, ports, command) display in details pane (`TEST-managed-service-metadata-ui`)
  Covers: `BR-7`, `Edge-5`, `Edge-6`, `C1`
  File: `pkg/cli/tui/tui_managed_split_test.go`
- Managed split view renders correctly (`TEST-managed-split-view-ui`)
  Covers: `BR-3`, `BR-4`, `BR-5`, `BR-6`, `BR-7`, `C1`, `Edge-1`, `Edge-2`, `Edge-4`, `Edge-5`, `Edge-6`
  File: `pkg/cli/tui/tui_managed_split_test.go`
- Managed status markers render correctly (`TEST-managed-status-markers-ui`)
  Covers: `BR-1`, `BR-2`, `C3`
  File: `pkg/cli/tui/tui_ui_test.go`

## Requirement Coverage Summary

| Requirement ID | Route Policy | Direct Test Plans | Indirect Test Plans |
|---|---|---|---|
| `C1` | tests | `TEST-managed-service-metadata-ui`, `TEST-managed-split-view-ui` | - |
| `C2` | tests | `TEST-managed-keyboard-interactions` | - |
| `C3` | tests | `TEST-managed-status-markers-ui` | - |
| `Edge-1` | tests | `TEST-managed-split-view-ui` | - |
| `Edge-2` | tests | `TEST-managed-split-view-ui` | - |
| `Edge-3` | tests | `TEST-immediate-exit-service` | - |
| `Edge-4` | tests | `TEST-managed-split-view-ui` | - |
| `Edge-5` | tests | `TEST-managed-service-metadata-ui`, `TEST-managed-split-view-ui` | - |
| `Edge-6` | tests | `TEST-managed-service-metadata-ui`, `TEST-managed-split-view-ui` | - |
| `Edge-7` | tests | `TEST-action-buttons-ui` | - |
| `Edge-8` | tests | `TEST-action-buttons-ui` | - |
