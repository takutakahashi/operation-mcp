package executor

import (
	"time"

	"github.com/takutakahashi/operation-mcp/pkg/config"
)

// SSHConfigConverter converts config.SSHConfig to executor.SSHConfig
func SSHConfigConverter(cfg *config.SSHConfig) *SSHConfig {
	sshCfg := NewSSHConfig()

	// If no config provided, return default config
	if cfg == nil {
		return sshCfg
	}

	// Copy values from config.SSHConfig to executor.SSHConfig
	if cfg.Host != "" {
		sshCfg.Host = cfg.Host
	}

	if cfg.Port > 0 {
		sshCfg.Port = cfg.Port
	}

	if cfg.User != "" {
		sshCfg.User = cfg.User
	}

	if cfg.Password != "" {
		sshCfg.Password = cfg.Password
	}

	if cfg.KeyPath != "" {
		sshCfg.KeyPath = cfg.KeyPath
	}

	if cfg.VerifyHost != nil {
		sshCfg.VerifyHost = *cfg.VerifyHost
	}

	if cfg.HostKeyPath != "" {
		sshCfg.HostKeyPath = cfg.HostKeyPath
	}

	if cfg.Timeout > 0 {
		sshCfg.Timeout = time.Duration(cfg.Timeout) * time.Second
	}

	return sshCfg
}
