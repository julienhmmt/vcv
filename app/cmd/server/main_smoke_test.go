package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const smokeStartupTimeout = 5 * time.Second

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

// waitForHealth polls /api/health until it answers 200 or the timeout elapses.
func waitForHealth(t *testing.T, port int, timeout time.Duration) {
	t.Helper()
	client := &http.Client{Timeout: 500 * time.Millisecond}
	url := fmt.Sprintf("http://127.0.0.1:%d/api/health", port)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server did not become healthy within %s", timeout)
}

func TestRun_ShutsDownOnContextCancel(t *testing.T) {
	port := freePort(t)
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")
	settingsContent := fmt.Sprintf(`{
		"app": {
			"env": "dev",
			"port": %d,
			"logging": {
				"level": "info",
				"format": "json",
				"output": "stdout"
			}
		},
		"vaults": []
	}`, port)

	err := os.WriteFile(settingsFile, []byte(settingsContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test settings file: %v", err)
	}

	// Change to temp directory to ensure the settings file is found
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- run(ctx)
	}()

	waitForHealth(t, port, smokeStartupTimeout)
	cancel()

	select {
	case runErr := <-done:
		if runErr != nil {
			t.Fatalf("run returned error: %v", runErr)
		}
	case <-time.After(smokeStartupTimeout):
		t.Fatalf("run did not shut down in time")
	}
}
