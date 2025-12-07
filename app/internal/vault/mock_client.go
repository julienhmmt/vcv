package vault

import (
	"context"

	"github.com/stretchr/testify/mock"

	"vcv/internal/certs"
)

// MockClient is a testify mock implementing Client.
type MockClient struct {
	mock.Mock
}

func (m *MockClient) CheckConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) GetCRL(ctx context.Context) ([]byte, error) {
	args := m.Called(ctx)
	if data, ok := args.Get(0).([]byte); ok {
		return data, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockClient) GetCertificateDetails(ctx context.Context, serialNumber string) (certs.DetailedCertificate, error) {
	args := m.Called(ctx, serialNumber)
	return args.Get(0).(certs.DetailedCertificate), args.Error(1)
}

func (m *MockClient) GetCertificatePEM(ctx context.Context, serialNumber string) (certs.PEMResponse, error) {
	args := m.Called(ctx, serialNumber)
	return args.Get(0).(certs.PEMResponse), args.Error(1)
}

func (m *MockClient) InvalidateCache() {
	m.Called()
}

func (m *MockClient) ListCertificates(ctx context.Context) ([]certs.Certificate, error) {
	args := m.Called(ctx)
	if list, ok := args.Get(0).([]certs.Certificate); ok {
		return list, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockClient) RotateCRL(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) Shutdown() {
	m.Called()
}
