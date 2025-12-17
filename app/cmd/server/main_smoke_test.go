package main

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestMain_ShutsDownOnSignal(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("PORT", "0")
	t.Setenv("VCV_ADMIN_PASSWORD", "")
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_READ_TOKEN", "")
	t.Setenv("VAULT_ADDRS", "")
	t.Setenv("LOG_OUTPUT", "stdout")
	t.Setenv("LOG_FORMAT", "json")
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
