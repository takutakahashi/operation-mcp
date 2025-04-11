package executor

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHExecutor implements the Executor interface for remote command execution via SSH
type SSHExecutor struct {
	client  *ssh.Client
	config  *SSHConfig
	options *Options
}

// NewSSHExecutor creates a new SSHExecutor with the given configuration
func NewSSHExecutor(config *SSHConfig, options *Options) (*SSHExecutor, error) {
	if config == nil {
		return nil, fmt.Errorf("ssh config is required")
	}

	// Validate required fields
	if config.Host == "" {
		return nil, fmt.Errorf("host is required")
	}

	// Set default values for options if not specified
	if options == nil {
		options = NewOptions()
	}
	if options.Stdin == nil {
		options.Stdin = os.Stdin
	}
	if options.Stdout == nil {
		options.Stdout = os.Stdout
	}
	if options.Stderr == nil {
		options.Stderr = os.Stderr
	}

	// Create SSH client configuration
	sshConfig, err := createSSHConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh client config: %w", err)
	}

	// Connect to the SSH server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ssh server %s: %w", addr, err)
	}

	return &SSHExecutor{
		client:  client,
		config:  config,
		options: options,
	}, nil
}

// createSSHConfig creates an SSH client configuration from SSHConfig
func createSSHConfig(config *SSHConfig) (*ssh.ClientConfig, error) {
	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Timeout:         config.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// If host key verification is enabled, use the known_hosts file
	if config.VerifyHost && config.HostKeyPath != "" {
		hostKeyCallback, err := knownhosts.New(config.HostKeyPath)
		if err != nil {
			// If we can't read the known_hosts file, log a warning but continue
			fmt.Fprintf(os.Stderr, "Warning: cannot use known_hosts file %s: %v\n", config.HostKeyPath, err)
		} else {
			sshConfig.HostKeyCallback = hostKeyCallback
		}
	}

	// Set up authentication methods
	var authMethods []ssh.AuthMethod

	// Try key-based authentication first if a key path is provided
	if config.KeyPath != "" {
		keyAuth, err := publicKeyFile(config.KeyPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot use key file %s: %v\n", config.KeyPath, err)
		} else {
			authMethods = append(authMethods, keyAuth)
		}
	}

	// Add password authentication if provided
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	// If no auth methods are available, return an error
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}

	sshConfig.Auth = authMethods
	return sshConfig, nil
}

// publicKeyFile creates an AuthMethod from a private key file
func publicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(key), nil
}

// Execute runs a command on the remote server and connects its stdout/stderr to the current process
func (e *SSHExecutor) Execute(command []string) error {
	if e.client == nil {
		return fmt.Errorf("ssh client is not connected")
	}

	// Create a new SSH session
	session, err := e.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create ssh session: %w", err)
	}
	defer session.Close()

	// Set up IO
	session.Stdin = e.options.Stdin
	session.Stdout = e.options.Stdout
	session.Stderr = e.options.Stderr

	// Convert command slice to string
	cmdStr := strings.Join(command, " ")

	// Run the command
	return session.Run(cmdStr)
}

// ExecuteWithOutput runs a command on the remote server and returns its combined output
func (e *SSHExecutor) ExecuteWithOutput(command []string) (string, error) {
	if e.client == nil {
		return "", fmt.Errorf("ssh client is not connected")
	}

	// Create a new SSH session
	session, err := e.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create ssh session: %w", err)
	}
	defer session.Close()

	// Set up buffers for output
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	session.Stdin = e.options.Stdin

	// Convert command slice to string
	cmdStr := strings.Join(command, " ")

	// Run the command
	err = session.Run(cmdStr)
	if err != nil {
		// Return stderr if there's an error
		return stderr.String(), err
	}

	return stdout.String(), nil
}

// Close closes the SSH connection
func (e *SSHExecutor) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}

// SSHExecutorFactory creates SSH executors
type SSHExecutorFactory struct {
	config  *SSHConfig
	options *Options
}

// NewSSHExecutorFactory creates a new factory for SSH executors
func NewSSHExecutorFactory(config *SSHConfig, options *Options) *SSHExecutorFactory {
	return &SSHExecutorFactory{
		config:  config,
		options: options,
	}
}

// CreateExecutor creates a new SSHExecutor
func (f *SSHExecutorFactory) CreateExecutor() (Executor, error) {
	return NewSSHExecutor(f.config, f.options)
}
