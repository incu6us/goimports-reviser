package output

import (
	"fmt"
	"os"
)

// Handler defines the interface for output operations
// Following DIP: depend on abstractions
// Following ISP: focused interface
type Handler interface {
	// Write writes the formatted output
	Write(filePath string, content []byte, hasChange bool) error
}

// OutputMode represents the output mode type
type OutputMode string

const (
	ModeFile   OutputMode = "file"
	ModeWrite  OutputMode = "write"
	ModeStdout OutputMode = "stdout"
)

// Config holds output handler configuration
type Config struct {
	Mode         OutputMode
	ListDiff     bool
	IsStandardIn bool
}

// Factory creates appropriate output handler based on configuration
// Following OCP: easy to add new output types without modifying existing code
type Factory struct{}

// NewFactory creates a new output factory
func NewFactory() *Factory {
	return &Factory{}
}

// Create creates an output handler based on configuration
func (f *Factory) Create(cfg Config) (Handler, error) {
	switch cfg.Mode {
	case ModeStdout:
		return NewStdoutHandler(), nil
	case ModeFile, ModeWrite:
		if cfg.IsStandardIn {
			return NewStdoutHandler(), nil
		}
		return NewFileHandler(cfg.ListDiff), nil
	default:
		return nil, fmt.Errorf("invalid output mode %q specified", cfg.Mode)
	}
}

// StdoutHandler writes output to stdout
// Following SRP: single responsibility for stdout output
type StdoutHandler struct{}

// NewStdoutHandler creates a new stdout handler
func NewStdoutHandler() Handler {
	return &StdoutHandler{}
}

// Write writes content to stdout
func (s *StdoutHandler) Write(filePath string, content []byte, hasChange bool) error {
	fmt.Print(string(content))
	return nil
}

// FileHandler writes output to file
// Following SRP: single responsibility for file output
type FileHandler struct {
	listDiff bool
}

// NewFileHandler creates a new file handler
func NewFileHandler(listDiff bool) Handler {
	return &FileHandler{
		listDiff: listDiff,
	}
}

// Write writes content to file
func (f *FileHandler) Write(filePath string, content []byte, hasChange bool) error {
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		return fmt.Errorf("failed to write fixed result to file(%s): %w", filePath, err)
	}

	if f.listDiff && hasChange {
		fmt.Println(filePath)
	}

	return nil
}

// DiffListHandler only lists files that have changes
// Following SRP: single responsibility for listing differences
type DiffListHandler struct{}

// NewDiffListHandler creates a new diff list handler
func NewDiffListHandler() Handler {
	return &DiffListHandler{}
}

// Write lists the file path if there are changes
func (d *DiffListHandler) Write(filePath string, content []byte, hasChange bool) error {
	if hasChange {
		fmt.Println(filePath)
	}
	return nil
}
