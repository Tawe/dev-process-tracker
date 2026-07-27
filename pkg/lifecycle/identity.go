package lifecycle

import (
	"path/filepath"
	"strings"

	"github.com/devports/devpt/pkg/models"
)

// IdentityResult holds the result of an identity verification.
type IdentityResult struct {
	Verified bool
	Process  *models.ProcessRecord
	Status   string // "verified", "unknown", "not_found"
}

// ProjectResolver resolves a project root from a CWD path.
// Returns the project root, or empty string if unresolvable.
type ProjectResolver func(cwd string) string

// VerifyIdentity checks whether a live process matches a managed service
// using the ordered evidence chain:
//  1. PID + start time (definitive, if stored)
//  1b. Stored LastPID + path corroboration
//  2. Declared port (strong, if unique among managed services)
//  3. CWD + resolved command (grouping key)
//  4. Exact CWD match (unique CWD)
//  5. Exact project root match (unique root)
func VerifyIdentity(
	svc *models.ManagedService,
	processes []*models.ProcessRecord,
	allServices []*models.ManagedService,
) IdentityResult {
	return VerifyIdentityWithResolver(svc, processes, allServices, nil)
}

// VerifyIdentityWithResolver is like VerifyIdentity but accepts an optional
// project root resolver for more accurate project root matching.
//
// Evidence chain (ordered by strength):
//  1. PID + start time (definitive, if stored)
//  1b. Stored LastPID + path corroboration (strong, even without start time)
//  2. Declared port (strong, if unique among managed services)
//  3. CWD + resolved command (grouping key for related processes)
//  4. Exact CWD match (unique CWD, fallback for portless services)
//  5. Exact project root match (unique root, fallback)
func VerifyIdentityWithResolver(
	svc *models.ManagedService,
	processes []*models.ProcessRecord,
	allServices []*models.ManagedService,
	resolver ProjectResolver,
) IdentityResult {
	if svc == nil {
		return IdentityResult{Status: "not_found"}
	}

	// Precompute per-service identity data across all services
	type svcIdentity struct {
		cwd   string
		root  string
		ports map[int]bool
	}

	resolve := resolver
	if resolve == nil {
		resolve = func(cwd string) string { return cwd }
	}

	identities := make(map[*models.ManagedService]svcIdentity, len(allServices))
	cwdCount := make(map[string]int)
	rootCount := make(map[string]int)
	portCount := make(map[int]int) // how many managed services declare this port
	portOwner := make(map[int]*models.ManagedService) // port -> owning service (if unique)

	for _, s := range allServices {
		if s == nil {
			continue
		}
		svcCWD := normalizePath(s.CWD)
		svcRoot := normalizePath(resolve(s.CWD))
		ports := make(map[int]bool, len(s.Ports))
		for _, p := range s.Ports {
			ports[p] = true
		}
		identities[s] = svcIdentity{
			cwd:   svcCWD,
			root:  svcRoot,
			ports: ports,
		}
		if identities[s].cwd != "" {
			cwdCount[identities[s].cwd]++
		}
		if identities[s].root != "" {
			rootCount[identities[s].root]++
		}
		for p := range ports {
			portCount[p]++
			portOwner[p] = s
		}
	}

	myID := identities[svc]

	// Evidence 1: PID + start time (definitive)
	if svc.LastPID != nil && *svc.LastPID > 0 && svc.LastProcessStartTime != nil {
		for _, proc := range processes {
			if proc == nil || proc.PID != *svc.LastPID {
				continue
			}
			if proc.StartTime == nil {
				return IdentityResult{
					Verified: false,
					Process:  proc,
					Status:   "unknown",
				}
			}
			if proc.StartTime.Equal(*svc.LastProcessStartTime) {
				return IdentityResult{
					Verified: true,
					Process:  proc,
					Status:   "verified",
				}
			}
			return IdentityResult{
				Verified: false,
				Process:  proc,
				Status:   "unknown",
			}
		}
	}

	// Evidence 1b: Stored LastPID + path corroboration (strong, even without start time)
	// This takes precedence over port matching because a previously confirmed
	// process identity is more reliable than a port match (which could be a conflict).
	if svc.LastPID != nil && *svc.LastPID > 0 {
		for _, proc := range processes {
			if proc == nil || proc.PID != *svc.LastPID {
				continue
			}
			procCWD := normalizePath(proc.CWD)
			if myID.cwd != "" && procCWD != "" && myID.cwd == procCWD {
				return IdentityResult{
					Verified: true,
					Process:  proc,
					Status:   "verified",
				}
			}
			procRoot := normalizePath(proc.ProjectRoot)
			if myID.root != "" && procRoot != "" && myID.root == procRoot {
				return IdentityResult{
					Verified: true,
					Process:  proc,
					Status:   "verified",
				}
			}
			break // PID matched but no path evidence
		}
	}

	// Evidence 2: Declared port owned by exactly one plausible managed service.
	// Port is the primary runtime signal for services that declare one.
	// For shared-CWD services, port uniquely distinguishes them.
	for _, port := range svc.Ports {
		if port <= 0 {
			continue
		}
		if portCount[port] != 1 {
			continue // Not uniquely owned
		}
		for _, proc := range processes {
			if proc == nil || proc.Port != port {
				continue
			}
			// If both service and process have CWD info that conflicts, skip
			procCWD := normalizePath(proc.CWD)
			if myID.cwd != "" && procCWD != "" && myID.cwd != procCWD {
				continue
			}
			// If both have root info that conflicts, skip
			procRoot := normalizePath(proc.ProjectRoot)
			if myID.root != "" && procRoot != "" && myID.root != procRoot {
				continue
			}
			return IdentityResult{
				Verified: true,
				Process:  proc,
				Status:   "verified",
			}
		}
	}

	// Evidence 3: CWD + resolved command match.
	// When a service has a learned resolved command (captured after first start),
	// matching both CWD and the resolved command is strong evidence.
	if myID.cwd != "" && svc.ResolvedCommand != "" {
		for _, proc := range processes {
			if proc == nil {
				continue
			}
			procCWD := normalizePath(proc.CWD)
			if procCWD == "" || procCWD != myID.cwd {
				continue
			}
			if proc.Command != "" && commandMatches(svc.ResolvedCommand, proc.Command) {
				// CWD + command match — but only if the process isn't on a
				// declared port of another service (that would be a conflict).
				if proc.Port > 0 {
					if owner, ok := portOwner[proc.Port]; ok && owner != svc {
						continue // Belongs to another service
					}
				}
				return IdentityResult{
					Verified: true,
					Process:  proc,
					Status:   "verified",
				}
			}
		}
	}

	// Evidence 4: Exact CWD match (must be unique among managed services)
	// Fallback for portless services with unique CWDs.
	if myID.cwd != "" && cwdCount[myID.cwd] == 1 {
		for _, proc := range processes {
			if proc == nil {
				continue
			}
			procCWD := normalizePath(proc.CWD)
			if procCWD != "" && procCWD == myID.cwd {
				return IdentityResult{
					Verified: true,
					Process:  proc,
					Status:   "verified",
				}
			}
		}
	}

	// Evidence 5: Exact project root match (must be unique among managed services)
	if myID.root != "" && rootCount[myID.root] == 1 {
		for _, proc := range processes {
			if proc == nil {
				continue
			}
			procRoot := normalizePath(proc.ProjectRoot)
			if procRoot != "" && procRoot == myID.root {
				return IdentityResult{
					Verified: true,
					Process:  proc,
					Status:   "verified",
				}
			}
		}
	}

	return IdentityResult{
		Verified: false,
		Status:   "not_found",
	}
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimRight(p, "/")
	return p
}

// resolveProjectRoot returns the CWD itself as a simplistic project root.
// In production, this would use scanner.ProjectResolver, but we avoid that
// dependency here to keep the function pure and testable.
func resolveProjectRoot(cwd string) string {
	return cwd
}

// commandMatches checks whether a stored resolved command matches a live process command.
// Since the resolved command is captured from `ps` at spawn time, the strings should
// match exactly in most cases. The function handles minor path differences by
// comparing the entry point and remaining arguments.
func commandMatches(resolvedCmd, procCmd string) bool {
	r := strings.TrimSpace(resolvedCmd)
	p := strings.TrimSpace(procCmd)
	if r == p {
		return true
	}
	// When both start with the same interpreter (node, bun, python),
	// compare entry-point basename and remaining arguments.
	// e.g. "node /long/path/vite" matches "node /other/path/vite"
	// but "node /long/path/vite" does NOT match "node /path/vite preview --port 3070"
	resolvedParts := strings.Fields(r)
	procParts := strings.Fields(p)
	if len(resolvedParts) < 2 || len(procParts) < 2 {
		return false
	}
	if resolvedParts[0] != procParts[0] {
		return false
	}
	// Entry-point basename must match (handles path differences)
	if filepath.Base(resolvedParts[1]) != filepath.Base(procParts[1]) {
		return false
	}
	// Remaining arguments must match exactly (prevents vite matching vite preview)
	if len(resolvedParts) != len(procParts) {
		return false
	}
	for i := 2; i < len(resolvedParts); i++ {
		if resolvedParts[i] != procParts[i] {
			return false
		}
	}
	return true
}
