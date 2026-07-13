package vault

import (
	"context"
	"sync"
	"time"
)

// InstanceStatus is the connection status for a single vault instance.
type InstanceStatus struct {
	ID        string
	Connected bool
	Error     error
}

// CheckInstances checks vault clients in parallel with a per-instance timeout.
// ordered defines the response order; clients may be missing (treated as disconnected).
func CheckInstances(ctx context.Context, ordered []string, clients map[string]Client, timeout time.Duration) []InstanceStatus {
	results := make([]InstanceStatus, len(ordered))
	var wg sync.WaitGroup
	for i, id := range ordered {
		results[i] = InstanceStatus{ID: id, Connected: false}
		if clients == nil {
			continue
		}
		client, ok := clients[id]
		if !ok || client == nil {
			continue
		}
		wg.Add(1)
		go func(idx int, client Client) {
			defer wg.Done()
			checkCtx := ctx
			var cancel context.CancelFunc
			if timeout > 0 {
				checkCtx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
			err := client.CheckConnection(checkCtx)
			if err == nil {
				results[idx].Connected = true
				return
			}
			results[idx].Error = err
		}(i, client)
	}
	wg.Wait()
	return results
}
