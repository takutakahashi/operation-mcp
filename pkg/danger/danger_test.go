package danger

import (
	"testing"

	"github.com/takutakahashi/operation-mcp/pkg/config"
)

func TestCheckDangerLevelExclude(t *testing.T) {
	// Create a test manager
	actions := []config.Action{
		{
			DangerLevel: "high",
			Type:        "force",
			Message:     "This is a high danger operation.",
		},
	}
	mgr := NewManager(actions)

	// Create validation rules
	validations := []config.Validation{
		{
			DangerLevel: "high",
			Exclude:     []string{"kube-system", "kube-public"},
		},
	}

	// Test with excluded value
	proceed, err := mgr.CheckDangerLevel("high", "namespace", "kube-system", validations)
	if err == nil {
		t.Errorf("CheckDangerLevel should fail for excluded value")
	}
	if proceed {
		t.Errorf("CheckDangerLevel should return false for excluded value")
	}

	// Test with non-excluded value
	proceed, err = mgr.CheckDangerLevel("high", "namespace", "default", validations)
	if err != nil {
		t.Errorf("CheckDangerLevel failed for non-excluded value: %v", err)
	}
	if !proceed {
		t.Errorf("CheckDangerLevel should return true for non-excluded value")
	}

	// Test with empty danger level
	proceed, err = mgr.CheckDangerLevel("", "namespace", "kube-system", validations)
	if err != nil {
		t.Errorf("CheckDangerLevel failed for empty danger level: %v", err)
	}
	if !proceed {
		t.Errorf("CheckDangerLevel should return true for empty danger level")
	}

	// Test with non-existent danger level
	proceed, err = mgr.CheckDangerLevel("nonexistent", "namespace", "kube-system", validations)
	if err != nil {
		t.Errorf("CheckDangerLevel failed for non-existent danger level: %v", err)
	}
	if !proceed {
		t.Errorf("CheckDangerLevel should return true for non-existent danger level")
	}
}