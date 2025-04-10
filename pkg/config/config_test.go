package config

import (
        "os"
        "path/filepath"
        "testing"
)

func TestLoadConfig(t *testing.T) {
        // Create a temporary config file
        tempDir := t.TempDir()
        configPath := filepath.Join(tempDir, "config.yaml")

        configContent := `
actions:
  - danger_level: high
    type: confirm
    message: "This is a high danger operation. Proceed?"
  - danger_level: medium
    type: timeout
    message: "This is a medium danger operation. Proceeding in 5 seconds."
    timeout: 5
  - danger_level: low
    type: force
    message: "This is a low danger operation."

tools:
  - name: kubectl
    command:
      - kubectl
    params:
      namespace:
        description: The namespace to run the command in
        type: string
        required: true
        validate:
          - danger_level: high
            exclude:
              - kube-system
              - kube-public
    subtools:
      - name: get pod
        args: ["get", "pod", "-o", "json", "-n", "{{.namespace}}"]
      - name: describe pod
        params:
          pod:
            description: The pod to describe
            type: string
            required: true
        args: ["describe", "pod", "{{.pod}}", "-n", "{{.namespace}}"]
`

        if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
                t.Fatalf("Failed to write test config file: %v", err)
        }

        // Test loading the config
        cfg, err := LoadConfig(configPath)
        if err != nil {
                t.Fatalf("LoadConfig failed: %v", err)
        }

        // Verify actions
        if len(cfg.Actions) != 3 {
                t.Errorf("Expected 3 actions, got %d", len(cfg.Actions))
        }

        // Verify tools
        if len(cfg.Tools) != 1 {
                t.Errorf("Expected 1 tool, got %d", len(cfg.Tools))
        }

        // Verify tool name
        if cfg.Tools[0].Name != "kubectl" {
                t.Errorf("Expected tool name 'kubectl', got '%s'", cfg.Tools[0].Name)
        }

        // Verify tool command
        if len(cfg.Tools[0].Command) != 1 || cfg.Tools[0].Command[0] != "kubectl" {
                t.Errorf("Expected tool command ['kubectl'], got %v", cfg.Tools[0].Command)
        }

        // Verify tool parameters
        if len(cfg.Tools[0].Params) != 1 {
                t.Errorf("Expected 1 parameter, got %d", len(cfg.Tools[0].Params))
        }

        // Verify parameter details
        param, exists := cfg.Tools[0].Params["namespace"]
        if !exists {
                t.Errorf("Expected parameter 'namespace' not found")
        } else {
                if param.Type != "string" {
                        t.Errorf("Expected parameter type 'string', got '%s'", param.Type)
                }
                if !param.Required {
                        t.Errorf("Expected parameter to be required")
                }
                if len(param.Validate) != 1 {
                        t.Errorf("Expected 1 validation rule, got %d", len(param.Validate))
                }
        }

        // Verify subtools
        if len(cfg.Tools[0].Subtools) != 2 {
                t.Errorf("Expected 2 subtools, got %d", len(cfg.Tools[0].Subtools))
        }

        // Verify subtool name
        if cfg.Tools[0].Subtools[0].Name != "get pod" {
                t.Errorf("Expected subtool name 'get pod', got '%s'", cfg.Tools[0].Subtools[0].Name)
        }

        // Verify subtool args
        if len(cfg.Tools[0].Subtools[0].Args) != 6 {
                t.Errorf("Expected 6 args, got %d", len(cfg.Tools[0].Subtools[0].Args))
        }
}

func TestConfigValidate(t *testing.T) {
        // Test valid config
        validConfig := &Config{
                Actions: []Action{
                        {
                                DangerLevel: "high",
                                Type:        "confirm",
                                Message:     "This is a high danger operation. Proceed?",
                        },
                },
                Tools: []Tool{
                        {
                                Name:    "kubectl",
                                Command: []string{"kubectl"},
                                Params: map[string]Parameter{
                                        "namespace": {
                                                Description: "The namespace to run the command in",
                                                Type:        "string",
                                                Required:    true,
                                        },
                                },
                                Subtools: []Subtool{
                                        {
                                                Name: "get pod",
                                                Args: []string{"get", "pod", "-o", "json", "-n", "{{.namespace}}"},
                                        },
                                },
                        },
                },
        }

        if err := validConfig.Validate(); err != nil {
                t.Errorf("Validation failed for valid config: %v", err)
        }

        // Test invalid config - missing action type
        invalidConfig1 := &Config{
                Actions: []Action{
                        {
                                DangerLevel: "high",
                                // Missing Type
                                Message: "This is a high danger operation. Proceed?",
                        },
                },
                Tools: []Tool{
                        {
                                Name:    "kubectl",
                                Command: []string{"kubectl"},
                        },
                },
        }

        if err := invalidConfig1.Validate(); err == nil {
                t.Errorf("Validation should fail for config with missing action type")
        }

        // Test invalid config - invalid action type
        invalidConfig2 := &Config{
                Actions: []Action{
                        {
                                DangerLevel: "high",
                                Type:        "invalid",
                                Message:     "This is a high danger operation. Proceed?",
                        },
                },
                Tools: []Tool{
                        {
                                Name:    "kubectl",
                                Command: []string{"kubectl"},
                        },
                },
        }

        if err := invalidConfig2.Validate(); err == nil {
                t.Errorf("Validation should fail for config with invalid action type")
        }

        // Test invalid config - missing tool name
        invalidConfig3 := &Config{
                Actions: []Action{
                        {
                                DangerLevel: "high",
                                Type:        "confirm",
                                Message:     "This is a high danger operation. Proceed?",
                        },
                },
                Tools: []Tool{
                        {
                                // Missing Name
                                Command: []string{"kubectl"},
                        },
                },
        }

        if err := invalidConfig3.Validate(); err == nil {
                t.Errorf("Validation should fail for config with missing tool name")
        }

        // Test invalid config - missing command
        invalidConfig4 := &Config{
                Actions: []Action{
                        {
                                DangerLevel: "high",
                                Type:        "confirm",
                                Message:     "This is a high danger operation. Proceed?",
                        },
                },
                Tools: []Tool{
                        {
                                Name: "kubectl",
                                // Missing Command
                        },
                },
        }

        if err := invalidConfig4.Validate(); err == nil {
                t.Errorf("Validation should fail for config with missing command")
        }
}