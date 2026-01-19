package vault

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vcv/config"
	"vcv/internal/certs"
)

type fakeSizerClient struct {
	MockClient
	cacheSize int
}

func (c *fakeSizerClient) CacheSize() int {
	return c.cacheSize
}

func TestDisabledClient(t *testing.T) {
	client := &disabledClient{}
	err := client.CheckConnection(context.Background())
	assert.ErrorIs(t, err, ErrVaultNotConfigured)
	_, err = client.GetCertificateDetails(context.Background(), "id")
	assert.ErrorIs(t, err, ErrVaultNotConfigured)
	_, err = client.GetCertificatePEM(context.Background(), "id")
	assert.ErrorIs(t, err, ErrVaultNotConfigured)
	certsList, err := client.ListCertificates(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []certs.Certificate{}, certsList)
	client.InvalidateCache()
	client.Shutdown()
}

func TestParseCompositeCertificateID(t *testing.T) {
	tests := []struct {
		name         string
		orderedVault []string
		value        string
		expectErr    bool
		expectVault  string
		expectMount  string
	}{
		{name: "explicit", orderedVault: []string{"v1"}, value: "v2|pki:aa", expectErr: false, expectVault: "v2", expectMount: "pki:aa"},
		{name: "explicit invalid", orderedVault: []string{"v1"}, value: "v2|", expectErr: true},
		{name: "implicit uses first", orderedVault: []string{"v1"}, value: "pki:aa", expectErr: false, expectVault: "v1", expectMount: "pki:aa"},
		{name: "implicit empty list", orderedVault: []string{}, value: "pki:aa", expectErr: true},
		{name: "empty value", orderedVault: []string{"v1"}, value: " ", expectErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			vaultID, mountSerial, err := parseCompositeCertificateID(tt.orderedVault, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectVault, vaultID)
			assert.Equal(t, tt.expectMount, mountSerial)
		})
	}
}

func TestMultiClient_CheckConnection(t *testing.T) {
	t.Run("not configured", func(t *testing.T) {
		m := NewMultiClient([]config.VaultInstance{}, map[string]Client{})
		err := m.CheckConnection(context.Background())
		assert.ErrorIs(t, err, ErrVaultNotConfigured)
	})
	t.Run("missing client", func(t *testing.T) {
		instances := []config.VaultInstance{{ID: "v1"}}
		m := NewMultiClient(instances, map[string]Client{"v1": nil})
		err := m.CheckConnection(context.Background())
		assert.Error(t, err)
	})
	t.Run("delegates", func(t *testing.T) {
		instances := []config.VaultInstance{{ID: "v1"}}
		c1 := &MockClient{}
		c1.On("CheckConnection", mock.Anything).Return(nil)
		m := NewMultiClient(instances, map[string]Client{"v1": c1})
		err := m.CheckConnection(context.Background())
		assert.NoError(t, err)
		c1.AssertExpectations(t)
	})
}

func TestMultiClient_GetCertificateDetails_RoutesByExplicitVault(t *testing.T) {
	instances := []config.VaultInstance{{ID: "v1"}, {ID: "v2"}}
	c1 := &MockClient{}
	c2 := &MockClient{}
	expected := certs.DetailedCertificate{Certificate: certs.Certificate{ID: "pki:aa", CommonName: "cn"}, SerialNumber: "aa"}
	c2.On("GetCertificateDetails", mock.Anything, "pki:aa").Return(expected, nil)
	m := NewMultiClient(instances, map[string]Client{"v1": c1, "v2": c2})
	result, err := m.GetCertificateDetails(context.Background(), "v2|pki:aa")
	assert.NoError(t, err)
	assert.Equal(t, "v2|pki:aa", result.ID)
	assert.Equal(t, "aa", result.SerialNumber)
	c2.AssertExpectations(t)
}

func TestMultiClient_GetCertificateDetails_MissingClient(t *testing.T) {
	instances := []config.VaultInstance{{ID: "v1"}}
	m := NewMultiClient(instances, map[string]Client{})
	_, err := m.GetCertificateDetails(context.Background(), "v1|pki:aa")
	assert.Error(t, err)
}

func TestMultiClient_GetCertificatePEM(t *testing.T) {
	instances := []config.VaultInstance{{ID: "v1"}}
	c1 := &MockClient{}
	c1.On("GetCertificatePEM", mock.Anything, "pki:aa").Return(certs.PEMResponse{SerialNumber: "aa", PEM: "pem"}, nil)
	m := NewMultiClient(instances, map[string]Client{"v1": c1})
	result, err := m.GetCertificatePEM(context.Background(), "v1|pki:aa")
	assert.NoError(t, err)
	assert.Equal(t, "aa", result.SerialNumber)
	c1.AssertExpectations(t)
}

func TestMultiClient_ListCertificates_SortsAndPrefixes(t *testing.T) {
	instances := []config.VaultInstance{{ID: "v1"}, {ID: "v2"}}
	c1 := &MockClient{}
	c2 := &MockClient{}
	c1.On("ListCertificates", mock.Anything).Return([]certs.Certificate{{ID: "pki:b", CommonName: "beta"}}, nil)
	c2.On("ListCertificates", mock.Anything).Return([]certs.Certificate{{ID: "pki:a", CommonName: "alpha"}}, nil)
	m := NewMultiClient(instances, map[string]Client{"v1": c1, "v2": c2})
	result, err := m.ListCertificates(context.Background())
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "alpha", result[0].CommonName)
	assert.Equal(t, "v2|pki:a", result[0].ID)
	assert.Equal(t, "beta", result[1].CommonName)
	assert.Equal(t, "v1|pki:b", result[1].ID)
	c1.AssertExpectations(t)
	c2.AssertExpectations(t)
}

func TestMultiClient_ListCertificates_AllFailed(t *testing.T) {
	var instances []config.VaultInstance = []config.VaultInstance{{ID: "v1"}}
	var mockClient *MockClient = &MockClient{}
	var errBoom error = errors.New("boom")
	mockClient.On("ListCertificates", mock.Anything).Return([]certs.Certificate{}, errBoom)
	var client Client = NewMultiClient(instances, map[string]Client{"v1": mockClient})
	var err error
	_, err = client.ListCertificates(context.Background())
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestMultiClient_ListCertificatesByVault_ReturnsErrorsAndDurations(t *testing.T) {
	instances := []config.VaultInstance{{ID: "v1"}, {ID: "v2"}}
	c1 := &MockClient{}
	errBoom := errors.New("boom")
	c1.On("ListCertificates", mock.Anything).Return([]certs.Certificate{}, errBoom)
	m := NewMultiClient(instances, map[string]Client{"v1": c1, "v2": nil})
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	results := m.(CertificatesByVaultLister).ListCertificatesByVault(ctx)
	assert.Len(t, results, 2)
	assert.Equal(t, "v1", results[0].VaultID)
	assert.ErrorIs(t, results[0].ListError, errBoom)
	assert.GreaterOrEqual(t, results[0].Duration, time.Duration(0))
	assert.Equal(t, "v2", results[1].VaultID)
	assert.Error(t, results[1].ListError)
	c1.AssertExpectations(t)
}

func TestMultiClient_CacheSize_AggregatesUnique(t *testing.T) {
	instances := []config.VaultInstance{{ID: "v1"}, {ID: "v2"}, {ID: "v3"}}
	c1 := &fakeSizerClient{cacheSize: 2}
	c2 := &fakeSizerClient{cacheSize: 3}
	clients := map[string]Client{"v1": c1, "v2": c1, "v3": c2}
	m := NewMultiClient(instances, clients)
	size := m.(CacheSizer).CacheSize()
	assert.Equal(t, 5, size)
}

func TestMultiClient_InvalidateCache_And_Shutdown_Unique(t *testing.T) {
	instances := []config.VaultInstance{{ID: "v1"}, {ID: "v2"}}
	c1 := &MockClient{}
	c2 := &MockClient{}
	c1.On("InvalidateCache").Return()
	c2.On("InvalidateCache").Return()
	c1.On("Shutdown").Return()
	c2.On("Shutdown").Return()
	m := NewMultiClient(instances, map[string]Client{"v1": c1, "v2": c2})
	m.InvalidateCache()
	m.Shutdown()
	c1.AssertExpectations(t)
	c2.AssertExpectations(t)
}
