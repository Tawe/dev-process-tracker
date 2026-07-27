---
code: DEVPT-004
status: Implemented
dateCreated: 2026-04-02T08:50:48.448Z
type: Feature Enhancement
priority: High
relatedTickets: DEVPT-003
---

# Managed service state panel in TUI

## 1. Description

### Requirements Scope
`brief`

### Problem
- Managed services in the TUI are harder to scan than the CLI status output.
- Users cannot quickly distinguish healthy, stopped, starting, and failed services from the managed list alone.
- When a managed service stops unexpectedly, the TUI does not provide enough immediate context to understand what happened.

### Affected Areas
- Terminal UI presentation of managed services
- Managed service status communication
- Runtime troubleshooting flow for local development services

### Scope
- **In scope**: Improve how managed service state is presented in the TUI, make failures easier to spot visually, and show concise diagnostic context for the selected service.
- **Out of scope**: Redesigning the CLI status command, changing process supervision architecture, or defining the final internal implementation approach for lifecycle tracking.

## 2. Desired Outcome

### Success Conditions
- Users can identify managed service state at a glance from the TUI list.
- Visually distinct symbols or markers communicate whether a managed service is running, starting, stopped, or failed.
- Selecting a managed service switches the managed-services area into a 50|50 split view that keeps the list visible and shows concise details for the selected service.
- When a managed service has failed, the TUI shows a short headline and recent log context similar in spirit to the CLI status experience.
- The managed services area remains easy to read in a terminal with limited width.

### Constraints
- Must fit within the existing terminal-based workflow.
- Must remain readable with color and symbol usage in common terminal themes.
- Must preserve current managed service interactions such as selection and service actions.
- Must not require users to leave the TUI for basic state recognition and first-level troubleshooting.

### Non-Goals
- Not redefining how managed services are started or stopped internally.
- Not requiring a full crash-analysis or process-history system in this phase.
- Not replacing the detailed CLI status output.
- Not introducing a new workflow that depends on mouse usage.

## 3. Open Questions

| Area | Question | Constraints |
|------|----------|-------------|
| UX | Which symbols and visual treatments are most readable for running, stopped, starting, and failed states? | Must remain understandable in terminal environments |
| Layout | How should the managed-services split view balance list width and details width while staying readable? | Must fit narrow terminal widths |
| Diagnostics | What is the minimum useful detail for a failed service in TUI view? | Should support first-level troubleshooting without overwhelming the screen |
| Consistency | How closely should the TUI details mirror CLI status output? | Must avoid duplicating excessive verbosity in the list view |

### Known Constraints
- The TUI should remain scan-friendly for multiple managed services.
- Status communication should work even when only limited runtime context is available.
- The design should support future deeper diagnostics without forcing another major UI rewrite.

### Decisions Deferred
- Exact implementation approach for deriving service stop and crash details.
- Specific internal data model changes for richer lifecycle tracking.
- Final layout mechanics for any grouped or divided managed service sections.
- Task breakdown and technical solution details.

## 4. Acceptance Criteria

### Functional (Outcome-focused)
- [ ] Managed services show a visually distinct state marker in the TUI.
- [ ] Users can distinguish running, stopped, starting, and failed managed services without opening another view.
- [ ] Selecting a managed service reveals a 50|50 split managed-services view with the list on one side and a compact details pane for the selected service on the other side.
- [ ] Failed managed services show a concise headline explaining the failure or best available reason.
- [ ] Failed managed services show recent log context that helps users understand the issue.
- [ ] The details view helps users decide whether to restart, inspect logs, or leave the service stopped.
- [ ] Details pane displays working directory, port(s), and command for the selected managed service.
- [ ] Details pane shows comprehensive details for selected running services (PID, port, command, directory, project, start time, agent info, health status).

### Non-Functional
- [ ] The managed services area remains readable in standard terminal sizes.
- [ ] Visual state markers are easy to recognize quickly.
- [ ] Additional details do not make the TUI feel cluttered during normal use.

### Edge Cases
- No managed service is currently selected.
- Managed service has no captured logs.
- Managed service was intentionally stopped by the user.
- Managed service exits immediately after start.
- Multiple managed services share similar names or ports.
- Terminal width is too narrow for full details.
- Managed service has multiple ports.
- Managed service has empty or unset metadata fields (CWD, command, ports).

## 5. Verification

### How to Verify Success
- Manual verification:
  - Open the TUI with managed services in mixed states.
  - Confirm users can identify problematic services at a glance.
  - Select running, stopped, and failed services and verify the details area changes appropriately.
  - Confirm failed services show a short headline and recent log context when available.
- Automated verification:
  - Validate rendering of managed service state markers.
  - Validate managed-service detail rendering for different service states.
  - Validate 50|50 split behavior at normal widths.
- Validate placeholder details behavior when nothing is selected.
- Validate narrow-width behavior and truncation.
- Usability verification:
  - Compare the TUI experience against the CLI status flow for identifying and triaging service failures.

> Requirements trace projection: [requirements.trace.md](./DEVPT-004/requirements.trace.md)
>
> Requirements notes: [requirements.md](./DEVPT-004/requirements.md)
>
> BDD trace projection: [bdd.trace.md](./DEVPT-004/bdd.trace.md)
>
> BDD notes: [bdd.md](./DEVPT-004/bdd.md)
>
> Architecture trace projection: [architecture.trace.md](./DEVPT-004/architecture.trace.md)
>
> Architecture notes: [architecture.md](./DEVPT-004/architecture.md)
>
> Tests trace projection: [tests.trace.md](./DEVPT-004/tests.trace.md)
>
> Tests notes: [tests.md](./DEVPT-004/tests.md)
>
> Tasks trace projection: [tasks.trace.md](./DEVPT-004/tasks.trace.md)
>
> Tasks notes: [tasks.md](./DEVPT-004/tasks.md)

---

## 8. Clarifications

### UAT Session 2026-04-06

**Approved changes**:
- Details pane header shall display action buttons for service management (start/restart, stop, edit)
- Action buttons shall be context-sensitive based on service state
- Action buttons shall use icons and colors: restart (↻), start (▶), stop (■), edit (✎)
- Each section (running processes, managed list, details) shall scroll independently
- Charm bubbles v2.1.0 has no built-in button component; buttons will be custom-styled text elements

**Changed requirement IDs**:
- BR-8: added (details pane header shows action buttons)
- BR-9: added (action buttons are context-sensitive)
- BR-10: added (sections scroll independently)
- Edge-7: added (action buttons hidden/disabled when no service selected)
- Edge-8: added (action buttons handle transition states)

**Updated workflow documents**:
- `requirements.md` — new semantic decisions for action buttons and independent scrolling
- `bdd.md` — Journey 7 and 8 added, 3 new scenarios, scenario budget updated (10/12)
- `architecture.md` — Flow 3 (action buttons), Flow 4 (independent scrolling) added, new invariants
- `tests.md` — new test coverage for action buttons and independent scroll
- `tasks.md` — TASK-5 and TASK-6 added, M5 and M6 milestones added

**uat.md**: written (replaced previous version)

**Strict drift/lock**: not used

**Resolved decisions**:
- Button keyboard shortcuts: Already exist in current keymap - no new shortcuts needed
- Edit button action: Removed from scope - deferred to future phase
- Discovered services details and actions: Recommended approach documented in uat.md (show details with "Add to managed" button)

### UAT Session 2026-04-03

**Approved changes**:
- Details pane shall display working directory, port(s), and command for the selected managed service.
- Metadata fields placed after state line and before crash context.
- Empty/unset fields omitted gracefully (no blank lines).
- Multi-port display format must be compact.
- Metadata must appear alongside crash context for crashed services.

**Changed requirement IDs**:
- BR-3: refined (broadened to include service metadata in details pane)
- BR-7: added (details pane displays CWD, ports, command)
- Edge-5: added (graceful degradation for missing metadata)
- Edge-6: added (multi-port compact display)

**Updated workflow documents**:
- `requirements.md` — semantic decision + review note added
- `bdd.md` — Journey 2 refined, Journey 5 added, scenario budget updated
- `architecture.md` — Flow 2 step 4-5 updated, new invariants
- `tests.md` — new data mechanism tests, C1 coverage expanded
- `tasks.md` — TASK-4 added, M4 milestone added

**uat.md**: written

**Strict drift/lock**: not used

### Spec Audit 2026-04-03

Resolved 14 issues found during cross-stage alignment review:
- Fixed BR-3/BR-7 overlap (BR-3 structural, BR-7 content)
- Fixed CR status (Implemented → In Progress)
- Fixed CR acceptance criteria (added metadata criterion)
- Fixed BDD scenario Given/When/Then format
- Added missing obligation `OBL-managed-service-metadata-display`
- Added missing artifact `ART-tui-managed-split-test-go`
- Added scenario `crashed_service_shows_metadata_alongside_crash_context`
- Added edge case Edge-6 (multi-port display)
- Clarified architecture Flow 4 wording
- Clarified source/metadata/crash render order
- Updated tests.md mapping, verification, and data mechanism tests
- Fixed TASK-4 milestone header (M3 → M4)
- Linked TASK-4 done-when to test IDs
- Updated all trace projections
