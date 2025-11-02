package tunnel

import (
	"fmt"
	"sort"
)

// FetchAll retrieves all saved tunnels
func FetchAll() ([]Tunnel, error) {
	tunnels, err := LoadTunnels()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tunnels: %w", err)
	}

	// Sort by last used (most recent first), then by name
	sort.Slice(tunnels, func(i, j int) bool {
		if tunnels[i].LastUsed != tunnels[j].LastUsed {
			return tunnels[i].LastUsed > tunnels[j].LastUsed
		}
		return tunnels[i].Name < tunnels[j].Name
	})

	return tunnels, nil
}

// FetchByID retrieves a tunnel by its ID
func FetchByID(id string) (*Tunnel, error) {
	tunnels, err := LoadTunnels()
	if err != nil {
		return nil, err
	}

	for _, tunnel := range tunnels {
		if tunnel.ID == id {
			return &tunnel, nil
		}
	}

	return nil, fmt.Errorf("tunnel not found: %s", id)
}

// FetchByName retrieves a tunnel by its name
func FetchByName(name string) (*Tunnel, error) {
	tunnels, err := LoadTunnels()
	if err != nil {
		return nil, err
	}

	for _, tunnel := range tunnels {
		if tunnel.Name == name {
			return &tunnel, nil
		}
	}

	return nil, fmt.Errorf("tunnel not found: %s", name)
}

// FetchByIDs retrieves multiple tunnels by their IDs
func FetchByIDs(ids []string) ([]Tunnel, error) {
	allTunnels, err := LoadTunnels()
	if err != nil {
		return nil, err
	}

	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	var result []Tunnel
	for _, tunnel := range allTunnels {
		if idMap[tunnel.ID] {
			result = append(result, tunnel)
		}
	}

	if len(result) != len(ids) {
		return result, fmt.Errorf("some tunnels not found (requested: %d, found: %d)", len(ids), len(result))
	}

	return result, nil
}

// FetchByType retrieves all tunnels of a specific type
func FetchByType(tunnelType TunnelType) ([]Tunnel, error) {
	allTunnels, err := LoadTunnels()
	if err != nil {
		return nil, err
	}

	var result []Tunnel
	for _, tunnel := range allTunnels {
		if tunnel.Type == tunnelType {
			result = append(result, tunnel)
		}
	}

	return result, nil
}
