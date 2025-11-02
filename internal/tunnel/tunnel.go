package tunnel

import (
	"fmt"
	"strconv"
	"strings"
)

// TunnelType represents the type of SSH tunnel
type TunnelType string

const (
	// Local forwarding (-L): Forward local port to remote destination
	TypeLocal TunnelType = "local"
	// Remote forwarding (-R): Forward remote port to local destination
	TypeRemote TunnelType = "remote"
	// Dynamic forwarding (-D): SOCKS proxy
	TypeDynamic TunnelType = "dynamic"
)

// Tunnel represents a saved SSH port forwarding configuration
type Tunnel struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        TunnelType `json:"type"`
	LocalPort   int        `json:"local_port"`
	RemoteHost  string     `json:"remote_host,omitempty"`  // For local/remote forwarding
	RemotePort  int        `json:"remote_port,omitempty"`  // For local/remote forwarding
	BindAddress string     `json:"bind_address,omitempty"` // Optional bind address
	CreatedAt   string     `json:"created_at"`
	LastUsed    string     `json:"last_used,omitempty"`
}

// Validate checks if the tunnel configuration is valid
func (t *Tunnel) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("tunnel name cannot be empty")
	}

	if t.Type != TypeLocal && t.Type != TypeRemote && t.Type != TypeDynamic {
		return fmt.Errorf("invalid tunnel type: %s", t.Type)
	}

	if t.LocalPort < 1 || t.LocalPort > 65535 {
		return fmt.Errorf("invalid local port: %d (must be 1-65535)", t.LocalPort)
	}

	// For local and remote tunneling, we need remote host and port
	if t.Type == TypeLocal || t.Type == TypeRemote {
		if t.RemoteHost == "" {
			return fmt.Errorf("%s forwarding requires remote host", t.Type)
		}
		if t.RemotePort < 1 || t.RemotePort > 65535 {
			return fmt.Errorf("invalid remote port: %d (must be 1-65535)", t.RemotePort)
		}
	}

	return nil
}

// ToSSHFlag converts the tunnel to an SSH command line flag
func (t *Tunnel) ToSSHFlag() (string, error) {
	if err := t.Validate(); err != nil {
		return "", err
	}

	switch t.Type {
	case TypeLocal:
		// Format: -L [bind_address:]port:host:hostport
		if t.BindAddress != "" {
			return fmt.Sprintf("-L %s:%d:%s:%d", t.BindAddress, t.LocalPort, t.RemoteHost, t.RemotePort), nil
		}
		return fmt.Sprintf("-L %d:%s:%d", t.LocalPort, t.RemoteHost, t.RemotePort), nil

	case TypeRemote:
		// Format: -R [bind_address:]port:host:hostport
		if t.BindAddress != "" {
			return fmt.Sprintf("-R %s:%d:%s:%d", t.BindAddress, t.LocalPort, t.RemoteHost, t.RemotePort), nil
		}
		return fmt.Sprintf("-R %d:%s:%d", t.LocalPort, t.RemoteHost, t.RemotePort), nil

	case TypeDynamic:
		// Format: -D [bind_address:]port
		if t.BindAddress != "" {
			return fmt.Sprintf("-D %s:%d", t.BindAddress, t.LocalPort), nil
		}
		return fmt.Sprintf("-D %d", t.LocalPort), nil

	default:
		return "", fmt.Errorf("unknown tunnel type: %s", t.Type)
	}
}

// DisplayString returns a human-readable string representation of the tunnel
func (t *Tunnel) DisplayString() string {
	switch t.Type {
	case TypeLocal:
		return fmt.Sprintf("Local: %d → %s:%d", t.LocalPort, t.RemoteHost, t.RemotePort)
	case TypeRemote:
		return fmt.Sprintf("Remote: %d → %s:%d", t.LocalPort, t.RemoteHost, t.RemotePort)
	case TypeDynamic:
		return fmt.Sprintf("Dynamic SOCKS: %d", t.LocalPort)
	default:
		return "Unknown"
	}
}

// ParseSSHFlag parses an SSH tunnel flag string into a Tunnel struct
// Supports: -L 8080:localhost:80, -R 8080:localhost:80, -D 1080
func ParseSSHFlag(flag string) (*Tunnel, error) {
	parts := strings.Fields(flag)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid flag format: %s", flag)
	}

	tunnel := &Tunnel{}

	switch parts[0] {
	case "-L":
		tunnel.Type = TypeLocal
	case "-R":
		tunnel.Type = TypeRemote
	case "-D":
		tunnel.Type = TypeDynamic
	default:
		return nil, fmt.Errorf("unknown flag type: %s", parts[0])
	}

	// Parse the port specification
	spec := parts[1]

	if tunnel.Type == TypeDynamic {
		// Dynamic: [bind_address:]port
		if strings.Contains(spec, ":") {
			specParts := strings.Split(spec, ":")
			tunnel.BindAddress = specParts[0]
			port, err := strconv.Atoi(specParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", specParts[1])
			}
			tunnel.LocalPort = port
		} else {
			port, err := strconv.Atoi(spec)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", spec)
			}
			tunnel.LocalPort = port
		}
	} else {
		// Local/Remote: [bind_address:]port:host:hostport
		specParts := strings.Split(spec, ":")

		switch len(specParts) {
		case 3:
			// port:host:hostport
			port, err := strconv.Atoi(specParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", specParts[0])
			}
			tunnel.LocalPort = port
			tunnel.RemoteHost = specParts[1]
			remotePort, err := strconv.Atoi(specParts[2])
			if err != nil {
				return nil, fmt.Errorf("invalid remote port: %s", specParts[2])
			}
			tunnel.RemotePort = remotePort

		case 4:
			// bind_address:port:host:hostport
			tunnel.BindAddress = specParts[0]
			port, err := strconv.Atoi(specParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", specParts[1])
			}
			tunnel.LocalPort = port
			tunnel.RemoteHost = specParts[2]
			remotePort, err := strconv.Atoi(specParts[3])
			if err != nil {
				return nil, fmt.Errorf("invalid remote port: %s", specParts[3])
			}
			tunnel.RemotePort = remotePort

		default:
			return nil, fmt.Errorf("invalid port specification format: %s", spec)
		}
	}

	// Don't validate here since name is not required when parsing flags
	// Validation should be done separately when creating/updating tunnels
	return tunnel, nil
}
