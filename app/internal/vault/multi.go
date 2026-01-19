package vault

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"vcv/config"
	"vcv/internal/certs"
)

type multiClient struct {
	orderedVaultIDs []string
	clientsByVault  map[string]Client
}

func NewMultiClient(vaultInstances []config.VaultInstance, clientsByVault map[string]Client) Client {
	ordered := make([]string, 0, len(vaultInstances))
	seen := make(map[string]struct{}, len(vaultInstances))
	for _, instance := range vaultInstances {
		vaultID := strings.TrimSpace(instance.ID)
		if vaultID == "" {
			continue
		}
		if _, ok := clientsByVault[vaultID]; !ok {
			continue
		}
		if _, ok := seen[vaultID]; ok {
			continue
		}
		seen[vaultID] = struct{}{}
		ordered = append(ordered, vaultID)
	}
	if len(ordered) == 0 {
		for vaultID := range clientsByVault {
			if strings.TrimSpace(vaultID) == "" {
				continue
			}
			ordered = append(ordered, vaultID)
		}
		sort.Strings(ordered)
	}
	return &multiClient{orderedVaultIDs: ordered, clientsByVault: clientsByVault}
}

func (c *multiClient) CheckConnection(ctx context.Context) error {
	if len(c.orderedVaultIDs) == 0 {
		return ErrVaultNotConfigured
	}
	for _, vaultID := range c.orderedVaultIDs {
		client := c.clientsByVault[vaultID]
		if client == nil {
			return fmt.Errorf("missing vault client for %s", vaultID)
		}
		if err := client.CheckConnection(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *multiClient) GetCertificateDetails(ctx context.Context, serialNumber string) (certs.DetailedCertificate, error) {
	vaultID, mountSerial, err := parseCompositeCertificateID(c.orderedVaultIDs, serialNumber)
	if err != nil {
		return certs.DetailedCertificate{}, err
	}
	client := c.clientsByVault[vaultID]
	if client == nil {
		return certs.DetailedCertificate{}, fmt.Errorf("missing vault client for %s", vaultID)
	}
	details, err := client.GetCertificateDetails(ctx, mountSerial)
	if err != nil {
		return certs.DetailedCertificate{}, err
	}
	details.ID = fmt.Sprintf("%s|%s", vaultID, mountSerial)
	return details, nil
}

func (c *multiClient) GetCertificatePEM(ctx context.Context, serialNumber string) (certs.PEMResponse, error) {
	vaultID, mountSerial, err := parseCompositeCertificateID(c.orderedVaultIDs, serialNumber)
	if err != nil {
		return certs.PEMResponse{}, err
	}
	client := c.clientsByVault[vaultID]
	if client == nil {
		return certs.PEMResponse{}, fmt.Errorf("missing vault client for %s", vaultID)
	}
	return client.GetCertificatePEM(ctx, mountSerial)
}

func (c *multiClient) InvalidateCache() {
	unique := make(map[Client]struct{})
	for _, client := range c.clientsByVault {
		if client == nil {
			continue
		}
		unique[client] = struct{}{}
	}
	for client := range unique {
		client.InvalidateCache()
	}
}

func (c *multiClient) ListCertificates(ctx context.Context) ([]certs.Certificate, error) {
	if len(c.orderedVaultIDs) == 0 {
		return []certs.Certificate{}, ErrVaultNotConfigured
	}
	type result struct {
		vaultID      string
		certificates []certs.Certificate
		err          error
	}
	var resultChan chan result = make(chan result, len(c.orderedVaultIDs))
	var wg sync.WaitGroup
	for _, vaultID := range c.orderedVaultIDs {
		client := c.clientsByVault[vaultID]
		if client == nil {
			resultChan <- result{vaultID: vaultID, certificates: []certs.Certificate{}, err: fmt.Errorf("missing vault client for %s", vaultID)}
			continue
		}
		wg.Add(1)
		go func(id string, cl Client) {
			defer wg.Done()
			var certificates []certs.Certificate
			var err error
			certificates, err = cl.ListCertificates(ctx)
			if err != nil {
				resultChan <- result{vaultID: id, certificates: []certs.Certificate{}, err: err}
				return
			}
			resultChan <- result{vaultID: id, certificates: certificates, err: nil}
		}(vaultID, client)
	}
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	var all []certs.Certificate = make([]certs.Certificate, 0)
	var successCount int = 0
	var lastError error
	for res := range resultChan {
		if res.err != nil {
			lastError = res.err
			continue
		}
		successCount += 1
		for _, certificate := range res.certificates {
			prefixed := certificate
			prefixed.ID = fmt.Sprintf("%s|%s", res.vaultID, certificate.ID)
			all = append(all, prefixed)
		}
	}
	if successCount == 0 {
		if lastError != nil {
			return []certs.Certificate{}, lastError
		}
		return []certs.Certificate{}, ErrVaultNotConfigured
	}
	sort.Slice(all, func(leftIndex int, rightIndex int) bool {
		left := all[leftIndex]
		right := all[rightIndex]
		if left.CommonName != right.CommonName {
			return left.CommonName < right.CommonName
		}
		return left.ID < right.ID
	})
	return all, nil
}

func (c *multiClient) ListCertificatesByVault(ctx context.Context) []ListCertificatesByVaultResult {
	results := make([]ListCertificatesByVaultResult, 0, len(c.orderedVaultIDs))
	for _, vaultID := range c.orderedVaultIDs {
		client := c.clientsByVault[vaultID]
		if client == nil {
			results = append(results, ListCertificatesByVaultResult{VaultID: vaultID, Certificates: []certs.Certificate{}, Duration: 0, ListError: fmt.Errorf("missing vault client for %s", vaultID)})
			continue
		}
		start := time.Now()
		certificates, err := client.ListCertificates(ctx)
		duration := time.Since(start)
		if err != nil {
			results = append(results, ListCertificatesByVaultResult{VaultID: vaultID, Certificates: []certs.Certificate{}, Duration: duration, ListError: err})
			continue
		}
		prefixed := make([]certs.Certificate, 0, len(certificates))
		for _, certificate := range certificates {
			value := certificate
			value.ID = fmt.Sprintf("%s|%s", vaultID, certificate.ID)
			prefixed = append(prefixed, value)
		}
		results = append(results, ListCertificatesByVaultResult{VaultID: vaultID, Certificates: prefixed, Duration: duration, ListError: nil})
	}
	return results
}

func (c *multiClient) CacheSize() int {
	unique := make(map[Client]struct{})
	for _, client := range c.clientsByVault {
		if client == nil {
			continue
		}
		unique[client] = struct{}{}
	}
	total := 0
	for client := range unique {
		if sizer, ok := client.(CacheSizer); ok {
			total += sizer.CacheSize()
		}
	}
	return total
}

func (c *multiClient) Shutdown() {
	unique := make(map[Client]struct{})
	for _, client := range c.clientsByVault {
		if client == nil {
			continue
		}
		unique[client] = struct{}{}
	}
	for client := range unique {
		client.Shutdown()
	}
}

func parseCompositeCertificateID(orderedVaultIDs []string, value string) (string, string, error) {
	parts := strings.SplitN(value, "|", 2)
	if len(parts) == 2 {
		vaultID := strings.TrimSpace(parts[0])
		mountSerial := strings.TrimSpace(parts[1])
		if vaultID == "" || mountSerial == "" {
			return "", "", fmt.Errorf("invalid certificate id")
		}
		return vaultID, mountSerial, nil
	}
	if len(orderedVaultIDs) == 0 {
		return "", "", fmt.Errorf("invalid certificate id")
	}
	mountSerial := strings.TrimSpace(value)
	if mountSerial == "" {
		return "", "", fmt.Errorf("invalid certificate id")
	}
	return orderedVaultIDs[0], mountSerial, nil
}
