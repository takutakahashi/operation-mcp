package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// LocalExecutor implements the Executor interface for local command execution
type LocalExecutor struct {
	options *Options
}

// NewLocalExecutor creates a new LocalExecutor with the given options
func NewLocalExecutor(options *Options) *LocalExecutor {
	// If options is nil, use default options
	if options == nil {
		options = NewOptions()
	}

	// Set default values if not specified
	if options.Stdin == nil {
		options.Stdin = os.Stdin
	}
	if options.Stdout == nil {
		options.Stdout = os.Stdout
	}
	if options.Stderr == nil {
		options.Stderr = os.Stderr
	}

	return &LocalExecutor{
		options: options,
	}
}

// Execute runs a command locally and connects its stdout/stderr to the current process
func (e *LocalExecutor) Execute(command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = e.options.Stdin
	cmd.Stdout = e.options.Stdout
	cmd.Stderr = e.options.Stderr

	return cmd.Run()
}

// ExecuteWithOutput runs a command locally and returns its combined output
func (e *LocalExecutor) ExecuteWithOutput(command []string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("empty command")
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = e.options.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

// Close does nothing for LocalExecutor as there are no resources to release
func (e *LocalExecutor) Close() error {
	return nil
}

// LocalExecutorFactory creates local executors
type LocalExecutorFactory struct {
	options *Options
}

// NewLocalExecutorFactory creates a new factory for local executors
func NewLocalExecutorFactory(options *Options) *LocalExecutorFactory {
	return &LocalExecutorFactory{
		options: options,
	}
}

// CreateExecutor creates a new LocalExecutor
func (f *LocalExecutorFactory) CreateExecutor() (Executor, error) {
	return NewLocalExecutor(f.options), nil
}
