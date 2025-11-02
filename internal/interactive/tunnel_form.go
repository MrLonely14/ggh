package interactive

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MrLonely14/ggh/internal/tunnel"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tunnelFormModel struct {
	tunnel     *tunnel.Tunnel
	editing    bool
	focusIndex int
	inputs     []formInput
	err        string
	submitted  bool
	cancelled  bool
}

type formInput struct {
	label       string
	value       string
	placeholder string
	required    bool
}

const (
	inputName int = iota
	inputType
	inputLocalPort
	inputRemoteHost
	inputRemotePort
	inputBindAddress
	inputDescription
)

func newTunnelForm(t *tunnel.Tunnel) *tunnelFormModel {
	editing := t != nil

	inputs := []formInput{
		{label: "Name", placeholder: "my-tunnel", required: true},
		{label: "Type", placeholder: "local/remote/dynamic", required: true},
		{label: "Local Port", placeholder: "8080", required: true},
		{label: "Remote Host", placeholder: "localhost"},
		{label: "Remote Port", placeholder: "80"},
		{label: "Bind Address", placeholder: "0.0.0.0 (optional)"},
		{label: "Description", placeholder: "Tunnel description"},
	}

	if editing {
		inputs[inputName].value = t.Name
		inputs[inputType].value = string(t.Type)
		inputs[inputLocalPort].value = strconv.Itoa(t.LocalPort)
		inputs[inputRemoteHost].value = t.RemoteHost
		inputs[inputRemotePort].value = strconv.Itoa(t.RemotePort)
		inputs[inputBindAddress].value = t.BindAddress
		inputs[inputDescription].value = t.Description
	}

	return &tunnelFormModel{
		tunnel:  t,
		editing: editing,
		inputs:  inputs,
	}
}

func (m *tunnelFormModel) Init() tea.Cmd {
	return nil
}

func (m *tunnelFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			return m, nil

		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			return m, nil

		case "enter":
			if msg.Alt {
				// Alt+Enter submits the form
				return m, m.submit()
			}
			// Regular Enter moves to next field
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			return m, nil

		case "ctrl+s":
			return m, m.submit()

		case "backspace":
			if len(m.inputs[m.focusIndex].value) > 0 {
				m.inputs[m.focusIndex].value = m.inputs[m.focusIndex].value[:len(m.inputs[m.focusIndex].value)-1]
			}
			m.err = ""
			return m, nil

		default:
			if len(msg.Runes) > 0 {
				m.inputs[m.focusIndex].value += string(msg.Runes)
				m.err = ""
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *tunnelFormModel) View() string {
	if m.submitted || m.cancelled {
		return ""
	}

	var b strings.Builder

	title := "Create New Tunnel"
	if m.editing {
		title = "Edit Tunnel"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Padding(1, 0)

	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	for i, input := range m.inputs {
		// Skip remote host/port for dynamic tunnels
		if i == inputRemoteHost || i == inputRemotePort {
			if m.inputs[inputType].value == "dynamic" {
				continue
			}
		}

		isFocused := i == m.focusIndex

		labelStyle := lipgloss.NewStyle().
			Width(15).
			Foreground(lipgloss.Color("240"))

		if input.required {
			labelStyle = labelStyle.Foreground(lipgloss.Color("205"))
		}

		inputStyle := lipgloss.NewStyle().
			Width(40).
			Foreground(lipgloss.Color("255"))

		if isFocused {
			inputStyle = inputStyle.
				Foreground(lipgloss.Color("212")).
				Bold(true)
		}

		label := labelStyle.Render(input.label + ":")

		value := input.value
		if value == "" {
			value = input.placeholder
			inputStyle = inputStyle.Foreground(lipgloss.Color("240")).Italic(true)
		}

		if isFocused {
			value = value + "▌"
		}

		b.WriteString(fmt.Sprintf("%s %s\n", label, inputStyle.Render(value)))
	}

	if m.err != "" {
		errStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Padding(1, 0)
		b.WriteString("\n")
		b.WriteString(errStyle.Render("Error: " + m.err))
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(1, 0)

	help := "\ntab/↑/↓ navigate • ctrl+s save • esc cancel"
	b.WriteString(helpStyle.Render(help))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

func (m *tunnelFormModel) submit() tea.Cmd {
	// Validate inputs
	if m.inputs[inputName].value == "" {
		m.err = "Name is required"
		return nil
	}

	tunnelType := m.inputs[inputType].value
	if tunnelType != "local" && tunnelType != "remote" && tunnelType != "dynamic" {
		m.err = "Type must be 'local', 'remote', or 'dynamic'"
		return nil
	}

	if m.inputs[inputLocalPort].value == "" {
		m.err = "Local port is required"
		return nil
	}

	localPort, err := strconv.Atoi(m.inputs[inputLocalPort].value)
	if err != nil || localPort < 1 || localPort > 65535 {
		m.err = "Local port must be a number between 1 and 65535"
		return nil
	}

	// For local/remote, validate remote host and port
	var remotePort int
	if tunnelType == "local" || tunnelType == "remote" {
		if m.inputs[inputRemoteHost].value == "" {
			m.err = "Remote host is required for " + tunnelType + " forwarding"
			return nil
		}
		if m.inputs[inputRemotePort].value == "" {
			m.err = "Remote port is required for " + tunnelType + " forwarding"
			return nil
		}
		remotePort, err = strconv.Atoi(m.inputs[inputRemotePort].value)
		if err != nil || remotePort < 1 || remotePort > 65535 {
			m.err = "Remote port must be a number between 1 and 65535"
			return nil
		}
	}

	// Create or update tunnel
	t := &tunnel.Tunnel{
		Name:        m.inputs[inputName].value,
		Type:        tunnel.TunnelType(tunnelType),
		LocalPort:   localPort,
		RemoteHost:  m.inputs[inputRemoteHost].value,
		RemotePort:  remotePort,
		BindAddress: m.inputs[inputBindAddress].value,
		Description: m.inputs[inputDescription].value,
	}

	if m.editing && m.tunnel != nil {
		t.ID = m.tunnel.ID
		t.CreatedAt = m.tunnel.CreatedAt
		err = tunnel.Update(t)
	} else {
		err = tunnel.Create(t)
	}

	if err != nil {
		m.err = err.Error()
		return nil
	}

	m.submitted = true
	return tea.Quit
}

// updateForm handles form updates within the tunnel model
func (m tunnelModel) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Update the form
	_, cmd = m.formModel.Update(msg)

	// Check if form is done
	if m.formModel.submitted || m.formModel.cancelled {
		m.showingForm = false

		if m.formModel.submitted {
			// Reload tunnels
			tunnels, err := tunnel.FetchAll()
			if err == nil {
				m.tunnels = tunnels
				m.allRows = tunnelsToRows(tunnels)
				m.filteredRows = m.allRows
				m.table.SetRows(m.allRows)
			}
		}

		m.formModel = nil
		return m, cmd
	}

	return m, cmd
}
