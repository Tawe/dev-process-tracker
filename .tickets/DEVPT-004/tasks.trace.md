# Tasks

## Task List

- Wire managed diagnostics bridge (`TASK-1`)
  Owns: `ART-cli-tui-adapter-go`, `ART-tui-deps-go`
  Makes Green: `TEST-managed-split-view-ui`
- Build split selection pane (`TASK-2`)
  Owns: `ART-tui-helpers-go`, `ART-tui-table-go`
  Makes Green: `selected_service_shows_5050_split_pane`, `split_view_without_selection_shows_placeholder`, `TEST-managed-keyboard-interactions`
- Finish crash context and UI regression coverage (`TASK-3`)
  Owns: `ART-tui-helpers-go`, `ART-tui-managed-split-test-go`, `ART-tui-table-go`
  Makes Green: `crashed_service_shows_failure_headline`, `crashed_service_shows_recent_log_context`, `managed_list_shows_state_markers`, `TEST-immediate-exit-service`, `TEST-managed-split-view-ui`, `TEST-managed-status-markers-ui`
- Add service metadata to details pane (`TASK-4`)
  Owns: `ART-tui-helpers-go`, `ART-tui-managed-split-test-go`, `ART-tui-table-go`
  Makes Green: `crashed_service_shows_metadata_alongside_crash_context`, `selected_service_shows_service_metadata`, `TEST-managed-service-metadata-ui`
- Add action buttons to details pane header (`TASK-5`)
  Owns: `ART-tui-action-buttons-test-go`, `ART-tui-helpers-go`, `ART-tui-table-go`
  Makes Green: `action_buttons_are_context_sensitive`, `details_header_shows_action_buttons`, `TEST-action-buttons-ui`
- Verify and refine independent scrolling (`TASK-6`)
  Owns: `ART-tui-table-go`
  Makes Green: `sections_scroll_independently`, `TEST-independent-scroll`

## Artifact Ownership Summary

| Artifact ID | Owning Task IDs |
|---|---|
| `ART-cli-tui-adapter-go` | `TASK-1` |
| `ART-tui-action-buttons-test-go` | `TASK-5` |
| `ART-tui-deps-go` | `TASK-1` |
| `ART-tui-helpers-go` | `TASK-2`, `TASK-3`, `TASK-4`, `TASK-5` |
| `ART-tui-managed-split-test-go` | `TASK-3`, `TASK-4` |
| `ART-tui-table-go` | `TASK-2`, `TASK-3`, `TASK-4`, `TASK-5`, `TASK-6` |

## Makes Green Summary

| ID | Task IDs |
|---|---|
| `action_buttons_are_context_sensitive` | `TASK-5` |
| `crashed_service_shows_failure_headline` | `TASK-3` |
| `crashed_service_shows_metadata_alongside_crash_context` | `TASK-4` |
| `crashed_service_shows_recent_log_context` | `TASK-3` |
| `details_header_shows_action_buttons` | `TASK-5` |
| `managed_list_shows_state_markers` | `TASK-3` |
| `sections_scroll_independently` | `TASK-6` |
| `selected_service_shows_5050_split_pane` | `TASK-2` |
| `selected_service_shows_service_metadata` | `TASK-4` |
| `split_view_without_selection_shows_placeholder` | `TASK-2` |
| `TEST-action-buttons-ui` | `TASK-5` |
| `TEST-immediate-exit-service` | `TASK-3` |
| `TEST-independent-scroll` | `TASK-6` |
| `TEST-managed-keyboard-interactions` | `TASK-2` |
| `TEST-managed-service-metadata-ui` | `TASK-4` |
| `TEST-managed-split-view-ui` | `TASK-1`, `TASK-3` |
| `TEST-managed-status-markers-ui` | `TASK-3` |
