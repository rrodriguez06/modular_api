package workflow

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

// ErrInvalidTemplateID is returned when a template ID is not in the format "service.action"
var ErrInvalidTemplateID = fmt.Errorf("invalid template ID, must be in format 'service.action'")

// SplitTemplateID splits a template ID in the format "service.action" into its components
func SplitTemplateID(templateID string) []string {
	return strings.Split(templateID, ".")
}

// StepConditionType defines the type of condition for workflow steps
type StepConditionType string

const (
	// ConditionExists checks if a variable exists and is not nil
	ConditionExists StepConditionType = "exists"
	// ConditionEquals checks if a variable equals a value
	ConditionEquals StepConditionType = "equals"
	// ConditionContains checks if a variable contains a value (string or slice)
	ConditionContains StepConditionType = "contains"
	// ConditionGreaterThan checks if a variable is greater than a value
	ConditionGreaterThan StepConditionType = "greater_than"
	// ConditionLessThan checks if a variable is less than a value
	ConditionLessThan StepConditionType = "less_than"
)

// ErrorHandlingStrategy defines how errors in workflow steps are handled
type ErrorHandlingStrategy string

const (
	// ContinueOnError continues to the next step even if the current step fails
	ContinueOnError ErrorHandlingStrategy = "continue"
	// AbortOnError aborts the workflow if any step fails
	AbortOnError ErrorHandlingStrategy = "abort"
	// RetryOnError retries the step if it fails
	RetryOnError ErrorHandlingStrategy = "retry"
)

// StepCondition defines a condition that must be met for a workflow step to execute
type StepCondition struct {
	Type           StepConditionType `json:"type"`
	SourceVariable string            `json:"source_variable"`
	Value          interface{}       `json:"value,omitempty"`
}

// WorkflowStep defines a single step in a workflow
type WorkflowStep struct {
	ID            string                 `json:"id"`                       // Unique identifier for this step within the workflow
	Description   string                 `json:"description"`              // Human-readable description
	ServiceName   string                 `json:"service_name"`             // The service to use
	ActionName    string                 `json:"action_name"`              // The template action to use
	Parameters    map[string]interface{} `json:"parameters"`               // Fixed parameters for this step
	DynamicParams map[string]string      `json:"dynamic_params"`           // Parameters sourced from variables
	ResultMapping map[string]string      `json:"result_mapping"`           // Map response fields to variables
	Condition     *StepCondition         `json:"condition,omitempty"`      // Condition to execute this step
	ParallelWith  []string               `json:"parallel_with,omitempty"`  // IDs of steps to execute in parallel with
	ErrorHandling ErrorHandlingStrategy  `json:"error_handling,omitempty"` // How to handle errors
	MaxRetries    int                    `json:"max_retries,omitempty"`    // Maximum number of retries (for retry strategy)
	RetryDelayMs  int                    `json:"retry_delay_ms,omitempty"` // Delay between retries in milliseconds
}

// Workflow defines a sequence of API calls with dependencies between them
type Workflow struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []WorkflowStep         `json:"steps"`
	Variables   map[string]interface{} `json:"variables,omitempty"` // Default workflow variables
}

// WorkflowService defines the interface for working with workflows
type WorkflowService interface {
	// RegisterWorkflow adds a workflow to the registry
	RegisterWorkflow(workflow Workflow) error

	// ExecuteWorkflow runs a workflow with the given initial parameters
	// If result is not nil, the response of the last step will be unmarshalled into it
	ExecuteWorkflow(name string, initialParams map[string]interface{}, result interface{}) (map[string]interface{}, error)

	// GetWorkflow returns a workflow by name
	GetWorkflow(name string) (Workflow, bool)

	// ListWorkflows returns a list of all registered workflow names
	ListWorkflows() []string

	// SaveWorkflows saves all workflows to a file
	SaveWorkflows(filepath string) error

	// LoadWorkflows loads workflows from a file
	LoadWorkflows(filepath string) error
}

// stepExecutionResult holds the result of a workflow step execution
type stepExecutionResult struct {
	StepID string
	Result map[string]interface{}
	Error  error
}

// APIServiceExecutor defines the minimal interface that the workflow package needs from a service
type APIServiceExecutor interface {
	// ExecuteServiceAction executes an API request and unmarshals the result into the given interface
	ExecuteServiceAction(serviceName, actionName string, params map[string]interface{}, result interface{}) error
}

// WorkflowExecutor executes workflows using a modular API service
type WorkflowExecutor struct {
	service   APIServiceExecutor
	workflows map[string]Workflow
	mu        sync.RWMutex
}

// NewWorkflowExecutor creates a new workflow executor
func NewWorkflowExecutor(service APIServiceExecutor) *WorkflowExecutor {
	return &WorkflowExecutor{
		service:   service,
		workflows: make(map[string]Workflow),
	}
}

// RegisterWorkflow implements WorkflowService
func (we *WorkflowExecutor) RegisterWorkflow(workflow Workflow) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	// Validate workflow
	if workflow.Name == "" {
		return fmt.Errorf("workflow must have a name")
	}

	// Validate steps
	stepIDs := make(map[string]bool)
	for _, step := range workflow.Steps {
		if step.ID == "" {
			return fmt.Errorf("step in workflow %s must have an ID", workflow.Name)
		}

		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID %s in workflow %s", step.ID, workflow.Name)
		}
		stepIDs[step.ID] = true

		if step.ServiceName == "" || step.ActionName == "" {
			return fmt.Errorf("step %s in workflow %s must have a service name and action name",
				step.ID, workflow.Name)
		}

		// Validate parallel execution references
		for _, parallelID := range step.ParallelWith {
			if !stepIDs[parallelID] {
				return fmt.Errorf("step %s references unknown parallel step ID %s",
					step.ID, parallelID)
			}
		}
	}

	we.workflows[workflow.Name] = workflow
	return nil
}

// ExecuteWorkflow implements WorkflowService
func (we *WorkflowExecutor) ExecuteWorkflow(name string, initialParams map[string]interface{}, result interface{}) (map[string]interface{}, error) {
	we.mu.RLock()
	workflow, exists := we.workflows[name]
	we.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow %s not found", name)
	}

	// Create workflow context with variables
	variables := make(map[string]interface{})

	// Add default workflow variables
	for k, v := range workflow.Variables {
		variables[k] = v
	}

	// Add initial parameters (override defaults)
	for k, v := range initialParams {
		variables[k] = v
	}

	// Track executed steps to manage dependencies
	executedSteps := make(map[string]bool)
	stepResults := make(map[string]map[string]interface{})

	// Process steps
	for i := 0; i < len(workflow.Steps); i++ {
		step := workflow.Steps[i]

		// Check if this step should run in parallel with others
		parallelSteps := []WorkflowStep{step}
		for j := i + 1; j < len(workflow.Steps); j++ {
			nextStep := workflow.Steps[j]
			for _, parallelID := range nextStep.ParallelWith {
				if parallelID == step.ID {
					// This next step should run in parallel
					parallelSteps = append(parallelSteps, nextStep)
					// Mark this step as processed so we skip it in the main loop
					executedSteps[nextStep.ID] = true
				}
			}
		}

		// Skip if this step was already executed in parallel
		if executedSteps[step.ID] {
			continue
		}

		// Execute parallel steps
		results := we.executeParallelSteps(parallelSteps, variables)

		// Process results
		for _, result := range results {
			// Mark step as executed
			executedSteps[result.StepID] = true

			// Handle errors based on strategy
			if result.Error != nil {
				// Find the step with this ID
				var errorStep *WorkflowStep
				for _, s := range parallelSteps {
					if s.ID == result.StepID {
						errorStep = &s
						break
					}
				}

				// Default to abort on error if not specified
				strategy := AbortOnError
				if errorStep != nil && errorStep.ErrorHandling != "" {
					strategy = errorStep.ErrorHandling
				}

				// Handle error based on strategy
				switch strategy {
				case ContinueOnError:
					// Just continue to next step
					continue
				case RetryOnError:
					// Not implemented in this version
					// Would need loop and delay logic
					return nil, fmt.Errorf("retry strategy not implemented")
				case AbortOnError:
					// Default behavior - abort workflow
					return nil, fmt.Errorf("workflow step %s failed: %w", result.StepID, result.Error)
				}
			}

			// Store result for this step
			stepResults[result.StepID] = result.Result

			// Update variables based on result mapping
			// Find the step with this ID to get mapping
			for _, s := range parallelSteps {
				if s.ID == result.StepID {
					for responseField, variableName := range s.ResultMapping {
						// Extract value using dot notation
						value, ok := extractValue(result.Result, responseField)
						if ok {
							variables[variableName] = value
							log.Printf("Mapped result field '%s' to variable '%s' with value: %v",
								responseField, variableName, value)
						} else {
							log.Printf("Warning: Could not extract field '%s' from response for step %s",
								responseField, s.ID)

							// Debug: print the available fields in the result
							resultKeys := make([]string, 0)
							for k := range result.Result {
								resultKeys = append(resultKeys, k)
							}
							log.Printf("Available fields in response: %v", resultKeys)
						}
					}
					break
				}
			}
		}
	}

	// If result parameter is provided and we had any steps, map the last step's response to it
	if result != nil && len(workflow.Steps) > 0 {
		// Find the last step that was executed
		var lastStepResult map[string]interface{}
		var lastStepID string

		// Go through steps in reverse order to find the last executed one
		for i := len(workflow.Steps) - 1; i >= 0; i-- {
			step := workflow.Steps[i]
			if stepResult, exists := stepResults[step.ID]; exists {
				lastStepResult = stepResult
				lastStepID = step.ID
				break
			}
		}

		if lastStepResult != nil {
			// Convert to JSON and unmarshal to the result
			jsonData, err := json.Marshal(lastStepResult)
			if err != nil {
				return variables, fmt.Errorf("error marshaling last step result: %w", err)
			}

			if err := json.Unmarshal(jsonData, result); err != nil {
				return variables, fmt.Errorf("error unmarshaling last step result to provided result variable: %w", err)
			}

			log.Printf("Mapped last step (%s) response to result parameter", lastStepID)
		}
	}

	return variables, nil
}

// executeParallelSteps executes a set of steps in parallel
func (we *WorkflowExecutor) executeParallelSteps(steps []WorkflowStep, variables map[string]interface{}) []stepExecutionResult {
	var wg sync.WaitGroup
	resultChan := make(chan stepExecutionResult, len(steps))

	for _, step := range steps {
		wg.Add(1)
		go func(s WorkflowStep) {
			defer wg.Done()

			result := stepExecutionResult{
				StepID: s.ID,
			}

			// Check if condition is met
			if s.Condition != nil {
				conditionMet, err := evaluateCondition(s.Condition, variables)
				if err != nil {
					result.Error = fmt.Errorf("error evaluating condition for step %s: %w", s.ID, err)
					resultChan <- result
					return
				}

				if !conditionMet {
					// Condition not met, skip this step
					result.Result = make(map[string]interface{})
					resultChan <- result
					return
				}
			}

			// Prepare parameters
			params := make(map[string]interface{})

			// Process fixed parameters - check for template expressions
			for k, v := range s.Parameters {
				// If the parameter value is a string, check if it's a template expression
				if strValue, isString := v.(string); isString && isExpression(strValue) {
					evaluatedValue, err := evaluateExpression(strValue, variables)
					if err != nil {
						result.Error = fmt.Errorf("error evaluating expression for fixed parameter %s: %w", k, err)
						resultChan <- result
						return
					}
					params[k] = evaluatedValue
					log.Printf("Processed template parameter %s: '%s' -> '%v'", k, strValue, evaluatedValue)
				} else {
					// Not a template expression, use as-is
					params[k] = v
				}
			}

			// Add dynamic parameters
			for paramName, variableName := range s.DynamicParams {
				// Check if we need to evaluate an expression
				if isExpression(variableName) {
					evaluatedValue, err := evaluateExpression(variableName, variables)
					if err != nil {
						result.Error = fmt.Errorf("error evaluating expression for parameter %s: %w", paramName, err)
						resultChan <- result
						return
					}
					params[paramName] = evaluatedValue
					log.Printf("Processed dynamic parameter %s using expression '%s' -> '%v'",
						paramName, variableName, evaluatedValue)
				} else {
					// Simple variable reference
					if value, exists := variables[variableName]; exists {
						params[paramName] = value
						log.Printf("Set dynamic parameter %s from variable '%s' -> '%v'",
							paramName, variableName, value)
					} else {
						// If variable doesn't exist, log a warning
						log.Printf("Warning: Variable %s not found for parameter %s in step %s",
							variableName, paramName, s.ID)
					}
				}
			}

			// Execute the API request
			var apiResult map[string]interface{}
			err := we.service.ExecuteServiceAction(s.ServiceName, s.ActionName, params, &apiResult)
			if err != nil {
				result.Error = err
				resultChan <- result
				return
			}

			result.Result = apiResult
			resultChan <- result

		}(step)
	}

	// Wait for all steps to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	var results []stepExecutionResult
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// GetWorkflow implements WorkflowService
func (we *WorkflowExecutor) GetWorkflow(name string) (Workflow, bool) {
	we.mu.RLock()
	defer we.mu.RUnlock()

	workflow, exists := we.workflows[name]
	return workflow, exists
}

// ListWorkflows implements WorkflowService
func (we *WorkflowExecutor) ListWorkflows() []string {
	we.mu.RLock()
	defer we.mu.RUnlock()

	var names []string
	for name := range we.workflows {
		names = append(names, name)
	}

	return names
}

// SaveWorkflows implements WorkflowService
func (we *WorkflowExecutor) SaveWorkflows(filepath string) error {
	we.mu.RLock()
	defer we.mu.RUnlock()

	data, err := json.MarshalIndent(we.workflows, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling workflows: %w", err)
	}

	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing workflows to file: %w", err)
	}

	return nil
}

// LoadWorkflows implements WorkflowService
func (we *WorkflowExecutor) LoadWorkflows(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("error reading workflows file: %w", err)
	}

	var workflows map[string]Workflow
	err = json.Unmarshal(data, &workflows)
	if err != nil {
		return fmt.Errorf("error unmarshaling workflows: %w", err)
	}

	// Register each workflow (which also validates it)
	for _, workflow := range workflows {
		err = we.RegisterWorkflow(workflow)
		if err != nil {
			return fmt.Errorf("error registering workflow %s: %w", workflow.Name, err)
		}
	}

	return nil
}
