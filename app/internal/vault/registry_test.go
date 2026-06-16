package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"vcv/internal/config"
)

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name      string
		instances []config.VaultInstance
		expected  []string
	}{
		{
			name:      "empty instances",
			instances: []config.VaultInstance{},
			expected:  []string{},
		},
		{
			name: "single enabled instance",
			instances: []config.VaultInstance{
				{ID: "v1", Enabled: boolPtr(true)},
			},
			expected: []string{"v1"},
		},
		{
			name: "multiple enabled instances",
			instances: []config.VaultInstance{
				{ID: "v1", Enabled: boolPtr(true)},
				{ID: "v2", Enabled: boolPtr(true)},
			},
			expected: []string{"v1", "v2"},
		},
		{
			name: "disabled instances are excluded",
			instances: []config.VaultInstance{
				{ID: "v1", Enabled: boolPtr(true)},
				{ID: "v2", Enabled: boolPtr(false)},
				{ID: "v3", Enabled: boolPtr(true)},
			},
			expected: []string{"v1", "v3"},
		},
		{
			name: "nil enabled defaults to enabled",
			instances: []config.VaultInstance{
				{ID: "v1"},
			},
			expected: []string{"v1"},
		},
		{
			name: "empty id is included if enabled",
			instances: []config.VaultInstance{
				{ID: "", Enabled: boolPtr(true)},
				{ID: "v1", Enabled: boolPtr(true)},
			},
			expected: []string{"", "v1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry(tt.instances)
			assert.NotNil(t, r)
			ids := r.EnabledIDs()
			assert.ElementsMatch(t, tt.expected, ids)
		})
	}
}

func TestRegistry_Update(t *testing.T) {
	r := NewRegistry([]config.VaultInstance{
		{ID: "v1", Enabled: boolPtr(true)},
		{ID: "v2", Enabled: boolPtr(true)},
	})

	assert.True(t, r.IsEnabled("v1"))
	assert.True(t, r.IsEnabled("v2"))
	assert.False(t, r.IsEnabled("v3"))

	// Update to disable v2 and add v3
	r.Update([]config.VaultInstance{
		{ID: "v1", Enabled: boolPtr(true)},
		{ID: "v2", Enabled: boolPtr(false)},
		{ID: "v3", Enabled: boolPtr(true)},
	})

	assert.True(t, r.IsEnabled("v1"))
	assert.False(t, r.IsEnabled("v2"))
	assert.True(t, r.IsEnabled("v3"))

	// Update to empty
	r.Update([]config.VaultInstance{})
	assert.False(t, r.IsEnabled("v1"))
	assert.Len(t, r.EnabledIDs(), 0)
}

func TestRegistry_IsEnabled(t *testing.T) {
	r := NewRegistry([]config.VaultInstance{
		{ID: "v1", Enabled: boolPtr(true)},
		{ID: "v2", Enabled: boolPtr(false)},
	})

	assert.True(t, r.IsEnabled("v1"))
	assert.False(t, r.IsEnabled("v2"))
	assert.False(t, r.IsEnabled("nonexistent"))
	assert.False(t, r.IsEnabled(""))
}

func TestRegistry_EnabledIDs(t *testing.T) {
	r := NewRegistry([]config.VaultInstance{
		{ID: "v1", Enabled: boolPtr(true)},
		{ID: "v2", Enabled: boolPtr(true)},
		{ID: "v3", Enabled: boolPtr(false)},
	})

	ids := r.EnabledIDs()
	assert.Len(t, ids, 2)
	assert.ElementsMatch(t, []string{"v1", "v2"}, ids)

	// Should return a copy, not the internal slice
	ids = append(ids, "tampered")
	assert.Len(t, r.EnabledIDs(), 2)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := NewRegistry([]config.VaultInstance{
		{ID: "v1", Enabled: boolPtr(true)},
	})

	// This should not panic due to race conditions
	for i := 0; i < 100; i++ {
		go r.IsEnabled("v1")
		go r.EnabledIDs()
		go r.Update([]config.VaultInstance{{ID: "v1", Enabled: boolPtr(true)}})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
