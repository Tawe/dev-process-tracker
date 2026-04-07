package tui

import (
	"sort"
	"strings"

	"github.com/devports/devpt/pkg/models"
)

type sortMode int

const (
	sortRecent sortMode = iota
	sortName
	sortProject
	sortPort
	sortHealth
	sortModeCount
)

// sortModeLabel returns a human-readable label for the sort mode.
func sortModeLabel(s sortMode) string {
	switch s {
	case sortName:
		return "name"
	case sortProject:
		return "project"
	case sortPort:
		return "port"
	case sortHealth:
		return "health"
	default:
		return "recent"
	}
}

// sortServers sorts the given servers slice according to the current sort mode.
func (m topModel) sortServers(servers []*models.ServerInfo) {
	switch m.sortBy {
	case sortName:
		sort.Slice(servers, func(i, j int) bool {
			cmp := strings.Compare(strings.ToLower(m.serviceNameFor(servers[i])), strings.ToLower(m.serviceNameFor(servers[j])))
			if m.sortReverse {
				return cmp > 0
			}
			return cmp < 0
		})
	case sortProject:
		sort.Slice(servers, func(i, j int) bool {
			cmp := strings.Compare(strings.ToLower(projectOf(servers[i])), strings.ToLower(projectOf(servers[j])))
			if m.sortReverse {
				return cmp > 0
			}
			return cmp < 0
		})
	case sortPort:
		sort.Slice(servers, func(i, j int) bool {
			if m.sortReverse {
				return portOf(servers[i]) > portOf(servers[j])
			}
			return portOf(servers[i]) < portOf(servers[j])
		})
	case sortHealth:
		sort.Slice(servers, func(i, j int) bool {
			cmp := strings.Compare(m.health[portOf(servers[i])], m.health[portOf(servers[j])])
			if m.sortReverse {
				return cmp > 0
			}
			return cmp < 0
		})
	default:
		sort.Slice(servers, func(i, j int) bool { return pidOf(servers[i]) > pidOf(servers[j]) })
	}
}

// columnAtX returns the sortMode for the column at the given X coordinate.
// Returns -1 if the X is not within a clickable column header.
func (m *topModel) columnAtX(x int) sortMode {
	nameW, portW, pidW, projectW, healthW := 14, 6, 7, 14, 7
	sep := 2
	used := nameW + sep + portW + sep + pidW + sep + projectW + sep + healthW + sep
	cmdW := m.width - used
	if cmdW < 12 {
		cmdW = 12
	}

	// Column positions (start, end)
	nameEnd := nameW
	portStart := nameW + sep
	portEnd := portStart + portW
	pidStart := portEnd + sep
	pidEnd := pidStart + pidW
	projectStart := pidEnd + sep
	projectEnd := projectStart + projectW
	cmdStart := projectEnd + sep
	cmdEnd := cmdStart + cmdW
	healthStart := cmdEnd + sep
	healthEnd := healthStart + healthW

	switch {
	case x >= 0 && x < nameEnd:
		return sortName
	case x >= portStart && x < portEnd:
		return sortPort
	case x >= pidStart && x < pidEnd:
		return sortRecent // PID sorts by recent (default)
	case x >= projectStart && x < projectEnd:
		return sortProject
	case x >= cmdStart && x < cmdEnd:
		return sortRecent // Command column - no specific sort, use recent
	case x >= healthStart && x < healthEnd:
		return sortHealth
	default:
		return -1
	}
}

// toggleSortDirection flips the sort direction between ascending and descending.
// No effect when in "Recent" mode (natural order only).
func (m *topModel) toggleSortDirection() {
	if m.sortBy == sortRecent {
		return
	}
	m.sortReverse = !m.sortReverse
}

// cycleSort implements 3-state sort cycling: ascending (yellow) → reverse (orange) → reset to recent
func (m *topModel) cycleSort(col sortMode) {
	// If clicking the same column that's currently sorted
	if m.sortBy == col && m.sortBy != sortRecent {
		if !m.sortReverse {
			// State 1 → State 2: same column, now reverse
			m.sortReverse = true
		} else {
			// State 2 → State 3: reset to recent
			m.sortBy = sortRecent
			m.sortReverse = false
		}
	} else {
		// Different column or clicking recent: go to State 1 (ascending)
		m.sortBy = col
		m.sortReverse = false
	}
}
