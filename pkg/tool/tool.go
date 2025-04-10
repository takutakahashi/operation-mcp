package tool

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/takutakahashi/operation-mcp/pkg/config"
	"github.com/takutakahashi/operation-mcp/pkg/danger"
)

// Manager handles tool execution
type Manager struct {
	config        *config.Config
	dangerManager *danger.Manager
}

// NewManager creates a new tool manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config:        cfg,
		dangerManager: danger.NewManager(cfg.Actions),
	}
}

// FindTool finds a tool by its name
func (m *Manager) FindTool(toolPath string) ([]string, map[string]config.Parameter, string, error) {
	parts := strings.Split(toolPath, "_")
	if len(parts) < 1 {
		return nil, nil, "", fmt.Errorf("invalid tool path: %s", toolPath)
	}

	// Find the root tool
	var rootTool *config.Tool
	for i := range m.config.Tools {
		if m.config.Tools[i].Name == parts[0] {
			rootTool = &m.config.Tools[i]
			break
		}
	}

	if rootTool == nil {
		return nil, nil, "", fmt.Errorf("tool not found: %s", parts[0])
	}

	// Start with the root tool's command
	command := make([]string, len(rootTool.Command))
	copy(command, rootTool.Command)

	// Collect all parameters
	params := make(map[string]config.Parameter)
	for name, param := range rootTool.Params {
		params[name] = param
	}

	// If we only have the root tool, return it
	if len(parts) == 1 {
		return command, params, "", nil
	}

	// Navigate through subtools
	var currentSubtool *config.Subtool
	dangerLevel := ""

	// Join the remaining parts to form the subtool path
	subtoolPath := strings.Join(parts[1:], "_")

	// Find the matching subtool
	found := false
	for j := range rootTool.Subtools {
		subtoolName := strings.ReplaceAll(rootTool.Subtools[j].Name, " ", "_")
		if subtoolName == subtoolPath {
			currentSubtool = &rootTool.Subtools[j]
			found = true
			break
		}
	}

	if !found {
		return nil, nil, "", fmt.Errorf("subtool not found: %s", toolPath)
	}

	// Add subtool parameters
	for name, param := range currentSubtool.Params {
		params[name] = param
	}

	// Update danger level if specified
	if currentSubtool.DangerLevel != "" {
		dangerLevel = currentSubtool.DangerLevel
	}

	// Add the args from the final subtool
	if currentSubtool != nil {
		command = append(command, currentSubtool.Args...)
	}

	return command, params, dangerLevel, nil
}

// ExecuteTool executes a tool with the given parameters
func (m *Manager) ExecuteTool(toolPath string, paramValues map[string]string) error {
	// Find the tool
	command, params, dangerLevel, err := m.FindTool(toolPath)
	if err != nil {
		return err
	}

	// Validate required parameters
	for name, param := range params {
		if param.Required {
			value, exists := paramValues[name]
			if !exists || value == "" {
				return fmt.Errorf("required parameter missing: %s", name)
			}
		}
	}

	// Check danger level for parameters with validation rules
	for name, param := range params {
		value, exists := paramValues[name]
		if exists && len(param.Validate) > 0 {
			for _, validation := range param.Validate {
				proceed, err := m.dangerManager.CheckDangerLevel(
					validation.DangerLevel,
					name,
					value,
					param.Validate,
				)
				if err != nil {
					return err
				}
				if !proceed {
					return fmt.Errorf("operation aborted due to danger level check")
				}
			}
		}
	}

	// Check danger level for the tool itself
	if dangerLevel != "" {
		proceed, err := m.dangerManager.CheckDangerLevel(dangerLevel, "", "", nil)
		if err != nil {
			return err
		}
		if !proceed {
			return fmt.Errorf("operation aborted due to danger level check")
		}
	}

	// Replace template parameters in command args
	finalCommand := make([]string, len(command))
	for i, arg := range command {
		if strings.Contains(arg, "{{") {
			tmpl, err := template.New("arg").Parse(arg)
			if err != nil {
				return fmt.Errorf("error parsing template in argument: %w", err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, paramValues); err != nil {
				return fmt.Errorf("error executing template in argument: %w", err)
			}

			finalCommand[i] = buf.String()
		} else {
			finalCommand[i] = arg
		}
	}

	// Execute the command
	fmt.Printf("Executing: %s\n", strings.Join(finalCommand, " "))
	cmd := exec.Command(finalCommand[0], finalCommand[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
