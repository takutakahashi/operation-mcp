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

// ToolInfo represents a tool or subtool for hierarchical display
type ToolInfo struct {
	Name        string
	Description string
	Params      map[string]config.Parameter
	Subtools    []ToolInfo
}

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

// ExecuteRawTool executes a tool with the given raw arguments
func (m *Manager) ExecuteRawTool(toolPath string, args []string) error {
	// Find the tool and subtool
	command, params, dangerLevel, err := m.FindTool(toolPath)
	if err != nil {
		return err
	}

	// Extract parameter values from the command-line arguments
	paramValues := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			paramName := strings.TrimLeft(arg, "-")
			// Handle --param=value format
			if strings.Contains(paramName, "=") {
				parts := strings.SplitN(paramName, "=", 2)
				paramName = parts[0]
				paramValues[paramName] = parts[1]
				continue
			}

			// Handle -p value format
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				paramValues[paramName] = args[i+1]
				i++ // Skip the next arg since it's the value
			} else {
				// Handle boolean flags like -f
				paramValues[paramName] = "true"
			}
		}
	}

	// Check danger level for the subtool
	if dangerLevel != "" {
		proceed, err := m.dangerManager.CheckDangerLevel(dangerLevel, "", "", nil)
		if err != nil {
			return err
		}
		if !proceed {
			return fmt.Errorf("operation aborted due to danger level check")
		}
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

// ListTools returns all tools and subtools defined in the config
func (m *Manager) ListTools() []ToolInfo {
	if m.config == nil || len(m.config.Tools) == 0 {
		return []ToolInfo{}
	}
	
	result := make([]ToolInfo, 0, len(m.config.Tools))
	
	for _, tool := range m.config.Tools {
		toolInfo := ToolInfo{
			Name:        tool.Name,
			Description: "", // Config doesn't have description field for tools
			Params:      tool.Params,
			Subtools:    make([]ToolInfo, 0, len(tool.Subtools)),
		}
		
		// Add subtools recursively
		for _, subtool := range tool.Subtools {
			toolInfo.Subtools = append(toolInfo.Subtools, convertSubtoolToToolInfo(subtool, tool.Name))
		}
		
		result = append(result, toolInfo)
	}
	
	return result
}

// convertSubtoolToToolInfo converts a subtool configuration to ToolInfo structure
func convertSubtoolToToolInfo(subtool config.Subtool, parentName string) ToolInfo {
	name := strings.ReplaceAll(subtool.Name, " ", "_")
	
	toolInfo := ToolInfo{
		Name:        name,
		Description: "", // Config doesn't have description field for subtools
		Params:      subtool.Params,
		Subtools:    make([]ToolInfo, 0, len(subtool.Subtools)),
	}
	
	// Add nested subtools recursively
	for _, nested := range subtool.Subtools {
		toolInfo.Subtools = append(toolInfo.Subtools, 
			convertSubtoolToToolInfo(nested, parentName+"_"+name))
	}
	
	return toolInfo
}