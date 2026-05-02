package lifecycle

import (
	"github.com/devports/devpt/pkg/models"
)

// ReconciledService holds the result of reconciling a service against live state.
type ReconciledService struct {
	Status          string                // "running", "stopped", "crashed", "unknown"
	Verified        bool
	Process         *models.ProcessRecord
	HasStaleMetadata bool // true when LastPID exists but no verified process was found
}

// Reconcile scans live processes, matches against managed services by identity,
// classifies status, and clears stale metadata.
func Reconcile(
	svc *models.ManagedService,
	processes []*models.ProcessRecord,
	allServices []*models.ManagedService,
) ReconciledService {
	return ReconcileWithResolver(svc, processes, allServices, nil)
}

// ReconcileWithResolver is like Reconcile but accepts an optional project root resolver.
func ReconcileWithResolver(
	svc *models.ManagedService,
	processes []*models.ProcessRecord,
	allServices []*models.ManagedService,
	resolver ProjectResolver,
) ReconciledService {
	if svc == nil {
		return ReconciledService{Status: string(models.StatusUnknown)}
	}

	// Use identity verification to determine status
	identity := VerifyIdentityWithResolver(svc, processes, allServices, resolver)

	if identity.Verified {
		return ReconciledService{
			Status:   string(models.StatusRunning),
			Verified: true,
			Process:  identity.Process,
		}
	}

	// Check if identity is ambiguous (multiple services match)
	if isAmbiguousWithResolver(svc, processes, allServices, resolver) {
		return ReconciledService{
			Status:   string(models.StatusUnknown),
			Verified: false,
		}
	}

	// No verified process found — check for stale metadata
	if svc.LastPID != nil && *svc.LastPID > 0 {
		// Had a PID but no verified process now
		return ReconciledService{
			Status:          string(models.StatusCrashed),
			Verified:        false,
			HasStaleMetadata: true,
		}
	}

	return ReconciledService{
		Status:   string(models.StatusStopped),
		Verified: false,
	}
}

// isAmbiguous checks whether multiple managed services could plausibly
// own the same live process, making identity unresolvable.
func isAmbiguous(
	svc *models.ManagedService,
	processes []*models.ProcessRecord,
	allServices []*models.ManagedService,
) bool {
	return isAmbiguousWithResolver(svc, processes, allServices, nil)
}

func isAmbiguousWithResolver(
	svc *models.ManagedService,
	processes []*models.ProcessRecord,
	allServices []*models.ManagedService,
	resolver ProjectResolver,
) bool {
	svcCWD := normalizePath(svc.CWD)
	cwdCount := make(map[string]int)
	rootCount := make(map[string]int)
	portCount := make(map[int]int)

	// portOwner maps a uniquely-declared port to the service that owns it.
	portOwner := make(map[int]*models.ManagedService)

	resolve := resolver
	if resolve == nil {
		resolve = func(cwd string) string { return cwd }
	}

	for _, s := range allServices {
		if s == nil {
			continue
		}
		c := normalizePath(s.CWD)
		if c != "" {
			cwdCount[c]++
		}
		r := normalizePath(resolve(s.CWD))
		if r != "" {
			rootCount[r]++
		}
		for _, p := range s.Ports {
			portCount[p]++
			portOwner[p] = s
		}
	}

	// Check if any process matches this service in an ambiguous way
	for _, proc := range processes {
		if proc == nil {
			continue
		}
		procCWD := normalizePath(proc.CWD)
		procRoot := normalizePath(proc.ProjectRoot)

		// If this process is uniquely claimed by another service via port,
		// it cannot create ambiguity for the current service.
		if proc.Port > 0 && portCount[proc.Port] == 1 {
			if owner, ok := portOwner[proc.Port]; ok && owner != svc {
				continue
			}
		}

		// CWD match but not unique
		if svcCWD != "" && procCWD == svcCWD && cwdCount[svcCWD] > 1 {
			return true
		}
		// Root match but not unique
		svcRoot := normalizePath(resolve(svc.CWD))
		if svcRoot != "" && procRoot == svcRoot && rootCount[svcRoot] > 1 {
			return true
		}
		// Port match but not unique
		for _, port := range svc.Ports {
			if port > 0 && proc.Port == port && portCount[port] > 1 {
				return true
			}
		}
	}

	return false
}
