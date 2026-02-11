package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

func TestMain_ShutsDownOnSignal(t *testing.T) {
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
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	done := make(chan struct{})
	go func() {
		main()
		close(done)
	}()
	time.Sleep(100 * time.Millisecond)
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}
	if signalErr := proc.Signal(syscall.SIGTERM); signalErr != nil {
		t.Fatalf("failed to send SIGTERM: %v", signalErr)
	}
	select {
	case <-done:
		return
	case <-time.After(2 * time.Second):
		t.Fatalf("main did not shut down in time")
	}
}
