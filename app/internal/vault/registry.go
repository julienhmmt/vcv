package vault

import (
	"sync"

	"vcv/config"
)

// Registry tracks which vault IDs are currently enabled.
// It is safe for concurrent use and allows the admin panel
// to toggle vaults without restarting the application.
type Registry struct {
	mu         sync.RWMutex
	enabledIDs map[string]struct{}
}

// NewRegistry creates a Registry pre-populated with the enabled vault IDs.
func NewRegistry(instances []config.VaultInstance) *Registry {
	r := &Registry{enabledIDs: make(map[string]struct{})}
	r.Update(instances)
	return r
}

// Update replaces the enabled set with the IDs of currently enabled instances.
func (r *Registry) Update(instances []config.VaultInstance) {
	enabled := make(map[string]struct{}, len(instances))
	for _, inst := range instances {
		if config.IsVaultEnabled(inst) {
			enabled[inst.ID] = struct{}{}
		}
	}
	r.mu.Lock()
	r.enabledIDs = enabled
	r.mu.Unlock()
}

// IsEnabled reports whether the given vault ID is currently enabled.
func (r *Registry) IsEnabled(vaultID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.enabledIDs[vaultID]
	return ok
}

// EnabledIDs returns a snapshot of the currently enabled vault IDs.
func (r *Registry) EnabledIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.enabledIDs))
	for id := range r.enabledIDs {
		ids = append(ids, id)
	}
	return ids
}
