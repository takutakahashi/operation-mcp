package executor

import (
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// SSHConfig contains configuration options for SSH connections
type SSHConfig struct {
	// Host is the hostname or IP address of the SSH server
	Host string `yaml:"host"`

	// Port is the port number of the SSH server
	Port int `yaml:"port"`

	// User is the username to authenticate as
	User string `yaml:"user"`

	// Password is the password for password authentication
	// Note: key-based authentication is preferred
	Password string `yaml:"password,omitempty"`

	// KeyPath is the path to the private key file for key authentication
	KeyPath string `yaml:"key"`

	// VerifyHost determines whether to verify the host key
	VerifyHost bool `yaml:"verify_host"`

	// HostKeyPath is the path to the known_hosts file
	HostKeyPath string `yaml:"host_key_path,omitempty"`

	// Timeout is the maximum amount of time for the connection
	Timeout time.Duration `yaml:"timeout"`
}

// NewSSHConfig creates a new SSH configuration with default values
func NewSSHConfig() *SSHConfig {
	// Get current user for default username and home directory
	currentUser, err := user.Current()
	username := ""
	homeDir := ""
	
	if err == nil {
		username = currentUser.Username
		homeDir = currentUser.HomeDir
	}

	// Default SSH key path
	keyPath := ""
	if homeDir != "" {
		keyPath = filepath.Join(homeDir, ".ssh", "id_rsa")
		// If the default key doesn't exist, try id_ed25519
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			altKeyPath := filepath.Join(homeDir, ".ssh", "id_ed25519")
			if _, err := os.Stat(altKeyPath); err == nil {
				keyPath = altKeyPath
			}
		}
	}

	// Default known_hosts path
	knownHostsPath := ""
	if homeDir != "" {
		knownHostsPath = filepath.Join(homeDir, ".ssh", "known_hosts")
	}

	return &SSHConfig{
		Port:        22,
		User:        username,
		KeyPath:     keyPath,
		VerifyHost:  true,
		HostKeyPath: knownHostsPath,
		Timeout:     10 * time.Second,
	}
}