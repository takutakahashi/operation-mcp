package tool

import (
	"os"
	"testing"

	"github.com/takutakahashi/operation-mcp/pkg/config"
)

func TestFindTool(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Tools: []config.Tool{
			{
				Name:    "kubectl",
				Command: []string{"kubectl"},
				Params: map[string]config.Parameter{
					"namespace": {
						Description: "The namespace to run the command in",
						Type:        "string",
						Required:    true,
					},
				},
				Subtools: []config.Subtool{
					{
						Name: "get pod",
						Args: []string{"get", "pod", "-o", "json", "-n", "{{.namespace}}"},
					},
					{
						Name: "describe pod",
						Params: map[string]config.Parameter{
							"pod": {
								Description: "The pod to describe",
								Type:        "string",
								Required:    true,
							},
						},
						Args: []string{"describe", "pod", "{{.pod}}", "-n", "{{.namespace}}"},
					},
					{
						Name:        "delete pod",
						DangerLevel: "high",
						Params: map[string]config.Parameter{
							"pod": {
								Description: "The pod to delete",
								Type:        "string",
								Required:    true,
							},
						},
						Args: []string{"delete", "pod", "{{.pod}}", "-n", "{{.namespace}}"},
					},
				},
			},
		},
	}

	// Create a tool manager
	mgr := NewManager(cfg)

	// Test finding root tool
	command, params, dangerLevel, err := mgr.FindTool("kubectl")
	if err != nil {
		t.Fatalf("FindTool failed for root tool: %v", err)
	}
	if len(command) != 1 || command[0] != "kubectl" {
		t.Errorf("Expected command ['kubectl'], got %v", command)
	}
	if len(params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(params))
	}
	if dangerLevel != "" {
		t.Errorf("Expected empty danger level, got '%s'", dangerLevel)
	}

	// Test finding subtool
	command, params, dangerLevel, err = mgr.FindTool("kubectl_get_pod")
	if err != nil {
		t.Fatalf("FindTool failed for subtool: %v", err)
	}
	if len(command) != 7 {
		t.Errorf("Expected 7 command parts, got %d", len(command))
	}
	if command[0] != "kubectl" || command[1] != "get" || command[2] != "pod" {
		t.Errorf("Expected command starting with ['kubectl', 'get', 'pod'], got %v", command)
	}
	if len(params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(params))
	}
	if dangerLevel != "" {
		t.Errorf("Expected empty danger level, got '%s'", dangerLevel)
	}

	// Test finding subtool with danger level
	command, params, dangerLevel, err = mgr.FindTool("kubectl_delete_pod")
	if err != nil {
		t.Fatalf("FindTool failed for subtool with danger level: %v", err)
	}
	if len(command) != 6 {
		t.Errorf("Expected 6 command parts, got %d", len(command))
	}
	if command[0] != "kubectl" || command[1] != "delete" || command[2] != "pod" {
		t.Errorf("Expected command starting with ['kubectl', 'delete', 'pod'], got %v", command)
	}
	if len(params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(params))
	}
	if dangerLevel != "high" {
		t.Errorf("Expected danger level 'high', got '%s'", dangerLevel)
	}

	// Test finding non-existent tool
	_, _, _, err = mgr.FindTool("nonexistent")
	if err == nil {
		t.Errorf("FindTool should fail for non-existent tool")
	}

	// Test finding non-existent subtool
	_, _, _, err = mgr.FindTool("kubectl_nonexistent")
	if err == nil {
		t.Errorf("FindTool should fail for non-existent subtool")
	}
}

func TestExecuteRawTool(t *testing.T) {
	// Skip test if running in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Create a test config
	cfg := &config.Config{
		Tools: []config.Tool{
			{
				Name:    "echo",
				Command: []string{"echo"},
			},
			{
				Name:    "ls",
				Command: []string{"ls"},
			},
		},
	}

	// Create a tool manager
	mgr := NewManager(cfg)

	// Test executing a simple command
	err := mgr.ExecuteRawTool("echo", []string{"hello", "world"})
	if err != nil {
		t.Fatalf("ExecuteRawTool failed for echo: %v", err)
	}

	// Test executing with no arguments
	err = mgr.ExecuteRawTool("ls", []string{})
	if err != nil {
		t.Fatalf("ExecuteRawTool failed for ls with no args: %v", err)
	}

	// Test executing with arguments
	err = mgr.ExecuteRawTool("ls", []string{"-la"})
	if err != nil {
		t.Fatalf("ExecuteRawTool failed for ls with args: %v", err)
	}

	// Test executing non-existent tool
	err = mgr.ExecuteRawTool("nonexistent", []string{})
	if err == nil {
		t.Errorf("ExecuteRawTool should fail for non-existent tool")
	}

	// Test executing a command that will fail
	err = mgr.ExecuteRawTool("ls", []string{"/nonexistent/path"})
	if err == nil {
		t.Errorf("ExecuteRawTool should fail when command fails")
	}
}
