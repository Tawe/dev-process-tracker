Found it. Now I can answer all three questions.

---

## Question 2: Why "invalid pid: 0" in TUI but CLI works

Two completely different code paths:

**CLI `devpt stop pmdt-fe`** → `lifecycleManager().Stop(svc)` → reconciles live state → finds the running PID → stops it → works.

**TUI Ctrl+E on managed list** → `prepareManagedStopConfirm()` → reads `svc.LastPID` from registry → if nil/0, opens confirm modal with **no PID** → `executeConfirm` calls `m.app.StopProcess(c.pid, 5*time.Second)` → `processManager.Stop(0, ...)` → `if pid <= 0 { return "invalid pid: 0" }`.

The TUI stop path for managed services **bypasses lifecycle entirely**. It reads the registry PID directly and passes it raw to the process manager. No reconciliation. If the registry has no `LastPID` (cleared after a previous stop, or never set for a discovered process), the TUI sends PID 0 to the process manager.

**This is an arch issue:** the TUI has its own stop/restart logic that doesn't use the lifecycle layer. Two behaviors, one product.

---

## Question 1: What hotfix now

Three hotfixes, ranked by urgency:

| # | Fix | Why now |
|---|-----|---------|
| H1 | TUI managed stop must use lifecycle, not raw PID | "invalid pid: 0" is a broken user action |
| H2 | `isAmbiguous` — skip processes on undeclared ports | Blocks legitimate starts on shared-CWD projects |
| H3 | Restart port preflight (already done) | Prevents silent success on wrong port |

H2 and H3 are done. H1 is the remaining hotfix.

## Question 3: The plan

Here's the full plan with must-haves vs postponable.

---

### Phase 0 — Hotfixes ✅ DONE

All three hotfixes shipped:
- H1: TUI managed stop uses lifecycle (commit 4db59e0)
- H2: isAmbiguous skips processes on undeclared ports (commit 60bc8a8)
- H3: Restart port preflight (commit 60bc8a8)

---

### Phase 1 — Identity foundation ✅ DONE

**1.1 Record resolved command at spawn time** ✅ (commit 60bc8a8)

**1.2 Promote port to primary identity signal for services with declared ports** ✅ (commit 60bc8a8)

**1.3 Redefine ambiguity as "conflict"** ✅ (commit 60bc8a8)

**1.4 Update PROCESS_MANAGEMENT.md** ✅ (commit 96cc2f7)

---

### Phase 2 — Related processes (must-have, the grouping model)

**2.1 Discover related processes**

During scan/reconcile, after identifying primaries, find processes that share CWD+resolved-command with a managed service but are on a different port. Tag them as "related" in the service record.

**Business value:** Orphans and duplicate instances become visible. No more invisible processes running under the radar.

**2.2 Display related processes in TUI**

Collapsed row with `+N` badge. Expandable with Enter or `e`. Related rows show port, PID, and "(related)" status.

**Business value:** Operators see the full picture without clutter. Clean services look clean; noisy services show the noise.

**Investigation needed:** How expand/collapse works in the Bubble Tea grid. The managed list is currently a flat table. Expansion means inserting sub-rows that belong to a parent row. This affects selection logic, scrolling, and key handling. Need to prototype before committing to an approach.

**2.3 Display related processes in CLI**

`devpt ls` shows `+N` in a new column. `devpt ls --related` shows expanded. `devpt status <name>` always shows the full group.

**Business value:** CLI parity with TUI.

**2.4 Prune command**

`devpt prune` — stops related processes for all services. `devpt prune <name>` — stops related for one service. Primary is untouched.

**Business value:** One-command orphan cleanup. The operator doesn't need to find and kill PIDs manually.

---

### Phase 3 — Polish (postponable)

| Item | Business value | Why postponable |
|------|---------------|-----------------|
| `devpt stop --related` | Batch cleanup with primary | `prune` covers the common case |
| Related count in batch summary | Visibility during batch ops | Phase 2 display covers single-service view |
| Configurable identity signals per service | Flexibility for edge cases | Default strategy covers 99% of cases |
| `resolved_command` migration for existing entries | Backward compatibility | First start of each service teaches it automatically |
| TUI color-coding for related vs primary | Visual clarity | Text labels work initially |

---

### Dependency graph

```
H1 (TUI stop fix)         ← standalone, ship now
H2 (ambiguity fix)         ← already done
H3 (restart preflight)     ← already done

1.1 (resolved command)     ← enables 2.1
1.2 (port-primary identity) ← standalone
1.3 (conflict replaces ambiguity) ← depends on 1.2
1.4 (doc update)           ← depends on 1.2, 1.3

2.1 (discover related)     ← depends on 1.1, 1.2
2.2 (TUI expand/collapse)  ← depends on 2.1
2.3 (CLI related display)  ← depends on 2.1
2.4 (prune command)        ← depends on 2.1
```
