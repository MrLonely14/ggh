package cmd

import (
	"fmt"
	"github.com/MrLonely14/ggh/internal/command"
	"github.com/MrLonely14/ggh/internal/config"
	"github.com/MrLonely14/ggh/internal/history"
	"github.com/MrLonely14/ggh/internal/interactive"
	"github.com/MrLonely14/ggh/internal/ssh"
	"github.com/MrLonely14/ggh/internal/theme"
	"github.com/MrLonely14/ggh/internal/tunnel"
	"github.com/charmbracelet/bubbles/table"
	"os"
)

func Main() {
	command.CheckSSH()

	args := os.Args[1:]

	action, value := command.Which()
	switch action {
	case command.InteractiveHistory:
		args = interactive.History()
	case command.InteractiveConfig:
		args = interactive.Config("")
	case command.InteractiveConfigWithSearch:
		args = interactive.Config(value)
	case command.ListHistory:
		history.Print()
		return
	case command.ListConfig:
		config.Print()
		return
	case command.InteractiveTunnels:
		// Interactive tunnel management (create/edit/delete/select)
		interactive.SelectTunnels(true)
		return
	case command.ListTunnels:
		// List all tunnels in a table
		printTunnels()
		return
	case command.SelectTunnels:
		// Select tunnels and apply to next SSH connection
		selectedTunnels := interactive.SelectTunnels(true)
		if len(selectedTunnels) == 0 {
			return
		}
		// Now select SSH connection
		args = interactive.History()
		// Prepend tunnel args to SSH args
		args = prependTunnelArgs(selectedTunnels, args)
	default:
		history.AddHistoryFromArgs(args)
	}
	ssh.Run(args)
}

// printTunnels displays all tunnels in a formatted table
func printTunnels() {
	tunnels, err := tunnel.FetchAll()
	if err != nil {
		fmt.Printf("Error loading tunnels: %v\n", err)
		os.Exit(1)
	}

	if len(tunnels) == 0 {
		fmt.Println("No tunnels configured. Use 'ggh tunnels' to create one.")
		return
	}

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
		}
		rows = append(rows, row)
	}

	fmt.Println(theme.PrintTable(rows, theme.TunnelTable))
}

// prependTunnelArgs converts tunnels to SSH arguments and prepends them to existing args
func prependTunnelArgs(tunnels []tunnel.Tunnel, args []string) []string {
	tunnelArgs, err := tunnel.TunnelsToSSHArgs(tunnels)
	if err != nil {
		fmt.Printf("Error converting tunnels: %v\n", err)
		return args
	}

	// Update last used for all selected tunnels
	ids := make([]string, len(tunnels))
	for i, t := range tunnels {
		ids[i] = t.ID
	}
	_ = tunnel.UpdateLastUsedBatch(ids)

	// Print tunnel summary
	fmt.Println(tunnel.FormatTunnelsSummary(tunnels))
	fmt.Println()

	return append(tunnelArgs, args...)
}
