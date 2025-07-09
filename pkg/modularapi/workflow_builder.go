package modularapi

import (
	"strings"

	"github.com/rrodriguez06/modular_api/pkg/modularapi/workflow"
)

// WorkflowStepTemplate is a template for a workflow step that can be added to a workflow
type WorkflowStepTemplate struct {
	ID            string
	Description   string
	ServiceName   string
	ActionName    string
	Parameters    map[string]interface{}
	DynamicParams map[string]string
	ResultMapping map[string]string
	Condition     *workflow.StepCondition
	ParallelWith  []string
	ErrorHandling workflow.ErrorHandlingStrategy
	MaxRetries    int
}

// NewStepTemplate creates a new workflow step template
func NewWorkflowStepTemplate(id, description string, serviceName, actionName string) *WorkflowStepTemplate {
	return &WorkflowStepTemplate{
		ID:            id,
		Description:   description,
		ServiceName:   serviceName,
		ActionName:    actionName,
		Parameters:    make(map[string]interface{}),
		DynamicParams: make(map[string]string),
		ResultMapping: make(map[string]string),
	}
}

// WithParam adds a parameter to the step template.
// If the value is a string like "{{variable}}", it will be treated as a reference to a workflow variable
func (t *WorkflowStepTemplate) WithParam(name string, value interface{}) *WorkflowStepTemplate {
	// If value is a string and looks like a template variable
	if strValue, isString := value.(string); isString && strings.HasPrefix(strValue, "{{") && strings.HasSuffix(strValue, "}}") {
		// Extract the variable name without the braces
		varName := strings.TrimPrefix(strings.TrimSuffix(strValue, "}}"), "{{")

		// Add it as a dynamic parameter instead
		t.DynamicParams[name] = varName
	} else {
		// Store as a regular parameter
		t.Parameters[name] = value
	}
	return t
}

// WithDynamicParam adds a dynamic parameter to the step template
func (t *WorkflowStepTemplate) WithDynamicParam(paramName, variableName string) *WorkflowStepTemplate {
	t.DynamicParams[paramName] = variableName
	return t
}

// WithResultMap adds a result mapping to the step template
func (t *WorkflowStepTemplate) WithResultMap(responseField, variableName string) *WorkflowStepTemplate {
	t.ResultMapping[responseField] = variableName
	return t
}

// WithCondition adds a condition to the step template
func (t *WorkflowStepTemplate) WithCondition(condType workflow.StepConditionType, sourceVar string, value interface{}) *WorkflowStepTemplate {
	t.Condition = &workflow.StepCondition{
		Type:           condType,
		SourceVariable: sourceVar,
		Value:          value,
	}
	return t
}

// WithParallel specifies that this step runs in parallel with another step
func (t *WorkflowStepTemplate) WithParallel(parallelStepIDs ...string) *WorkflowStepTemplate {
	t.ParallelWith = append(t.ParallelWith, parallelStepIDs...)
	return t
}

// WithErrorHandling sets the error handling strategy for the step template
func (t *WorkflowStepTemplate) WithErrorHandling(strategy workflow.ErrorHandlingStrategy, maxRetries int) *WorkflowStepTemplate {
	t.ErrorHandling = strategy
	t.MaxRetries = maxRetries
	return t
}

// toWorkflowStep converts the template to a workflow.WorkflowStep
func (t *WorkflowStepTemplate) toWorkflowStep() workflow.WorkflowStep {
	return workflow.WorkflowStep{
		ID:            t.ID,
		Description:   t.Description,
		ServiceName:   t.ServiceName,
		ActionName:    t.ActionName,
		Parameters:    t.Parameters,
		DynamicParams: t.DynamicParams,
		ResultMapping: t.ResultMapping,
		Condition:     t.Condition,
		ParallelWith:  t.ParallelWith,
		ErrorHandling: t.ErrorHandling,
		MaxRetries:    t.MaxRetries,
	}
}

// WithWorkflow adds a new workflow to the service
func (b *ServiceBuilder) WithWorkflow(name, description string) *WorkflowBuilder {
	return &WorkflowBuilder{
		serviceBuilder: b,
		workflow: workflow.Workflow{
			Name:        name,
			Description: description,
			Steps:       []workflow.WorkflowStep{},
		},
	}
}

// WorkflowBuilder provides a fluent API for building workflows
type WorkflowBuilder struct {
	serviceBuilder *ServiceBuilder
	workflow       workflow.Workflow
}

// WithStep adds a workflow step to the workflow
func (wb *WorkflowBuilder) WithStep(template *WorkflowStepTemplate) *WorkflowBuilder {
	wb.workflow.Steps = append(wb.workflow.Steps, template.toWorkflowStep())
	return wb
}

// WithVariable adds a variable to the workflow
func (wb *WorkflowBuilder) WithVariable(name string, value interface{}) *WorkflowBuilder {
	if wb.workflow.Variables == nil {
		wb.workflow.Variables = make(map[string]interface{})
	}
	wb.workflow.Variables[name] = value
	return wb
}

// Build completes the workflow definition and returns to the service builder
func (wb *WorkflowBuilder) Build() *ServiceBuilder {
	if wb.serviceBuilder.workflows == nil {
		wb.serviceBuilder.workflows = make(map[string]workflow.Workflow)
	}
	wb.serviceBuilder.workflows[wb.workflow.Name] = wb.workflow
	return wb.serviceBuilder
}
