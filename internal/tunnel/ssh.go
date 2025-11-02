package tunnel

import (
	"fmt"
	"strings"
)

// ToSSHArgs converts a tunnel to SSH command arguments as a slice
func (t *Tunnel) ToSSHArgs() ([]string, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}

	var args []string

	switch t.Type {
	case TypeLocal:
		// Format: -L [bind_address:]port:host:hostport
		args = append(args, "-L")
		if t.BindAddress != "" {
			args = append(args, fmt.Sprintf("%s:%d:%s:%d", t.BindAddress, t.LocalPort, t.RemoteHost, t.RemotePort))
		} else {
			args = append(args, fmt.Sprintf("%d:%s:%d", t.LocalPort, t.RemoteHost, t.RemotePort))
		}

	case TypeRemote:
		// Format: -R [bind_address:]port:host:hostport
		args = append(args, "-R")
		if t.BindAddress != "" {
			args = append(args, fmt.Sprintf("%s:%d:%s:%d", t.BindAddress, t.LocalPort, t.RemoteHost, t.RemotePort))
		} else {
			args = append(args, fmt.Sprintf("%d:%s:%d", t.LocalPort, t.RemoteHost, t.RemotePort))
		}

	case TypeDynamic:
		// Format: -D [bind_address:]port
		args = append(args, "-D")
		if t.BindAddress != "" {
			args = append(args, fmt.Sprintf("%s:%d", t.BindAddress, t.LocalPort))
		} else {
			args = append(args, fmt.Sprintf("%d", t.LocalPort))
		}

	default:
		return nil, fmt.Errorf("unknown tunnel type: %s", t.Type)
	}

	return args, nil
}

// TunnelsToSSHArgs converts multiple tunnels to SSH command arguments
func TunnelsToSSHArgs(tunnels []Tunnel) ([]string, error) {
	var args []string

	for _, tunnel := range tunnels {
		tunnelArgs, err := tunnel.ToSSHArgs()
		if err != nil {
			return nil, fmt.Errorf("failed to convert tunnel '%s': %w", tunnel.Name, err)
		}
		args = append(args, tunnelArgs...)
	}

	return args, nil
}

// FormatTunnelsSummary creates a human-readable summary of active tunnels
func FormatTunnelsSummary(tunnels []Tunnel) string {
	if len(tunnels) == 0 {
		return "No tunnels active"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Active tunnels (%d):", len(tunnels)))

	for _, tunnel := range tunnels {
		line := fmt.Sprintf("  • %s: %s", tunnel.Name, tunnel.DisplayString())
		if tunnel.Description != "" {
			line += fmt.Sprintf(" (%s)", tunnel.Description)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
