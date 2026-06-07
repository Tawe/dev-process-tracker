# BDD

## Scenarios By Requirement Family

### BR-1

- Managed list shows service state markers (`managed_list_shows_state_markers`)
  Covers: `BR-1`, `BR-2`
  Given: the managed services list is displayed in the TUI
  When: managed services exist in various states (running, starting, stopped, crashed)
  Then: each service shows a distinct visual state marker (▶ for running, appropriate markers for other states)

### BR-2

- Managed list shows service state markers (`managed_list_shows_state_markers`)
  Covers: `BR-1`, `BR-2`
  Given: the managed services list is displayed in the TUI
  When: managed services exist in various states (running, starting, stopped, crashed)
  Then: each service shows a distinct visual state marker (▶ for running, appropriate markers for other states)

### BR-3

- Crashed service shows failure headline (`crashed_service_shows_failure_headline`)
  Covers: `BR-3`, `BR-4`
  Given: a crashed managed service is selected
  When: the details pane renders
  Then: a concise failure headline appears using the best available reason
- Crashed service shows recent log context (`crashed_service_shows_recent_log_context`)
  Covers: `BR-3`, `BR-5`
  Given: a crashed managed service with recent logs is selected
  When: the details pane renders
  Then: recent log context appears to help users triage the failure
- Selected managed service shows 50|50 split details pane (`selected_service_shows_5050_split_pane`)
  Covers: `BR-3`
  Given: a managed service is selected in the TUI
  When: the managed services area renders
  Then: the area splits 50|50 with the list on one side and service details on the other

### BR-4

- Crashed service shows failure headline (`crashed_service_shows_failure_headline`)
  Covers: `BR-3`, `BR-4`
  Given: a crashed managed service is selected
  When: the details pane renders
  Then: a concise failure headline appears using the best available reason
- Crashed service shows metadata alongside crash context (`crashed_service_shows_metadata_alongside_crash_context`)
  Covers: `BR-7`, `BR-4`, `BR-5`
  Given: a crashed managed service with metadata is selected
  When: the details pane renders
  Then: metadata appears before crash context in the details pane

### BR-5

- Crashed service shows metadata alongside crash context (`crashed_service_shows_metadata_alongside_crash_context`)
  Covers: `BR-7`, `BR-4`, `BR-5`
  Given: a crashed managed service with metadata is selected
  When: the details pane renders
  Then: metadata appears before crash context in the details pane
- Crashed service shows recent log context (`crashed_service_shows_recent_log_context`)
  Covers: `BR-3`, `BR-5`
  Given: a crashed managed service with recent logs is selected
  When: the details pane renders
  Then: recent log context appears to help users triage the failure

### BR-6

- Split view without selection shows placeholder (`split_view_without_selection_shows_placeholder`)
  Covers: `BR-6`
  Given: the managed services split view is visible
  When: no managed service is selected
  Then: the details pane shows a placeholder prompting the user to select a service

### BR-7

- Crashed service shows metadata alongside crash context (`crashed_service_shows_metadata_alongside_crash_context`)
  Covers: `BR-7`, `BR-4`, `BR-5`
  Given: a crashed managed service with metadata is selected
  When: the details pane renders
  Then: metadata appears before crash context in the details pane
- Selected managed service shows service metadata (`selected_service_shows_service_metadata`)
  Covers: `BR-7`
  Given: a managed service is selected in the split view
  When: the details pane renders
  Then: the service's working directory, port(s), and command are displayed

### BR-8

- Details header shows action buttons (`details_header_shows_action_buttons`)
  Covers: `BR-8`
  Given: a managed service is selected in the TUI
  When: the details pane header renders
  Then: action buttons (start/restart, stop) appear in the details pane header

### BR-9

- Action buttons are context-sensitive (`action_buttons_are_context_sensitive`)
  Covers: `BR-9`
  Given: a managed service is selected with a specific state (stopped, running, or crashed)
  When: the details pane header renders action buttons
  Then: buttons show appropriate actions for the service state (start for stopped, restart for running/crashed, stop for running)

### BR-10

- Sections scroll independently (`sections_scroll_independently`)
  Covers: `BR-10`
  Given: the TUI displays running processes, managed list, and details pane
  When: user scrolls in one section
  Then: only that section scrolls while others maintain their scroll position

## Coverage Summary

| Requirement ID | Scenario Count | Scenario IDs |
|---|---:|---|
| `BR-1` | 1 | `managed_list_shows_state_markers` |
| `BR-2` | 1 | `managed_list_shows_state_markers` |
| `BR-3` | 3 | `crashed_service_shows_failure_headline`, `crashed_service_shows_recent_log_context`, `selected_service_shows_5050_split_pane` |
| `BR-4` | 2 | `crashed_service_shows_failure_headline`, `crashed_service_shows_metadata_alongside_crash_context` |
| `BR-5` | 2 | `crashed_service_shows_metadata_alongside_crash_context`, `crashed_service_shows_recent_log_context` |
| `BR-6` | 1 | `split_view_without_selection_shows_placeholder` |
| `BR-7` | 2 | `crashed_service_shows_metadata_alongside_crash_context`, `selected_service_shows_service_metadata` |
| `BR-8` | 1 | `details_header_shows_action_buttons` |
| `BR-9` | 1 | `action_buttons_are_context_sensitive` |
| `BR-10` | 1 | `sections_scroll_independently` |
