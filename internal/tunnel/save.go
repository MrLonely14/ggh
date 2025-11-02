package tunnel

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Create adds a new tunnel to storage
func Create(tunnel *Tunnel) error {
	if err := tunnel.Validate(); err != nil {
		return fmt.Errorf("invalid tunnel: %w", err)
	}

	tunnels, err := LoadTunnels()
	if err != nil {
		return err
	}

	// Check for duplicate name
	for _, t := range tunnels {
		if t.Name == tunnel.Name {
			return fmt.Errorf("tunnel with name '%s' already exists", tunnel.Name)
		}
	}

	// Generate ID if not set
	if tunnel.ID == "" {
		tunnel.ID = uuid.New().String()
	}

	// Set created timestamp
	tunnel.CreatedAt = time.Now().Format(time.RFC3339)

	tunnels = append(tunnels, *tunnel)

	return SaveTunnels(tunnels)
}

// Update modifies an existing tunnel
func Update(tunnel *Tunnel) error {
	if err := tunnel.Validate(); err != nil {
		return fmt.Errorf("invalid tunnel: %w", err)
	}

	tunnels, err := LoadTunnels()
	if err != nil {
		return err
	}

	found := false
	for i, t := range tunnels {
		if t.ID == tunnel.ID {
			// Preserve original creation time
			tunnel.CreatedAt = t.CreatedAt
			tunnels[i] = *tunnel
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("tunnel not found: %s", tunnel.ID)
	}

	return SaveTunnels(tunnels)
}

// Delete removes a tunnel from storage
func Delete(id string) error {
	tunnels, err := LoadTunnels()
	if err != nil {
		return err
	}

	newTunnels := make([]Tunnel, 0, len(tunnels))
	found := false

	for _, tunnel := range tunnels {
		if tunnel.ID == id {
			found = true
			continue
		}
		newTunnels = append(newTunnels, tunnel)
	}

	if !found {
		return fmt.Errorf("tunnel not found: %s", id)
	}

	return SaveTunnels(newTunnels)
}

// UpdateLastUsed updates the last used timestamp for a tunnel
func UpdateLastUsed(id string) error {
	tunnels, err := LoadTunnels()
	if err != nil {
		return err
	}

	found := false
	for i, tunnel := range tunnels {
		if tunnel.ID == id {
			tunnels[i].LastUsed = time.Now().Format(time.RFC3339)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("tunnel not found: %s", id)
	}

	return SaveTunnels(tunnels)
}

// UpdateLastUsedBatch updates the last used timestamp for multiple tunnels
func UpdateLastUsedBatch(ids []string) error {
	tunnels, err := LoadTunnels()
	if err != nil {
		return err
	}

	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	now := time.Now().Format(time.RFC3339)
	for i, tunnel := range tunnels {
		if idMap[tunnel.ID] {
			tunnels[i].LastUsed = now
		}
	}

	return SaveTunnels(tunnels)
}
