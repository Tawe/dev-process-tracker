# Requirements: DEVPT-004

**Source**: [DEVPT-004](../DEVPT-004-investigate-split-managed-services-and-detailed-st.md)
**Generated**: 2026-04-02

## Overview

This ticket improves how managed service state is communicated inside the managed-services area of the TUI.
The primary outcome is faster visual triage: users should be able to spot problematic services immediately and, once a service is selected, inspect its details in a 50|50 split view without leaving the managed services surface.
The requirements intentionally stay at the UX and behavior level and defer lifecycle-capture implementation choices to architecture.

## Constraint Carryover

| Constraint ID | Must Appear In |
|---------------|----------------|
| C1 | architecture.md (50|50 split and narrow-width fallback), tests.md (terminal width coverage), tasks.md (verification) |
| C2 | architecture.md (interaction preservation), tests.md (keyboard regression coverage), tasks.md (do-not-break scope) |
| C3 | architecture.md (status presentation rules), tests.md (theme/readability coverage), tasks.md (UX verification) |

## Semantic Decisions

- State recognition: the managed list must communicate state directly in the list view; users should not need a separate screen just to know whether a service is healthy or failed.
- Selection behavior: selecting a managed service should keep the list visible and open a side-by-side details pane rather than replacing the list with a separate full-width detail block.
- Split proportion: at normal terminal widths, the managed-services split should be treated as a 50|50 layout between list and details.
- Empty selection: when no service is selected, the details pane should remain visible and show a placeholder prompt rather than collapsing away.
- Service metadata: the details pane must display key operational metadata for the selected service — working directory, port(s), and command — so users can inspect service configuration at a glance without leaving the TUI. Metadata is a content concern separate from the split-view structure (BR-3).
- Failure context: the minimum useful failed-state context is a short headline plus recent log context when available; full diagnostics remain the responsibility of deeper views and later stages.
- Stopped vs. crashed: intentionally stopped services and failed services are different user-facing states and must not be conflated in the managed service experience.
- Narrow layouts: when width is limited, the state marker and primary failure signal take precedence over secondary details.
- Action buttons: the details pane header must include context-sensitive action buttons (start/restart, stop, edit) that allow users to manage the selected service directly from the details view.
- Button styling: action buttons must use icons and colors to make them visually distinct and recognizable across different terminal color schemes.
- Independent scrolling: each section (running processes, managed list, details pane) must scroll independently to allow users to inspect long lists or detailed information without losing context in other sections.

## Review Notes

- The requirement set remains brief because the CR already defines the user-facing goal clearly.
- BDD should focus on list recognition, 50|50 split-view selection behavior, placeholder details behavior, and failed-service context visibility.
- Architecture should define exactly how the 50|50 split degrades gracefully when width is constrained.
- Service metadata display is additive to existing details-pane content and should not displace crash context or state information.
- BR-3 owns the split-view structural contract (50|50 layout, list persistence, details pane existence). BR-7 owns what content appears in the details pane.
- BR-8 owns action button rendering in the details header. BR-9 owns button context-sensitivity. BR-10 owns independent scrolling behavior.

---
Use `requirements.trace.md` for canonical requirement rows and route summaries.
