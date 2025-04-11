package executor

import (
	"io"
)

// Executor defines the interface for command execution
type Executor interface {
	// Execute runs a command and connects its stdout/stderr to the current process
	Execute(command []string) error

	// ExecuteWithOutput runs a command and returns its combined output
	ExecuteWithOutput(command []string) (string, error)

	// Close releases any resources held by the executor
	Close() error
}

// Factory creates an appropriate executor based on configuration
type Factory interface {
	// CreateExecutor creates an executor based on configuration
	CreateExecutor() (Executor, error)
}

// Options contains common options for executors
type Options struct {
	// Stdin is the input stream for the executed command
	Stdin io.Reader

	// Stdout is the output stream for the executed command
	Stdout io.Writer

	// Stderr is the error stream for the executed command
	Stderr io.Writer
}

// NewOptions creates a default Options struct
func NewOptions() *Options {
	return &Options{}
}

// WithStdin sets the stdin option
func (o *Options) WithStdin(stdin io.Reader) *Options {
	o.Stdin = stdin
	return o
}

// WithStdout sets the stdout option
func (o *Options) WithStdout(stdout io.Writer) *Options {
	o.Stdout = stdout
	return o
}

// WithStderr sets the stderr option
func (o *Options) WithStderr(stderr io.Writer) *Options {
	o.Stderr = stderr
	return o
}