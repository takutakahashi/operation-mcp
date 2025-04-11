package main

import (
	"fmt"
	"time"

	"github.com/takutakahashi/operation-mcp/pkg/executor"
)

// createExecutor creates an executor based on command-line flags
func createExecutor() (executor.Executor, error) {
	// If remote mode is not enabled, use a local executor
	if !remoteMode {
		return executor.NewLocalExecutor(nil), nil
	}
	
	// Create SSH config
	sshConfig := executor.NewSSHConfig()
	
	// Override with command-line flags
	if sshHost != "" {
		sshConfig.Host = sshHost
	} else if cfg != nil && cfg.SSH.Host != "" {
		sshConfig.Host = cfg.SSH.Host
	}
	
	if sshUser != "" {
		sshConfig.User = sshUser
	} else if cfg != nil && cfg.SSH.User != "" {
		sshConfig.User = cfg.SSH.User
	}
	
	if sshKeyPath != "" {
		sshConfig.KeyPath = sshKeyPath
	} else if cfg != nil && cfg.SSH.KeyPath != "" {
		sshConfig.KeyPath = cfg.SSH.KeyPath
	}
	
	if sshPassword != "" {
		sshConfig.Password = sshPassword
	} else if cfg != nil && cfg.SSH.Password != "" {
		sshConfig.Password = cfg.SSH.Password
	}
	
	if sshPort != 22 {
		sshConfig.Port = sshPort
	} else if cfg != nil && cfg.SSH.Port != 0 {
		sshConfig.Port = cfg.SSH.Port
	}
	
	// Convert from config.SSHConfig.VerifyHost (pointer) to executor.SSHConfig.VerifyHost (bool)
	configVerifyHost := true
	if cfg != nil && cfg.SSH.VerifyHost != nil {
		configVerifyHost = *cfg.SSH.VerifyHost
	}
	
	// Command line flag overrides config file
	if sshVerifyHost != configVerifyHost {
		sshConfig.VerifyHost = sshVerifyHost
	} else {
		sshConfig.VerifyHost = configVerifyHost
	}
	
	if sshTimeout != 10*time.Second {
		sshConfig.Timeout = sshTimeout
	} else if cfg != nil && cfg.SSH.Timeout > 0 {
		sshConfig.Timeout = time.Duration(cfg.SSH.Timeout) * time.Second
	}
	
	// Validate the SSH config
	if sshConfig.Host == "" {
		return nil, fmt.Errorf("SSH host is required in remote mode")
	}
	
	// Create SSH executor
	return executor.NewSSHExecutor(sshConfig, nil)
}