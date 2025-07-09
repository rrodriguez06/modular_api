package modularapi

import (
	"github.com/rrodriguez06/modular_api/internal/log"
)

// ExecutionOption defines a function type that configures execution
type ExecutionOption func(*executionConfig)

// executionConfig holds the internal configuration for execution
type executionConfig struct {
	WorkflowVars *map[string]interface{}
	LogLevel     *log.LogLevel
	// Other options could be added here in the future
}

// WithWorkflowVars creates an option to capture workflow variables
func WithWorkflowVars(vars *map[string]interface{}) ExecutionOption {
	return func(c *executionConfig) {
		c.WorkflowVars = vars
	}
}

// WithLogLevel creates an option to set logging level for the execution
func WithLogLevel(level log.LogLevel) ExecutionOption {
	return func(c *executionConfig) {
		c.LogLevel = &level
	}
}

// RequestOption defines a function type that configures individual API requests
type RequestOption func(*requestConfig)

// requestConfig holds the internal configuration for API requests
type requestConfig struct {
	LogLevel *log.LogLevel
	// Other options could be added here in the future
}

// WithRequestLogLevel creates an option to set logging level for a specific request
func WithRequestLogLevel(level log.LogLevel) RequestOption {
	return func(c *requestConfig) {
		c.LogLevel = &level
	}
}
