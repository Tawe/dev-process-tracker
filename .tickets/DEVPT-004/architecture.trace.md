# Architecture

## Obligations

- Action buttons must render in details pane header with proper styling and context sensitivity (`OBL-action-buttons-in-header`)
  Derived From: `BR-8`, `BR-9`
  Artifacts: `ART-tui-table-go`, `ART-tui-helpers-go`, `ART-tui-action-buttons-test-go`
- Details pane must show crash headline and recent logs for crashed services (`OBL-crash-context-display`)
  Derived From: `BR-4`, `BR-5`
  Artifacts: `ART-tui-table-go`, `ART-tui-helpers-go`, `ART-tui-managed-split-test-go`
- Each viewport must maintain independent scroll state (`OBL-independent-viewport-scroll`)
  Derived From: `BR-10`
  Artifacts: `ART-tui-table-go`
- Details pane must display service metadata (CWD, ports, command) (`OBL-managed-service-metadata-display`)
  Derived From: `BR-7`
  Artifacts: `ART-tui-table-go`, `ART-tui-helpers-go`
- Details pane must show placeholder when no service selected (`OBL-placeholder-details`)
  Derived From: `BR-6`
  Artifacts: `ART-tui-table-go`, `ART-tui-helpers-go`
- Managed services must render as 50|50 split when service selected (`OBL-split-pane-rendering`)
  Derived From: `BR-3`
  Artifacts: `ART-tui-table-go`, `ART-tui-managed-split-test-go`
- TUI adapter must expose latest managed service log path (`OBL-tui-log-path-bridge`)
  Derived From: `BR-5`
  Artifacts: `ART-cli-tui-adapter-go`, `ART-tui-deps-go`

## Artifacts

| Artifact ID | Path | Kind | Referencing Obligations |
|---|---|---|---|
| `ART-cli-tui-adapter-go` | `pkg/cli/tui_adapter.go` | runtime | `OBL-tui-log-path-bridge` |
| `ART-tui-action-buttons-test-go` | `pkg/cli/tui/tui_action_buttons_test.go` | test | `OBL-action-buttons-in-header` |
| `ART-tui-deps-go` | `pkg/cli/tui/deps.go` | runtime | `OBL-tui-log-path-bridge` |
| `ART-tui-helpers-go` | `pkg/cli/tui/helpers.go` | runtime | `OBL-action-buttons-in-header`, `OBL-crash-context-display`, `OBL-managed-service-metadata-display`, `OBL-placeholder-details` |
| `ART-tui-managed-split-test-go` | `pkg/cli/tui/tui_managed_split_test.go` | test | `OBL-crash-context-display`, `OBL-split-pane-rendering` |
| `ART-tui-table-go` | `pkg/cli/tui/table.go` | runtime | `OBL-action-buttons-in-header`, `OBL-crash-context-display`, `OBL-independent-viewport-scroll`, `OBL-managed-service-metadata-display`, `OBL-placeholder-details`, `OBL-split-pane-rendering` |

## Derivation Summary

| Requirement ID | Obligation Count | Obligation IDs |
|---|---:|---|
| `BR-3` | 1 | `OBL-split-pane-rendering` |
| `BR-4` | 1 | `OBL-crash-context-display` |
| `BR-5` | 2 | `OBL-crash-context-display`, `OBL-tui-log-path-bridge` |
| `BR-6` | 1 | `OBL-placeholder-details` |
| `BR-7` | 1 | `OBL-managed-service-metadata-display` |
| `BR-8` | 1 | `OBL-action-buttons-in-header` |
| `BR-9` | 1 | `OBL-action-buttons-in-header` |
| `BR-10` | 1 | `OBL-independent-viewport-scroll` |
