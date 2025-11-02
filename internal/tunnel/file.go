package tunnel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	tunnelsDir      = ".ggh"
	tunnelsFileName = "tunnels.json"
)

// TunnelStore represents the persistent storage structure
type TunnelStore struct {
	Tunnels []Tunnel `json:"tunnels"`
}

// GetTunnelsFilePath returns the absolute path to the tunnels file
func GetTunnelsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, tunnelsDir, tunnelsFileName), nil
}

// ensureTunnelsDir creates the .ggh directory if it doesn't exist
func ensureTunnelsDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	gghDir := filepath.Join(homeDir, tunnelsDir)

	if _, err := os.Stat(gghDir); os.IsNotExist(err) {
		if err := os.MkdirAll(gghDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", gghDir, err)
		}
	}

	return nil
}

// LoadTunnels reads tunnels from the tunnels.json file
func LoadTunnels() ([]Tunnel, error) {
	filePath, err := GetTunnelsFilePath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []Tunnel{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tunnels file: %w", err)
	}

	// Handle empty file
	if len(data) == 0 {
		return []Tunnel{}, nil
	}

	var store TunnelStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse tunnels file: %w", err)
	}

	return store.Tunnels, nil
}

// SaveTunnels writes tunnels to the tunnels.json file
func SaveTunnels(tunnels []Tunnel) error {
	if err := ensureTunnelsDir(); err != nil {
		return err
	}

	filePath, err := GetTunnelsFilePath()
	if err != nil {
		return err
	}

	store := TunnelStore{
		Tunnels: tunnels,
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tunnels: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tunnels file: %w", err)
	}

	return nil
}
