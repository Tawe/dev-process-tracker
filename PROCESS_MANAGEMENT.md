# Process Management Behavioral Contract

Defines the correct workflow and operator-facing behavior for managed service lifecycle operations: `start`, `stop`, `restart`, and batch execution.

This is a process contract, not an implementation note. It defines what must be true before, during, and after each lifecycle action.

This document standardizes the workflow algorithm and operator experience. It is intentionally stricter than the current implementation. Where the implementation is simpler, this document defines the target behavior to converge toward.

---

## 1. Operating Model

### 1.1 Sources of Truth

The system has three different kinds of state:

- **Desired state**: the managed service definition in the registry
- **Observed state**: what the system can prove right now by scanning processes and ports
- **Operation state**: an in-progress lifecycle action owned by exactly one operator flow

The key rule:

> Observed state is authoritative for whether a service is running.  
> Registry state stores configuration and last confirmed ownership metadata.

Because this is a daemonless workflow, the registry cannot be treated as continuously current. A process can die immediately after a successful write. Every command must reconcile live state before acting.

### 1.2 Durable State vs Command Phase

The contract separates persistent service status from command-local execution phase.

Persistent service status is what operators may rely on between commands:

- **running**
- **stopped**
- **crashed**
- **unknown**

Command phase is transient and exists only while a lifecycle command owns the service:

- **starting**
- **stopping**
- **restarting**

Unless the system introduces persisted operation records, command phase is not durable state and must not be shown later as if it were.

### 1.3 Service Identity

A service must never be identified by PID alone.

Identity must be verified using an ordered evidence chain:

1. **PID + start time** — definitive after confirmed start
1b. **Stored LastPID + path corroboration** — strong, even without start time
2. **Declared port** — primary runtime signal for services that declare one
3. **CWD + resolved command** — grouping key for discovering related processes
4. **Exact CWD match** — fallback for portless services with unique CWDs
5. **Exact project root match** — fallback for portless services with unique roots

Additionally, the system captures a **resolved command** at spawn time — the actual
command visible via `ps` after the OS interprets the declared command. For example,
`bunx vite` resolves to `node .../node_modules/.bin/vite`. This learned mapping enables
reliable runtime identity matching without fuzzy heuristics.

If PID reuse is possible and identity cannot be proven, the service must be treated as **unknown**, not **running**.

#### Shared-CWD Services

Multiple services sharing the same CWD (e.g., backend, frontend, preview from the same project) is a first-class scenario. In this case, **port** is the primary distinguishing signal. Services without declared ports in a shared-CWD must have unique commands or risk being classified as `unknown`.

#### Identity Groups (Target Architecture)

The system is evolving toward a **service group** model where processes sharing CWD + resolved command are grouped together:

| Member | Match | Meaning |
|--------|-------|---------|
| **Primary** | Declared port match | The tracked instance |
| **Related** | Same CWD+command, different port | Sibling process (orphan, duplicate) |
| **Conflict** | Declared port held by process outside group | Someone else using our port |
| **Untracked** | Doesn't match any group | Not our concern |

### 1.4 Operation Ownership

Only one lifecycle operation may own a service at a time.

Before `start`, `stop`, or `restart`, the system must acquire a per-service operation lock.

If the lock cannot be acquired:

- Do not continue optimistically
- Report that another operation is already in progress
- Exit with a blocked result

### 1.5 Registry Write Rule

The registry may store:

- service definition
- last confirmed PID
- last confirmed process start time
- last confirmed readiness timestamp
- last log path or log session metadata

The registry must not be used as the sole proof that a service is alive.

---

## 2. Status, Phase, and Outcomes

```mermaid
stateDiagram-v2
    [*] --> stopped

    stopped --> starting : start
    starting --> running : ready
    starting --> stopped : start failed

    running --> stopping : stop
    stopping --> stopped : stopped

    running --> restarting : restart
    restarting --> running : ready
    restarting --> stopped : restart failed

    running --> crashed : observed dead
    crashed --> stopped : reconcile
    crashed --> starting : restart
```

### 2.1 Persistent Service Status

- **running**: a live process identity has been verified and readiness has passed when required
- **stopped**: no verified running instance exists
- **crashed**: the last confirmed instance is gone and the tool has evidence of an unexpected exit or stale last-run metadata
- **unknown**: a process may exist, but ownership cannot be proven safely

### 2.2 Command Phase

- **starting**: a start operation owns the service and readiness is being verified
- **stopping**: shutdown is in progress and the current instance may still own resources
- **restarting**: one verified instance is being replaced by another

These are command-local phases, not durable statuses, unless a future operation journal explicitly persists them.

### 2.3 Command Outcomes

Every lifecycle command must end in one of these outcomes:

- **success**: requested state change completed
- **noop**: requested end state already existed
- **blocked**: action was prevented by a lock, conflict, or unsafe ambiguity that may be resolved externally
- **failed**: action was attempted but could not complete
- **invalid**: the request or service definition is invalid
- **not_found**: the requested service identifier matched nothing

This standard replaces vague failure-only reporting with explicit operator-facing outcomes.

### 2.4 Outcome Rules

- use **blocked** for lock contention, identity ambiguity, or external resource conflicts
- use **invalid** for malformed commands, missing working directories, or impossible service definitions
- use **not_found** when resolution fails before any lifecycle work begins
- do not collapse all non-success results into **failed**

---

## 3. Universal Workflow

Every lifecycle operation must follow the same high-level algorithm.

```mermaid
flowchart TD
    A[Resolve service] --> B{Service exists}
    B -- No --> X1[Outcome: not_found]
    B -- Yes --> C[Validate request and service contract]
    C --> D{Valid}
    D -- No --> X2[Outcome: invalid]
    D -- Yes --> E[Acquire service lock]
    E --> F{Lock acquired}
    F -- No --> X3[Outcome: blocked]
    F -- Yes --> G[Reconcile live state]
    G --> H[Run command-specific flow]
    H --> I[Persist confirmed metadata]
    I --> J[Release lock]
```

### 3.1 Reconcile Live State

Before any mutation:

- scan current listeners and processes
- match live processes against managed services by identity, not just PID
- clear stale metadata that can no longer be verified
- classify the service as `running`, `stopped`, `crashed`, or `unknown`

If the service is `unknown`, the system must not take destructive action until identity is clarified.

### 3.2 Lock Protocol

Per-service locking must follow these rules:

- lock scope is one managed service identifier
- lock owner records command type and acquisition timestamp
- lock acquisition is exclusive
- stale locks must be recoverable by timeout or explicit verification that the owner is gone
- batch operations acquire and release one service lock at a time unless a higher-level planner is explicitly introduced

If a lock cannot be acquired safely, return `blocked` and do not continue optimistically.

### 3.3 Persist Only Confirmed Facts

Write registry metadata only after a fact has been confirmed:

- do not record a PID before the child is proven alive
- do not mark a service running before readiness passes
- do not clear stop metadata until the process is confirmed gone

### 3.4 Identity Verification Algorithm

Identity verification must use ordered evidence, not ad hoc matching.

Evidence chain (ordered by strength):

1. **PID + start time**: stored PID matches a live process with the same OS-reported start time
1b. **Stored LastPID + path corroboration**: stored PID matches a live process whose CWD or project root matches
2. **Declared port**: a uniquely-declared port matches a live process, with CWD/root as corroboration (not requirement)
3. **CWD + resolved command**: both working directory and the OS-resolved command match
4. **Exact CWD match**: working directory matches and is unique among all managed services
5. **Exact project root match**: project root matches and is unique among all managed services

Verification rules:

- at least one path-based or uniquely-owned port-based signal must exist
- PID alone is never sufficient
- command string alone is never sufficient
- stored LastPID with path corroboration takes precedence over port matching — a previously confirmed identity is more reliable than a port match which could be a conflict
- if multiple managed services remain plausible after matching, classify as `unknown`
- if evidence conflicts, prefer safety over convenience and classify as `unknown`
- a process on a port that no managed service declares is irrelevant and must not poison identity checks

---

## 4. Start

### 4.1 Start Flow

```mermaid
flowchart TD
    A[Resolve and lock] --> B[Reconcile live state]
    B --> C{Already running}
    C -- Yes --> Z1[No-op: already running]
    C -- No --> D[Run preflight]
    D --> E{Preflight passed}
    E -- No --> Z2[Outcome: invalid or blocked]
    E -- Yes --> F[Spawn process]
    F --> G[Verify process identity]
    G --> H[Wait for readiness]
    H --> I{Ready}
    I -- Yes --> J[Record confirmed run]
    J --> Z3[Success: started]
    I -- No --> K[Collect diagnostics]
    K --> L[Cleanup failed start]
    L --> Z4[Failed: start did not complete]
```

### 4.2 Start Rules

- `start` is end-state oriented: its job is to ensure the service is running
- if a verified instance is already running, return `noop`
- if a stale registry entry exists, clear it during reconciliation before any fork
- if identity is ambiguous, return `blocked`
- never spawn a second instance just because the registry is stale

### 4.3 Preflight Requirements

Before any fork:

- working directory exists and is a directory
- command parses into an executable and arguments
- executable can be resolved
- all declared ports are free, or are already owned by the same verified instance
- required files or env assumptions are present when the service contract requires them

Preflight failures caused by invalid service definition return `invalid`.

Preflight failures caused by external contention, such as port conflicts, return `blocked`.

### 4.3.1 Resolved Command Capture

After a successful start, the system reads the OS-reported command line from the spawned process (`ps -p <PID> -o command=`) and stores it as `resolved_command` in the registry. This learned mapping (e.g., `bunx vite` → `node .../vite`) enables reliable identity matching during future reconciles without fuzzy heuristics.

The declared command is used for spawning. The resolved command is used for identity only.

### 4.4 Readiness Policy

Readiness is a service policy, not an ad hoc runtime guess.

Allowed readiness modes:

- **process-only**: child remains alive for the startup window
- **port-bound**: declared port is bound by the verified child
- **http-health**: HTTP readiness endpoint returns success
- **log-signal**: a declared log pattern appears
- **multi-check**: more than one condition must pass

If the service model supports explicit readiness configuration, the service definition must declare which mode applies.

If no explicit readiness policy exists yet, the fallback policy is:

- `port-bound` for services with declared ports
- `process-only` for services without declared ports

This fallback is transitional. A future richer service contract may replace it.

### 4.5 Start Failure Handling

If start fails:

- collect a short diagnostic summary
- include log tail when available
- kill the child if it is still alive but not ready
- do not write unconfirmed PID data
- return `failed`

### 4.6 Required Message Format

Start messages must use decisive operator language and must state the resolved outcome.

- `Success: started "api" on port 3000 (PID 4821).`
- `No-op: "api" is already running on port 3000 (PID 4821).`
- `Blocked: port 3000 is in use by PID 4821 (python). Stop it or change the service port.`
- `Invalid: "api" has a missing working directory: /path/to/project.`
- `Failed: "api" did not become ready within 20s. Check logs with devpt logs api.`

---

## 5. Stop

### 5.1 Stop Flow

```mermaid
flowchart TD
    A[Resolve and lock] --> B[Reconcile live state]
    B --> C{Already stopped}
    C -- Yes --> Z1[No-op: already stopped]
    C -- No --> D{Identity verified}
    D -- No --> Z2[Blocked: unsafe to kill]
    D -- Yes --> E[Send SIGTERM]
    E --> F{Exited in time}
    F -- Yes --> G[Confirm resource release]
    F -- No --> H[Send SIGKILL]
    H --> I{Exited}
    I -- No --> Z3[Failed: process still alive]
    I -- Yes --> G[Confirm resource release]
    G --> J[Clear confirmed run metadata]
    J --> Z4[Success: stopped]
```

### 5.2 Stop Rules

- `stop` is idempotent: if the service is already stopped, return `noop`
- if the registry contains stale metadata and no verified live instance exists, clear the stale data and return `noop`
- never kill a process when service identity is ambiguous
- terminate gracefully first, then escalate
- confirm that the process is gone before clearing ownership metadata
- if service status is `unknown`, refuse destructive action and return `blocked`

### 5.3 Stop Failure Handling

If forced kill fails:

- report the PID and why termination failed
- tell the operator whether elevated permissions may be required
- leave the service in `blocked` or `failed`, not falsely `stopped`

### 5.4 Required Message Format

Stop messages must state whether the final state is already satisfied, blocked, or failed.

- `Success: stopped "worker" (PID 3105).`
- `No-op: "worker" is already stopped.`
- `No-op: stale PID 3105 was cleared for "worker".`
- `Blocked: PID 3105 cannot be proven to belong to "worker"; refusing to kill.`
- `Failed: PID 3105 did not exit after SIGTERM and SIGKILL. Sudo may be required.`

---

## 6. Restart

### 6.1 Restart Flow

```mermaid
flowchart TD
    A[Resolve and lock] --> B[Reconcile live state]
    B --> C{Running now}
    C -- Yes --> D[Stop verified instance]
    C -- No --> E[Clear stale metadata]
    D --> F{Stopped cleanly}
    F -- No --> Z1[Blocked: old instance remains]
    F -- Yes --> G[Wait for resources to clear]
    E --> G[Wait for resources to clear]
    G --> H{Preflight passed}
    H -- No --> Z2[Blocked: cannot restart safely]
    H -- Yes --> I[Spawn new instance]
    I --> J[Verify identity and readiness]
    J --> K{Ready}
    K -- Yes --> L[Record confirmed run]
    L --> Z3[Success: restarted]
    K -- No --> Z4[Failed: old instance gone, new instance not ready]
```

### 6.2 Restart Rules

- `restart` means replace the current instance with a fresh verified instance
- the old instance must be confirmed gone before the new one is accepted
- if the old instance cannot be stopped, return `blocked`
- if the old instance is already gone, clean stale metadata and continue
- if start fails after stop succeeds, report that the service is now stopped, not running
- if the service was already stopped, the operator-facing message must say that restart resolved as a fresh start
- restart must run the same preflight checks as start — declared ports must be free before spawning
- if a declared port is held by another process after stopping the old instance, return `blocked` with the conflict details
- never silently accept a process on a fallback port as a successful restart

### 6.3 Freshness Rule

When a previous instance existed, the new confirmed run must differ by identity from the old one. A restart that simply rediscovers the same old instance is not a valid restart.

### 6.4 Required Message Format

- `Success: restarted "api" with a fresh instance (old PID 3105, new PID 4821).`
- `Success: started "worker" because no verified instance was running.`
- `Blocked: could not restart "web" because the old instance still owns port 3000.`
- `Failed: "api" was stopped, but the replacement instance did not become ready.`

---

## 7. Batch Operations

Batch commands must optimize operator clarity, not just throughput.

### 7.1 Batch Flow

```mermaid
flowchart TD
    A[Expand identifiers] --> B[Show execution plan]
    B --> C[Process services in stable order]
    C --> D[Run per-service workflow]
    D --> E[Collect outcome]
    E --> F{More services}
    F -- Yes --> C
    F -- No --> G[Print summary]
```

### 7.2 Batch Rules

- expand patterns before execution
- deduplicate matches
- process services in a stable and predictable order
- continue after per-service failures unless the command explicitly declares fail-fast behavior
- return non-zero if any service failed
- distinguish `success`, `noop`, `blocked`, `failed`, `invalid`, and `not_found` in the summary

### 7.3 Dependency-Aware UX

If services have declared dependencies, the batch planner must:

- start dependencies before dependents
- stop dependents before dependencies
- restart in dependency-aware order

If dependency data is unavailable, the batch planner must use a stable deterministic order and report that dependency ordering was unavailable.

Dependency ordering is an extension policy. If the service model does not yet carry dependency data, the batch system must not invent it.

### 7.4 Summary Format

The batch summary must report:

- total matched
- succeeded
- noop
- blocked
- failed
- invalid
- not found
- per-service reason for every non-success outcome

Example:

```text
Matched 4 services
2 succeeded, 1 noop, 1 blocked

- api: started
- worker: started
- web: already running
- redis: port 6379 is in use by PID 4821
```

---

## 8. Error Reporting

All lifecycle messages must answer three questions:

- what was attempted
- what actually happened
- what the operator must do next

Bad:

- `failed to start`
- `process error`

Good:

- `Blocked: port 9055 is in use by PID 4821 (python). Stop that process or change the service port.`
- `Failed: "api" exited during startup before binding port 9055. Recent logs are available via devpt logs api.`
- `Invalid: "worker" has an invalid command definition.`
- `Blocked: another restart is already in progress for "worker". Retry after it completes.`

---

## 9. Non-Negotiable Rules

- never trust registry PID data without live reconciliation
- never identify a service by PID alone
- never record a run before identity and readiness are confirmed
- never kill a process whose identity is ambiguous
- never report `running` unless observed state proves it
- never report `stopped` until shutdown is confirmed
- never hide stale metadata cleanup
- never let concurrent operations mutate the same service without a lock
- never present transient command phase as durable service state unless operation records exist
- never silently accept a process on the wrong port as the managed instance
- never skip port preflight checks during restart
- never let a process on an undeclared port block identity checks for services with declared ports

These rules exist to protect operator trust. Once the tool lies about lifecycle state, every downstream command becomes unreliable.
