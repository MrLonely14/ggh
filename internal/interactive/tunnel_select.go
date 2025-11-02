package interactive

import (
	"fmt"
	"os"
	"strings"

	"github.com/byawitz/ggh/internal/settings"
	"github.com/byawitz/ggh/internal/theme"
	"github.com/byawitz/ggh/internal/tunnel"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tunnelModel struct {
	table        table.Model
	allRows      []table.Row
	filteredRows []table.Row
	filtering    bool
	filterText   string
	selectedIDs  map[string]bool // For multi-select
	tunnels      []tunnel.Tunnel
	exit         bool
	windowWidth  int
	windowHeight int
	tableWidth   int
	tableHeight  int
	multiSelect  bool
	showingForm  bool
	formModel    *tunnelFormModel
}

func (m tunnelModel) Init() tea.Cmd { return nil }

func (m tunnelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If form is showing, delegate to form
	if m.showingForm && m.formModel != nil {
		return m.updateForm(msg)
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		w, h, cols := theme.AdjustTableDimensions(
			m.table.Columns(),
			m.windowWidth,
			m.windowHeight,
		)

		m.tableWidth = w
		m.tableHeight = h

		m.table.SetColumns(cols)
		m.resetTableHeight()

		if settings.Get().Fullscreen {
			return m, tea.EnterAltScreen
		}

		return m, tea.ExitAltScreen

	case tea.KeyMsg:
		if m.filtering {
			switch msg.Type {
			case tea.KeyRunes:
				m.filterText += string(msg.Runes)
				m.applyFilter()
				return m, nil
			case tea.KeyBackspace:
				if len(m.filterText) > 0 {
					m.filterText = m.filterText[:len(m.filterText)-1]
					m.applyFilter()
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "/":
			if !m.filtering {
				m.filtering = true
				m.filterText = ""
				m.applyFilter()
				return m, nil
			}
			return m, nil

		case "n":
			// Create new tunnel
			m.showingForm = true
			m.formModel = newTunnelForm(nil)
			return m, nil

		case "e":
			// Edit selected tunnel
			selectedRow := m.table.SelectedRow()
			if selectedRow == nil {
				return m, nil
			}
			tunnelID := selectedRow[len(selectedRow)-1] // ID is stored in hidden last column
			for _, t := range m.tunnels {
				if t.ID == tunnelID {
					m.showingForm = true
					m.formModel = newTunnelForm(&t)
					return m, nil
				}
			}
			return m, nil

		case "d":
			// Delete selected tunnel
			selectedRow := m.table.SelectedRow()
			if selectedRow == nil {
				return m, nil
			}
			tunnelID := selectedRow[len(selectedRow)-1]

			if err := tunnel.Delete(tunnelID); err == nil {
				// Remove from display
				rows := []table.Row{}
				newTunnels := []tunnel.Tunnel{}
				for i, row := range m.table.Rows() {
					if row[len(row)-1] != tunnelID {
						rows = append(rows, row)
						newTunnels = append(newTunnels, m.tunnels[i])
					}
				}
				m.allRows = rows
				m.tunnels = newTunnels
				m.table.SetRows(m.allRows)

				if len(m.table.Rows()) == 0 {
					m.exit = true
					return m, tea.Quit
				}
			}
			return m, nil

		case " ":
			// Toggle selection for multi-select
			if m.multiSelect {
				selectedRow := m.table.SelectedRow()
				if selectedRow != nil {
					tunnelID := selectedRow[len(selectedRow)-1]
					if m.selectedIDs[tunnelID] {
						delete(m.selectedIDs, tunnelID)
					} else {
						m.selectedIDs[tunnelID] = true
					}
				}
			}
			return m, nil

		case "w":
			// Toggle fullscreen mode
			newsettings := settings.Get()
			newsettings.Fullscreen = !newsettings.Fullscreen
			if err := settings.Save(newsettings); err == nil {
				m.resetTableHeight()

				if newsettings.Fullscreen {
					return m, tea.EnterAltScreen
				}

				return m, tea.ExitAltScreen
			}
			return m, nil

		case "q", "ctrl+c", "esc":
			if m.filtering {
				m.stopFiltering()
				return m, nil
			}
			m.exit = true
			return m, tea.Quit

		case "enter":
			if m.multiSelect && len(m.selectedIDs) > 0 {
				// Return all selected tunnels
				return m, tea.Quit
			}
			// Return single selected tunnel
			selectedRow := m.table.SelectedRow()
			if selectedRow != nil {
				tunnelID := selectedRow[len(selectedRow)-1]
				m.selectedIDs[tunnelID] = true
			}
			return m, tea.Quit
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m tunnelModel) View() string {
	if m.showingForm && m.formModel != nil {
		return m.formModel.View()
	}

	if m.exit && !m.multiSelect {
		return ""
	}

	if m.tableHeight < 3 {
		msg := fmt.Sprintf(
			"Too small (%d lines). Need ≥ 6 lines.",
			m.windowHeight,
		)

		styled := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FF0060", Dark: "#FF79C6"}).
			Render(msg)

		pad := max((m.windowWidth-lipgloss.Width(styled))/2, 0)
		centered := strings.Repeat(" ", pad) + styled

		return centered
	}

	return theme.BaseStyle.Render(m.table.View()) + "\n" + m.helpView()
}

func (m tunnelModel) helpView() string {
	var blocks []string

	if m.filtering {
		blocks = append(blocks, "esc quit filter")
	} else {
		km := table.DefaultKeyMap()
		blocks = append(blocks,
			fmt.Sprintf("%s %s", km.LineUp.Help().Key, km.LineUp.Help().Desc),
			fmt.Sprintf("%s %s", km.LineDown.Help().Key, km.LineDown.Help().Desc),
			"n new",
			"e edit",
			"d delete",
		)

		if m.multiSelect {
			blocks = append(blocks, "space select")
		}

		blocks = append(blocks,
			"w window/full",
			"/ filter",
			"q quit",
		)
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#B2B2B2", Dark: "#4A4A4A"}).
		Render(strings.Join(blocks, " • "))

	if m.filtering {
		prompt := lipgloss.NewStyle().
			Foreground(lipgloss.Color("57")).
			Render("/" + m.filterText)
		return " " + help + prompt
	}

	selectedCount := len(m.selectedIDs)
	if selectedCount > 0 {
		countMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Render(fmt.Sprintf(" [%d selected]", selectedCount))
		return " " + help + countMsg
	}

	return " " + help
}

func (m *tunnelModel) applyFilter() {
	if m.filterText == "" {
		m.filteredRows = m.allRows
	} else {
		var out []table.Row
		lowerFilter := strings.ToLower(m.filterText)
		for _, row := range m.allRows {
			rowStr := strings.ToLower(strings.Join(row[:len(row)-1], " ")) // Exclude ID column
			if strings.Contains(rowStr, lowerFilter) {
				out = append(out, row)
			}
		}
		m.filteredRows = out
	}
	m.table.SetRows(m.filteredRows)
}

func (m *tunnelModel) stopFiltering() {
	m.filtering = false
	m.filterText = ""
	m.filteredRows = m.allRows
	m.table.SetRows(m.filteredRows)
}

func (m *tunnelModel) resetTableHeight() {
	m.table.SetHeight(theme.GetTableHeight(
		m.tableHeight,
		len(m.filteredRows),
	))
	m.table.SetWidth(m.tableWidth)
}

// SelectTunnels shows an interactive tunnel selection interface
// If multiSelect is true, users can select multiple tunnels
func SelectTunnels(multiSelect bool) []tunnel.Tunnel {
	tunnels, err := tunnel.FetchAll()
	if err != nil {
		fmt.Printf("Error loading tunnels: %v\n", err)
		return nil
	}

	if len(tunnels) == 0 {
		fmt.Println("No tunnels configured. Use 'n' to create a new tunnel.")
	}

	rows := tunnelsToRows(tunnels)

	t := table.New(
		table.WithColumns(theme.GetColumns(theme.TunnelTable)),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = theme.HeaderStyle
	s.Selected = theme.SelectedStyle

	t.SetStyles(s)

	m := tunnelModel{
		table:        t,
		allRows:      rows,
		filteredRows: rows,
		tunnels:      tunnels,
		multiSelect:  multiSelect,
		selectedIDs:  make(map[string]bool),
	}

	var p *tea.Program
	if settings.Get().Fullscreen {
		p = tea.NewProgram(m, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(m)
	}

	finalModel, err := p.Run()
	if err != nil {
		fmt.Println("Error running tunnel selector:", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(tunnelModel); ok {
		if m.exit {
			return nil
		}

		// Return selected tunnels
		var selected []tunnel.Tunnel
		for _, t := range m.tunnels {
			if m.selectedIDs[t.ID] {
				selected = append(selected, t)
			}
		}
		return selected
	}

	return nil
}

// tunnelsToRows converts tunnels to table rows
func tunnelsToRows(tunnels []tunnel.Tunnel) []table.Row {
	rows := make([]table.Row, 0, len(tunnels))

	for _, t := range tunnels {
		remote := "-"
		if t.Type != tunnel.TypeDynamic {
			remote = fmt.Sprintf("%s:%d", t.RemoteHost, t.RemotePort)
		}

		desc := t.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}

		row := table.Row{
			t.Name,
			string(t.Type),
			fmt.Sprintf("%d", t.LocalPort),
			remote,
			desc,
			t.ID, // Hidden column for ID
		}
		rows = append(rows, row)
	}

	return rows
}
