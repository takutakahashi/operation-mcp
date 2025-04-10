package danger

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/takutakahashi/operation-mcp/pkg/config"
)

// Manager handles danger level management
type Manager struct {
	actions map[string]config.Action
}

// NewManager creates a new danger manager
func NewManager(actions []config.Action) *Manager {
	actionMap := make(map[string]config.Action)
	for _, action := range actions {
		actionMap[action.DangerLevel] = action
	}
	return &Manager{
		actions: actionMap,
	}
}

// CheckDangerLevel checks if an operation can proceed based on its danger level
func (m *Manager) CheckDangerLevel(dangerLevel string, paramName string, paramValue string, validations []config.Validation) (bool, error) {
	if dangerLevel == "" {
		// No danger level specified, proceed
		return true, nil
	}

	// Check if the parameter value is in the exclude list
	for _, validation := range validations {
		if validation.DangerLevel == dangerLevel {
			for _, exclude := range validation.Exclude {
				if paramValue == exclude {
					return false, fmt.Errorf("parameter %s with value %s is excluded for danger level %s",
						paramName, paramValue, dangerLevel)
				}
			}
		}
	}

	// Get the action for this danger level
	action, exists := m.actions[dangerLevel]
	if !exists {
		// No action defined for this danger level, proceed with warning
		fmt.Printf("Warning: No action defined for danger level %s\n", dangerLevel)
		return true, nil
	}

	// Handle based on action type
	switch action.Type {
	case "confirm":
		return m.handleConfirm(action)
	case "timeout":
		return m.handleTimeout(action)
	case "force":
		return m.handleForce(action)
	default:
		return false, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// handleConfirm handles the confirm action type
func (m *Manager) handleConfirm(action config.Action) (bool, error) {
	message := action.Message
	if message == "" {
		message = fmt.Sprintf("This operation has danger level %s. Do you want to proceed? (y/n): ",
			action.DangerLevel)
	}

	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("error reading response: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// handleTimeout handles the timeout action type
func (m *Manager) handleTimeout(action config.Action) (bool, error) {
	message := action.Message
	if message == "" {
		message = fmt.Sprintf("This operation has danger level %s. It will proceed in %d seconds. Press Ctrl+C to cancel.",
			action.DangerLevel, action.Timeout)
	}

	fmt.Println(message)

	// Wait for the timeout
	for i := action.Timeout; i > 0; i-- {
		fmt.Printf("\rProceeding in %d seconds...", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Println("\rProceeding now...                ")

	return true, nil
}

// handleForce handles the force action type
func (m *Manager) handleForce(action config.Action) (bool, error) {
	message := action.Message
	if message == "" {
		message = fmt.Sprintf("Warning: This operation has danger level %s.", action.DangerLevel)
	}

	fmt.Println(message)
	return true, nil
}
