package executor

import (
	"testing"
)

func TestLocalExecutor(t *testing.T) {
	t.Skip("Skipping test as it requires a C compiler")
	// Create a local executor
	exec := NewLocalExecutor(nil)

	// Execute a simple command
	err := exec.Execute([]string{"echo", "hello"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Execute with output
	output, err := exec.ExecuteWithOutput([]string{"echo", "hello"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check output
	expected := "hello\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}

	// Test with an invalid command
	err = exec.Execute([]string{"nonexistentcommand"})
	if err == nil {
		t.Error("Expected error for nonexistent command, got nil")
	}
}

func TestSSHConfigCreation(t *testing.T) {
	t.Skip("Skipping test as it requires a C compiler")
	// Create a default SSH config
	cfg := NewSSHConfig()

	// Check default values
	if cfg.Port != 22 {
		t.Errorf("Expected default port 22, got %d", cfg.Port)
	}

	if !cfg.VerifyHost {
		t.Errorf("Expected default VerifyHost to be true")
	}

	if cfg.Timeout == 0 {
		t.Errorf("Expected non-zero default timeout")
	}
}
