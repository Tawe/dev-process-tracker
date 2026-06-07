package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"

	"github.com/devports/devpt/pkg/health"
	"github.com/devports/devpt/pkg/models"
)

type processTable struct {
	runningVP          viewport.Model
	managedListVP      viewport.Model
	selectedDetailsVP  viewport.Model

	lastRunningHeight  int
	lastManagedHeight  int
	lastListWidth      int
	lastRunningContent string
	lastListContent    string
	lastDetailsContent string
}

func newProcessTable() processTable {
	return processTable{
		runningVP:          viewport.New(),
		managedListVP:      viewport.New(),
		selectedDetailsVP:  viewport.New(),
	}
}

func (t *processTable) heightFor(termHeight, aboveLines, belowLines int) int {
	h := termHeight - aboveLines - belowLines
	if h < 3 {
		h = 3
	}
	return h
}

func (t *processTable) Render(m *topModel, width int) string {
	visible := m.visibleServers()
	managed := m.managedServices()
	displayNames := m.displayNames(visible)

	topLines := m.tableTopLines(width)
	bottomLines := m.tableBottomLines(width)
	totalHeight := t.heightFor(m.height, topLines, bottomLines)
	runningContent := m.renderRunningTable(width, visible, displayNames)
	managedHeader := m.renderManagedHeader(width, managed)
	listContent := m.renderManagedList(width/2, managed)
	detailsContent := m.renderSelectedServiceDetails(width-width/2, visible, managed)
	runningLines := 1 + strings.Count(runningContent, "\n")
	listLines := 1 + strings.Count(listContent, "\n")
	detailsLines := 1 + strings.Count(detailsContent, "\n")
	managedLines := max(listLines, detailsLines)
	runningHeight, managedHeight := t.sectionHeights(totalHeight, runningLines, managedLines)

	t.lastRunningHeight = runningHeight
	t.lastManagedHeight = managedHeight
	t.lastListWidth = width / 2

	t.runningVP.SetWidth(width)
	t.runningVP.SetHeight(runningHeight)
	if t.lastRunningContent != runningContent {
		t.runningVP.SetContent(runningContent)
		t.lastRunningContent = runningContent
	}

	t.managedListVP.SetWidth(width / 2)
	t.managedListVP.SetHeight(managedHeight)
	if t.lastListContent != listContent {
		t.managedListVP.SetContent(listContent)
		t.lastListContent = listContent
	}

	t.selectedDetailsVP.SetWidth(width - width/2)
	t.selectedDetailsVP.SetHeight(managedHeight)
	if t.lastDetailsContent != detailsContent {
		t.selectedDetailsVP.SetContent(detailsContent)
		t.lastDetailsContent = detailsContent
	}

	if m.tableFollowSelection {
		t.scrollToSelection(m, visible, managed)
	}

	listView := t.managedListVP.View()
	detailsView := t.selectedDetailsVP.View()

	return t.runningVP.View() + "\n" + managedHeader + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, listView, detailsView)
}

func (m *topModel) tableTopLines(width int) int {
	// Header line + blank line before the table content.
	return 2
}

func (m *topModel) tableBottomLines(width int) int {
	lines := renderedLineCount(m.renderFooter(width))
	if sl := m.renderStatusLine(width); sl != "" {
		lines += renderedLineCount(sl)
	}
	return lines
}

func (m *topModel) hasStatusLine() bool {
	if m.cmdStatus != "" {
		return true
	}
	// With split view, details pane shows service context - no need for status line
	return false
}

func (m *topModel) renderStatusLine(width int) string {
	text := ""
	if m.cmdStatus != "" {
		text = m.cmdStatus
	}
	// With split view, the details pane shows service state - no duplication in status line
	if text == "" {
		return ""
	}
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	return s.Render(fitLine(text, width))
}

func (m *topModel) renderFooter(width int) string {
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	h := m.help
	h.SetWidth(width)
	return strings.TrimRight(s.Render(h.View(m.footerKeyMap())), "\n")
}

func (m *topModel) footerKeyMap() keyMap {
	k := m.keys
	k.Search = key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", m.footerFilterLabel()),
	)
	if m.groupHighlightNamespace != nil {
		green := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Render("group mode")
		k.GroupToggle = key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", green),
		)
	}
	return k
}

func (m *topModel) footerFilterLabel() string {
	switch {
	case m.mode == viewModeSearch:
		inputWidth := runewidth.StringWidth(m.searchInput.Value()) + 1
		if inputWidth < 1 {
			inputWidth = 1
		}
		if inputWidth > 24 {
			inputWidth = 24
		}
		m.searchInput.SetWidth(inputWidth)
		return m.searchInput.View()
	case strings.TrimSpace(m.searchQuery) != "":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(m.searchQuery)
	default:
		return "filter"
	}
}

func (t *processTable) sectionHeights(totalHeight, runningLines, managedLines int) (int, int) {
	if totalHeight < 3 {
		return 1, 1
	}

	separator := 1
	minManaged := 3
	maxRunning := totalHeight - separator - minManaged
	if maxRunning < 1 {
		maxRunning = 1
	}

	runningHeight := runningLines
	if runningHeight > maxRunning {
		runningHeight = maxRunning
	}
	if runningHeight < 1 {
		runningHeight = 1
	}

	managedHeight := totalHeight - separator - runningHeight
	if managedHeight < 1 {
		managedHeight = 1
	}
	if managedLines > 0 && managedHeight > managedLines {
		managedHeight = managedLines
	}

	return runningHeight, managedHeight
}

func (t *processTable) scrollToSelection(m *topModel, visible []*models.ServerInfo, managed []*models.ManagedService) {
	if m.focus == focusRunning && m.selected >= 0 && m.selected < len(visible) {
		selectedLine := 2 + m.selected
		t.scrollViewportToLine(&t.runningVP, selectedLine)
	} else if m.focus == focusManaged && m.managedSel >= 0 && m.managedSel < len(managed) {
		selectedLine := m.managedSel
		t.scrollViewportToLine(&t.managedListVP, selectedLine)
	}
}

func (t *processTable) scrollViewportToLine(vp *viewport.Model, selectedLine int) {
	totalLines := vp.TotalLineCount()
	visibleLines := vp.VisibleLineCount()
	currentOffset := vp.YOffset()

	if selectedLine < currentOffset || selectedLine >= currentOffset+visibleLines {
		desired := selectedLine - visibleLines/3
		if desired < 0 {
			desired = 0
		}
		if desired > totalLines-visibleLines {
			desired = totalLines - visibleLines
		}
		if desired < 0 {
			desired = 0
		}
		vp.SetYOffset(desired)
	}
}

func (m *topModel) renderRunningTable(width int, visible []*models.ServerInfo, displayNames []string) string {
	headerStyle := lipgloss.NewStyle()
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)  // yellow for ascending
	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true) // orange for reverse

	nameW, portW, pidW, projectW, healthW := 14, 6, 7, 14, 7
	sep := 2
	used := nameW + sep + portW + sep + pidW + sep + projectW + sep + healthW + sep
	cmdW := width - used
	if cmdW < 12 {
		cmdW = 12
	}

	// Compute styles first based on sort state
	nameStyle := headerStyle
	portStyle := headerStyle
	projectStyle := headerStyle
	healthStyle := headerStyle

	switch m.sortBy {
	case sortName:
		if m.sortReverse {
			nameStyle = orangeStyle
		} else {
			nameStyle = yellowStyle
		}
	case sortPort:
		if m.sortReverse {
			portStyle = orangeStyle
		} else {
			portStyle = yellowStyle
		}
	case sortProject:
		if m.sortReverse {
			projectStyle = orangeStyle
		} else {
			projectStyle = yellowStyle
		}
	case sortHealth:
		if m.sortReverse {
			healthStyle = orangeStyle
		} else {
			healthStyle = yellowStyle
		}
	}

	nameHeader := nameStyle.Render(fixedCell(fmt.Sprintf("Name (%d)", len(visible)), nameW))
	portHeader := portStyle.Render(fixedCell("Port", portW))
	pidHeader := headerStyle.Render(fixedCell("PID", pidW))
	projectHeader := projectStyle.Render(fixedCell("Project", projectW))
	commandHeader := headerStyle.Render(fixedCell("Command", cmdW))
	healthHeader := healthStyle.Render(fixedCell("Health", healthW))

	header := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s",
		nameHeader, pad(sep),
		portHeader, pad(sep),
		pidHeader, pad(sep),
		projectHeader, pad(sep),
		commandHeader, pad(sep),
		healthHeader,
	)
	divider := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s",
		fixedCell(strings.Repeat("─", nameW), nameW), pad(sep),
		fixedCell(strings.Repeat("─", portW), portW), pad(sep),
		fixedCell(strings.Repeat("─", pidW), pidW), pad(sep),
		fixedCell(strings.Repeat("─", projectW), projectW), pad(sep),
		fixedCell(strings.Repeat("─", cmdW), cmdW), pad(sep),
		fixedCell(strings.Repeat("─", healthW), healthW),
	)

	if len(visible) == 0 {
		if m.searchQuery != "" {
			return fitLine("(no matching servers for filter)", width)
		}
		return fitLine("(no matching servers)", width)
	}

	var lines []string
	lines = append(lines, fitAnsiLine(header, width))
	lines = append(lines, fitLine(divider, width))

	rowIndices := make([]int, len(visible))
	for i, srv := range visible {
		rowIndices[i] = len(lines)

		project := projectOf(srv)
		port := "-"
		pid := 0
		cmd := "-"
		icon := "…"
		if srv.ProcessRecord != nil {
			pid = srv.ProcessRecord.PID
			cmd = srv.ProcessRecord.Command
			if srv.ProcessRecord.Port > 0 {
				port = fmt.Sprintf("%d", srv.ProcessRecord.Port)
				if cached := m.health[srv.ProcessRecord.Port]; cached != "" {
					icon = cached
				}
			}
		}

		truncatedCmd := cmd
		if runewidth.StringWidth(cmd) > cmdW {
			truncatedCmd = runewidth.Truncate(cmd, cmdW, "...")
		}

		line := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s",
			fixedCell(displayNames[i], nameW), pad(sep),
			portCell(port, portW), pad(sep),
			fixedCell(fmt.Sprintf("%d", pid), pidW), pad(sep),
			fixedCell(project, projectW), pad(sep),
			fixedCell(truncatedCmd, cmdW), pad(sep),
			fixedCell(icon, healthW),
		)
		// Use fitAnsiLine because portCell may contain OSC8 hyperlinks
		// (runewidth.StringWidth in fitLine doesn't understand escape sequences)
		lines = append(lines, fitAnsiLine(line, width))
	}

	// Apply row styles using shared color logic: group members, selection, confirm target.
	confirmActive := m.activeModalKind() == modalConfirm
	confirmTarget := m.confirmTargetName()
	confirmPID := 0
	if m.confirm != nil && m.confirm.kind == confirmStopPID {
		confirmPID = m.confirm.pid
	}
	for i, srv := range visible {
		name := m.serviceNameFor(srv)
		isGroup := m.groupHighlightNamespace != nil && extractNamespace(name) == *m.groupHighlightNamespace
		isConfirm := confirmActive && ((confirmTarget != "" && name == confirmTarget) ||
			(confirmPID != 0 && srv.ProcessRecord != nil && srv.ProcessRecord.PID == confirmPID))
		c := rowColorsFor(m.focus == focusRunning, i == m.selected, isConfirm, isGroup, confirmActive)
		if c.bg == "" {
			continue
		}
		style := lipgloss.NewStyle().Background(lipgloss.Color(c.bg))
		if c.fg != "" {
			style = style.Foreground(lipgloss.Color(c.fg))
		}
		lines[rowIndices[i]] = style.Render(lines[rowIndices[i]])
	}

	out := strings.Join(lines, "\n")
	if m.showHealthDetail && m.selected >= 0 && m.selected < len(visible) {
		port := 0
		if visible[m.selected].ProcessRecord != nil {
			port = visible[m.selected].ProcessRecord.Port
		}
		if d := m.healthDetails[port]; d != nil {
			out += "\n" + fitLine(fmt.Sprintf("Health detail: %s %dms %s", health.StatusIcon(d.Status), d.ResponseMs, d.Message), width)
		}
	}

	return out
}

func (m *topModel) renderManagedHeader(width int, managed []*models.ManagedService) string {
	text := fmt.Sprintf("Managed Services (%d) ", len(managed))
	fillW := width - runewidth.StringWidth(text)
	if fillW < 0 {
		fillW = 0
	}
	header := text + strings.Repeat("─", fillW)
	return lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(fitLine(header, width))
}

func (m *topModel) renderManagedList(width int, managed []*models.ManagedService) string {
	if len(managed) == 0 {
		return fitLine(`No managed services yet. Use ^A then: add myapp /path/to/app "npm run dev" 3000`, width)
	}

	portOwners := make(map[int]int)
	for _, svc := range managed {
		for _, p := range svc.Ports {
			portOwners[p]++
		}
	}

	var lines []string
	for i, svc := range managed {
		state := m.serviceStatus(svc.Name)
		if state == "stopped" {
			if _, ok := m.starting[svc.Name]; ok {
				state = "starting"
			}
		}

		// Build plain text first, then apply styling
		symbolChar := managedStatusSymbol(state)
		symbolColor := managedStatusColor(state)
		plainLine := fmt.Sprintf("%s %s [%s]", symbolChar, svc.Name, state)

		conflicting := false
		for _, p := range svc.Ports {
			if portOwners[p] > 1 {
				conflicting = true
				break
			}
		}
		if conflicting {
			plainLine = fmt.Sprintf("%s (port conflict)", plainLine)
		} else if len(svc.Ports) > 1 {
			plainLine = fmt.Sprintf("%s (ports: %v)", plainLine, svc.Ports)
		}

		// Determine background for this row via shared color logic
		confirmActive := m.activeModalKind() == modalConfirm
		isConfirm := m.confirmTargetName() == svc.Name && confirmActive
		isGroup := m.groupHighlightNamespace != nil && extractNamespace(svc.Name) == *m.groupHighlightNamespace
		c := rowColorsFor(m.focus == focusManaged, i == m.managedSel, isConfirm, isGroup, confirmActive)
		rowBg := c.bg
		rowFg := c.fg

		var line string
		if rowBg != "" {
			// Single render path for any row with background — no strings.Replace, no ANSI breakage.
			style := lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Width(width)
			if rowFg != "" {
				style = style.Foreground(lipgloss.Color(rowFg))
			}
			line = style.Render(fitLine(plainLine, width))
		} else {
			// No background — safe to color symbol separately.
			symbolStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(symbolColor)).Bold(true).Render(symbolChar)
			line = strings.Replace(plainLine, symbolChar, symbolStyled, 1)
			line = fitAnsiLine(line, width)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m *topModel) renderSelectedServiceDetails(width int, visible []*models.ServerInfo, managed []*models.ManagedService) string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	header := headerStyle.Render("Selected service details")

	// If focus is on running services, show details for the selected running service
	if m.focus == focusRunning {
		if m.selected < 0 || m.selected >= len(visible) {
			placeholder := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Select a running service to inspect details")
			return header + "\n" + fitLine(placeholder, width)
		}

		srv := visible[m.selected]
		var lines []string
		lines = append(lines, fitLine(header, width))

		// Service name
		name := m.serviceNameFor(srv)
		if name != "-" {
			lines = append(lines, fitLine(fmt.Sprintf(" Name: %s", name), width))
		}

		// Source
		if srv.Source != "" {
			lines = append(lines, fitLine(fmt.Sprintf(" Source: %s", srv.Source), width))
		}

		// Status
		if srv.Status != "" {
			lines = append(lines, fitLine(fmt.Sprintf(" Status: %s", srv.Status), width))
		}

		// Process details
		if srv.ProcessRecord != nil {
			lines = append(lines, fitLine(fmt.Sprintf(" PID: %d", srv.ProcessRecord.PID), width))
			if srv.ProcessRecord.Port > 0 {
				lines = append(lines, fitLine(fmt.Sprintf(" Port: %d (%s)", srv.ProcessRecord.Port, srv.ProcessRecord.Protocol), width))
			}
			if srv.ProcessRecord.Command != "" {
				lines = append(lines, fitLine(fmt.Sprintf(" Cmd: %s", srv.ProcessRecord.Command), width))
			}
			if srv.ProcessRecord.CWD != "" {
				lines = append(lines, fitLine(fmt.Sprintf(" Dir: %s", srv.ProcessRecord.CWD), width))
			}
			if srv.ProcessRecord.ProjectRoot != "" {
				lines = append(lines, fitLine(fmt.Sprintf(" Project: %s", srv.ProcessRecord.ProjectRoot), width))
			}
			if srv.ProcessRecord.StartTime != nil {
				lines = append(lines, fitLine(fmt.Sprintf(" Started: %s", srv.ProcessRecord.StartTime.Format("2006-01-02 15:04:05")), width))
			}
			if srv.ProcessRecord.AgentTag != nil {
				lines = append(lines, fitLine(fmt.Sprintf(" Agent: %s (%s)", srv.ProcessRecord.AgentTag.AgentName, srv.ProcessRecord.AgentTag.Source), width))
			}
		}

		// Managed service reference
		if srv.ManagedService != nil {
			lines = append(lines, fitLine(fmt.Sprintf(" Managed: %s", srv.ManagedService.Name), width))
		}

		// Health check details
		if srv.ProcessRecord != nil && srv.ProcessRecord.Port > 0 {
			if d := m.healthDetails[srv.ProcessRecord.Port]; d != nil {
				lines = append(lines, fitLine(fmt.Sprintf(" Health: %s (%dms) %s", health.StatusIcon(d.Status), d.ResponseMs, d.Message), width))
			}
		}

		// Crash info
		if srv.Status == "crashed" {
			if srv.CrashReason != "" {
				lines = append(lines, fitLine(fmt.Sprintf(" Headline: %s", srv.CrashReason), width))
			}
			for _, logLine := range nonEmptyTail(srv.CrashLogTail, 3) {
				lines = append(lines, fitLine(" "+strings.TrimSpace(logLine), width))
			}
			if srv.ManagedService != nil {
				if logPath, err := m.app.LatestServiceLogPath(srv.ManagedService.Name); err == nil && strings.TrimSpace(logPath) != "" {
					lines = append(lines, fitLine(fmt.Sprintf(" Log: %s", logPath), width))
				}
			}
		}

		return strings.Join(lines, "\n")
	}

	// Otherwise, show details for the selected managed service
	if m.managedSel < 0 || m.managedSel >= len(managed) {
		placeholder := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Select a managed service to inspect status")
		return header + "\n" + fitLine(placeholder, width)
	}

	svc := managed[m.managedSel]
	state := m.serviceStatus(svc.Name)
	if state == "stopped" {
		if _, ok := m.starting[svc.Name]; ok {
			state = "starting"
		}
	}

	symbol := lipgloss.NewStyle().Foreground(lipgloss.Color(managedStatusColor(state))).Bold(true).Render(managedStatusSymbol(state))

	var lines []string
	lines = append(lines, fitLine(header, width))
	lines = append(lines, fitLine(fmt.Sprintf(" %s %s [%s]", symbol, svc.Name, state), width))

	if srv := m.serverInfoForService(svc.Name); srv != nil && srv.Source != "" {
		lines = append(lines, fitLine(fmt.Sprintf(" Source: %s", srv.Source), width))
	}

	// Service metadata: CWD, ports, command (rendered after source, before crash context)
	if svc.CWD != "" {
		lines = append(lines, fitLine(fmt.Sprintf(" Dir: %s", svc.CWD), width))
	}
	if len(svc.Ports) > 0 {
		lines = append(lines, fitLine(fmt.Sprintf(" Port: %s", formatPorts(svc.Ports)), width))
	}
	if svc.Command != "" {
		lines = append(lines, fitLine(fmt.Sprintf(" Cmd: %s", svc.Command), width))
	}

	// Show current process info if service is running
	if srv := m.serverInfoForService(svc.Name); srv != nil && srv.ProcessRecord != nil {
		lines = append(lines, fitLine(fmt.Sprintf(" PID: %d", srv.ProcessRecord.PID), width))
		if srv.ProcessRecord.StartTime != nil {
			lines = append(lines, fitLine(fmt.Sprintf(" Started: %s", srv.ProcessRecord.StartTime.Format("2006-01-02 15:04:05")), width))
		}
		if d := m.healthDetails[srv.ProcessRecord.Port]; d != nil {
			lines = append(lines, fitLine(fmt.Sprintf(" Health: %s (%dms) %s", health.StatusIcon(d.Status), d.ResponseMs, d.Message), width))
		}
	}

	if state == "crashed" {
		if reason := m.crashReasonForService(svc.Name); reason != "" {
			lines = append(lines, fitLine(fmt.Sprintf(" Headline: %s", reason), width))
		}
		if logPath, err := m.app.LatestServiceLogPath(svc.Name); err == nil && strings.TrimSpace(logPath) != "" {
			lines = append(lines, fitLine(fmt.Sprintf(" Log: %s", logPath), width))
		}
		if srv := m.serverInfoForService(svc.Name); srv != nil {
			for _, logLine := range nonEmptyTail(srv.CrashLogTail, 3) {
				lines = append(lines, fitLine(" "+strings.TrimSpace(logLine), width))
			}
		}
	}

	return strings.Join(lines, "\n")
}

func (t *processTable) updateFocusedViewport(focus viewFocus, msg tea.Msg) tea.Cmd {
	if focus == focusManaged {
		var cmd tea.Cmd
		t.managedListVP, cmd = t.managedListVP.Update(msg)
		return cmd
	}
	var cmd tea.Cmd
	t.runningVP, cmd = t.runningVP.Update(msg)
	return cmd
}

func (t *processTable) updateViewportForTableY(viewportY int, viewportX int, msg tea.Msg) tea.Cmd {
	if viewportY < 0 {
		return nil
	}
	if viewportY < t.lastRunningHeight {
		var cmd tea.Cmd
		t.runningVP, cmd = t.runningVP.Update(msg)
		return cmd
	}
	if viewportY == t.lastRunningHeight {
		return nil
	}

	localManagedY := viewportY - t.lastRunningHeight - 1
	if localManagedY >= 0 && localManagedY < t.lastManagedHeight {
		// Route scroll to list or details viewport based on X position
		if viewportX < t.lastListWidth {
			var cmd tea.Cmd
			t.managedListVP, cmd = t.managedListVP.Update(msg)
			return cmd
		}
		var cmd tea.Cmd
		t.selectedDetailsVP, cmd = t.selectedDetailsVP.Update(msg)
		return cmd
	}
	return nil
}

// rowColors holds the foreground and background ANSI color codes for a table row.
type rowColors struct {
	bg string // empty means no background
	fg string // empty means default foreground
}

// rowColorsFor computes the visual style for a table row based on its state.
// Parameters:
//   - isFocusedPanel: this row's panel has keyboard focus
//   - isSelected: this row is the cursor selection in its panel
//   - isConfirmTarget: this row is the target of an active confirm dialog
//   - isGroupMember: this row belongs to the active group highlight namespace
//   - confirmActive: a confirm modal is currently shown
//
// Priority (first match wins):
//  1. Confirm target or selected+group during confirm → amber/orange (178/0)
//  2. Focused select  → bright blue  (57/15)
//  3. Group member    → dimmed orange (94) during confirm, blue (61) otherwise
//  4. Unfocused select → gray         (8/15)
func rowColorsFor(isFocusedPanel, isSelected, isConfirmTarget, isGroupMember, confirmActive bool) rowColors {
	switch {
	case isConfirmTarget || (confirmActive && isSelected && isGroupMember):
		return rowColors{bg: "178", fg: "0"}
	case isSelected && isFocusedPanel:
		return rowColors{bg: "57", fg: "15"}
	case isGroupMember:
		if confirmActive {
			return rowColors{bg: "94"}
		}
		return rowColors{bg: "61"}
	case isSelected:
		return rowColors{bg: "8", fg: "15"}
	default:
		return rowColors{}
	}
}

// managedClickRegion reports which managed sub-region a click falls in.
// It mirrors the X-based routing in updateViewportForTableY.
type managedRegion int

const (
	managedRegionList    managedRegion = iota // left pane: selectable items
	managedRegionDetails                      // right pane: read-only details
	managedRegionOutside                      // header separator or outside managed area
)

func (t *processTable) managedClickRegion(managedViewportY, clickX int) managedRegion {
	if managedViewportY < 0 || managedViewportY >= t.lastManagedHeight {
		return managedRegionOutside
	}
	if clickX < t.lastListWidth {
		return managedRegionList
	}
	return managedRegionDetails
}

func (t *processTable) runningYOffset() int {
	return t.runningVP.YOffset()
}

func (t *processTable) managedYOffset() int {
	return t.managedListVP.YOffset()
}

func pad(n int) string {
	return strings.Repeat(" ", n)
}

// portCell renders a port value as a fixed-width cell.
// When the port is a number, it wraps it in an OSC 8 hyperlink to http://localhost:<port>.
// When the port is "-" (no port), it renders as plain text.
// Uses ansi.StringWidth for correct width calculation with escape sequences.
func portCell(port string, width int) string {
	if port == "-" {
		return fixedCell(port, width)
	}
	return fixedHyperlinkCell(port, "http://localhost:"+port, width)
}

func (m *topModel) displayNames(servers []*models.ServerInfo) []string {
	q := strings.ToLower(strings.TrimSpace(m.currentFilterQuery()))
	if m.cachedDisplayNames != nil &&
		m.cachedDisplayNamesVersion == m.serversVersion &&
		m.cachedDisplayNamesSvcVer == m.servicesVersion &&
		m.cachedDisplayNamesQuery == q &&
		m.cachedDisplayNamesSortBy == m.sortBy &&
		m.cachedDisplayNamesReverse == m.sortReverse {
		return m.cachedDisplayNames
	}

	base := make([]string, len(servers))
	projectToSvc := make(map[string]string)
	for _, svc := range m.app.ListServices() {
		cwd := strings.TrimRight(strings.TrimSpace(svc.CWD), "/")
		if cwd != "" {
			projectToSvc[cwd] = svc.Name
		}
	}
	for i, srv := range servers {
		base[i] = m.serviceNameFor(srv)
		if base[i] == "-" && srv.ProcessRecord != nil {
			root := strings.TrimRight(strings.TrimSpace(srv.ProcessRecord.ProjectRoot), "/")
			cwd := strings.TrimRight(strings.TrimSpace(srv.ProcessRecord.CWD), "/")
			if mapped := projectToSvc[root]; mapped != "" {
				base[i] = mapped
			} else if mapped := projectToSvc[cwd]; mapped != "" {
				base[i] = mapped
			}
		}
	}

	count := make(map[string]int)
	for _, n := range base {
		count[n]++
	}
	type row struct{ idx, pid int }
	group := make(map[string][]row)
	for i, n := range base {
		group[n] = append(group[n], row{idx: i, pid: pidOf(servers[i])})
	}
	out := make([]string, len(base))
	for name, rows := range group {
		if count[name] <= 1 || name == "-" {
			for _, r := range rows {
				out[r.idx] = name
			}
			continue
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].pid < rows[j].pid })
		for i, r := range rows {
			out[r.idx] = fmt.Sprintf("%s~%d", name, i+1)
		}
	}
	m.cachedDisplayNames = out
	m.cachedDisplayNamesQuery = q
	m.cachedDisplayNamesSortBy = m.sortBy
	m.cachedDisplayNamesReverse = m.sortReverse
	m.cachedDisplayNamesVersion = m.serversVersion
	m.cachedDisplayNamesSvcVer = m.servicesVersion
	return out
}
