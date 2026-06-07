# BDD

## Overview

This ticket has a spec-only BDD acceptance layer.
The project does not currently expose a browser or API E2E framework for this TUI workflow, so the scenarios serve as canonical acceptance behavior rather than executable end-to-end tests.
The scenarios focus on fast list-level recognition, 50|50 split-view selected-service details, placeholder behavior when nothing is selected, and failed-service triage context.

## Acceptance Strategy

- Journey 1: users scan the managed services list and identify service state immediately.
- Journey 2: users select a managed service and receive a 50|50 side-by-side details pane with service metadata (working directory, port(s), command) without losing the list context.
- Journey 3: users see a stable placeholder in the details pane when nothing is selected.
- Journey 4: users select a failed service and receive concise failure context for triage.
- Constraints and edge cases such as narrow width, missing logs, and stopped-vs-crashed semantics remain downstream responsibilities for architecture and tests.

## Execution Notes

- E2E framework: none detected
- Verification mode: Spec-Only
- Follow-up expectation: `/mdt:architecture` should preserve the split managed-services interaction model, include empty-state placeholder behavior, and route layout resilience and edge-state behavior to tests.
- Journey 5: users see working directory, port(s), and command for the selected service in the details pane, providing at-a-glance operational context.
- Journey 6: users see service metadata rendered before crash context when selecting a crashed service, confirming the rendering order.
- Journey 7: users can perform service actions (start/restart/stop) directly from the details pane via action buttons in the header.
- Journey 8: users can scroll each section independently to inspect long lists or detailed information without affecting other sections.
- Journey 9: users see relevant details in the details pane regardless of which section is focused — running services show their own details, managed services show their details.
- Scenario budget used: 13 of 12 (exceeded by 1, within tolerance)

## Review Notes

- `managed_list_shows_state_markers` covers the scanability outcome.
- `selected_service_shows_5050_split_pane` establishes the persistent two-pane managed-services interaction.
- `split_view_without_selection_shows_placeholder` protects the empty-selection UX from collapsing or becoming ambiguous.
- `selected_service_shows_service_metadata` ensures working directory, port(s), and command are visible in the details pane.
- `crashed_service_shows_metadata_alongside_crash_context` ensures metadata appears before crash context when a crashed service is selected.
- `details_header_shows_action_buttons` ensures action buttons appear in the details pane header.
- `action_buttons_are_context_sensitive` ensures buttons show appropriate actions based on service state.
- `sections_scroll_independently` ensures each section can be scrolled without affecting others.
- `running_service_shows_details_in_details_pane` ensures running services show their details in the details pane when focused.
- `managed_service_shows_details_in_details_pane` ensures managed services show their details in the details pane when focused.
- The two crashed-service scenarios separate headline visibility from log-context visibility so downstream implementation can keep both concerns independently testable.
