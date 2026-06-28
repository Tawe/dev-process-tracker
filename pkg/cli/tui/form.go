package tui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/devports/devpt/pkg/models"
)

const (
	formFieldName = iota
	formFieldDir
	formFieldCommand
	formFieldPorts
	formFieldCount
)

var formFieldLabels = [formFieldCount]string{"Name", "Dir", "Command", "Ports"}

// formSubmit is the validated intent produced when the user submits the form.
// Pure data — the topModel applies it via AppDeps.
type formSubmit struct {
	isEdit  bool
	oldName string
	name    string
	cwd     string
	command string
	ports   []int
	rename  bool // isEdit && name != oldName
}

// formState holds the Add/Edit modal form. Key routing and submit validation
// are pure and testable without a tea.Program.
type formState struct {
	fields   [formFieldCount]textinput.Model
	focus    int
	isEdit   bool
	origName string
	running  bool
	err      string
}

func newFormState() formState {
	var fs formState
	for i := range fs.fields {
		ti := textinput.New()
		ti.Prompt = ""
		ti.CharLimit = 0
		fs.fields[i] = ti
	}
	fs.focusField(formFieldName)
	return fs
}

// formForAdd returns an empty Add form.
func formForAdd() formState {
	fs := newFormState()
	fs.fields[formFieldPorts].Placeholder = "3000, 8080"
	return fs
}

// formForEdit pre-fills the form from an existing managed service.
func formForEdit(svc *models.ManagedService) formState {
	fs := newFormState()
	fs.isEdit = true
	fs.origName = svc.Name
	fs.running = svc.LastPID != nil
	fs.fields[formFieldName].SetValue(svc.Name)
	fs.fields[formFieldDir].SetValue(svc.CWD)
	fs.fields[formFieldCommand].SetValue(svc.Command)
	ports := make([]string, 0, len(svc.Ports))
	for _, p := range svc.Ports {
		ports = append(ports, strconv.Itoa(p))
	}
	fs.fields[formFieldPorts].SetValue(strings.Join(ports, ", "))
	return fs
}

// focusField blurs all fields and focuses index i (wraps in either direction).
func (fs *formState) focusField(i int) {
	n := formFieldCount
	fs.focus = ((i % n) + n) % n
	for j := range fs.fields {
		if j == fs.focus {
			fs.fields[j].Focus()
		} else {
			fs.fields[j].Blur()
		}
	}
}

// runningHint returns the advisory shown when editing a running service.
func (fs *formState) runningHint() string {
	if fs.running {
		return "Applies on next start/restart"
	}
	return ""
}

// update routes a key while the form is open. It returns a non-nil submit only
// on a valid Enter, and quit=true on Esc.
func (fs *formState) update(msg tea.KeyPressMsg) (submit *formSubmit, quit bool, cmd tea.Cmd) {
	switch msg.String() {
	case "esc":
		return nil, true, nil
	case "tab":
		fs.focusField(fs.focus + 1)
		return nil, false, nil
	case "shift+tab":
		fs.focusField(fs.focus - 1)
		return nil, false, nil
	case "enter":
		intent, errMsg := fs.submit()
		if errMsg != "" {
			fs.err = errMsg
			return nil, false, nil
		}
		return intent, false, nil
	default:
		fs.err = ""
		var c tea.Cmd
		fs.fields[fs.focus], c = fs.fields[fs.focus].Update(msg)
		return nil, false, c
	}
}

// submit validates and parses the form into a formSubmit, or returns an error
// message string (the form stays open so the user can fix it).
func (fs *formState) submit() (*formSubmit, string) {
	name := strings.TrimSpace(fs.fields[formFieldName].Value())
	cwd := strings.TrimSpace(fs.fields[formFieldDir].Value())
	command := fs.fields[formFieldCommand].Value()
	if err := models.ValidateManagedServiceFields(name, command); err != nil {
		return nil, err.Error()
	}
	ports, err := parseFormPorts(fs.fields[formFieldPorts].Value())
	if err != nil {
		return nil, err.Error()
	}
	intent := &formSubmit{
		isEdit:  fs.isEdit,
		oldName: fs.origName,
		name:    name,
		cwd:     cwd,
		command: command,
		ports:   ports,
	}
	if fs.isEdit && name != fs.origName {
		intent.rename = true
	}
	return intent, ""
}

// parseFormPorts parses comma- and/or space-separated integer ports.
func parseFormPorts(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
	ports := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %s", p)
		}
		ports = append(ports, n)
	}
	return ports, nil
}

// --- topModel glue ---

func (m *topModel) openAddForm() {
	fs := formForAdd()
	m.form = &fs
}

func (m *topModel) openEditForm() {
	managed := m.managedServices()
	if m.managedSel < 0 || m.managedSel >= len(managed) {
		m.cmdStatus = "No managed service selected"
		return
	}
	fs := formForEdit(managed[m.managedSel])
	m.form = &fs
}

func (m *topModel) closeForm() { m.form = nil }

// handleFormKey routes a key to the open form and applies submit/cancel.
func (m *topModel) handleFormKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	intent, quit, cmd := m.form.update(msg)
	if quit {
		m.closeForm()
		return m, nil
	}
	if intent != nil {
		status, err := m.applyFormSubmit(intent)
		if err != nil {
			m.form.err = err.Error()
			return m, nil
		}
		m.closeForm()
		m.refresh()
		m.cmdStatus = status
		return m, nil
	}
	return m, cmd
}

// applyFormSubmit persists a validated submit intent, returning a status
// message. On error the form stays open and shows the error.
func (m *topModel) applyFormSubmit(s *formSubmit) (string, error) {
	if !s.isEdit {
		if err := m.app.RegisterService(s.name, s.cwd, s.command, s.ports); err != nil {
			return "", err
		}
		return fmt.Sprintf("Added %q", s.name), nil
	}
	// Edit: rename first (the risky, all-or-nothing op), then update fields.
	target := s.oldName
	if s.rename {
		if err := m.app.RenameService(s.oldName, s.name); err != nil {
			return "", err
		}
		target = s.name
	}
	if err := m.app.UpdateServiceFields(target, s.cwd, s.command, s.ports); err != nil {
		return "", err
	}
	if s.rename {
		return fmt.Sprintf("Renamed %q → %q", s.oldName, s.name), nil
	}
	return fmt.Sprintf("Updated %q", target), nil
}

func (m *topModel) renderFormModal(width int) string {
	if m.form == nil {
		return ""
	}
	fs := m.form
	title := "Add service"
	if fs.isEdit {
		title = "Edit service"
	}
	var b strings.Builder
	for i, label := range formFieldLabels {
		marker := " "
		if i == fs.focus {
			marker = "▸"
		}
		fmt.Fprintf(&b, "%s %-7s %s\n", marker, label, fs.fields[i].View())
	}
	body := strings.TrimRight(b.String(), "\n")
	if fs.err != "" {
		body += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fs.err)
	}
	hint := "Tab next field · Enter save · Esc cancel"
	if h := fs.runningHint(); h != "" {
		hint = h + " · " + hint
	}
	accent := "12"
	if fs.isEdit {
		accent = "11"
	}
	return renderModal(title, body, hint, width, 64, accent)
}
