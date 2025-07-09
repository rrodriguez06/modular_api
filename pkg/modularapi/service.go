package modularapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rrodriguez06/modular_api/internal/log"
	"github.com/rrodriguez06/modular_api/pkg/modularapi/client"
	"github.com/rrodriguez06/modular_api/pkg/modularapi/config"
	"github.com/rrodriguez06/modular_api/pkg/modularapi/template"
	"github.com/rrodriguez06/modular_api/pkg/modularapi/workflow"
)

// Service is the main interface for the modular API service
type Service interface {
	// Request preparation and execution
	PrepareRequest(serviceName, action string, params map[string]interface{}) (*http.Request, error)
	MakeRequest(req *http.Request, result interface{}) error
	MakeStreamingRequest(req *http.Request, w http.ResponseWriter) (string, error)
	PerformRequest(serviceName, action string, params map[string]interface{}, result interface{}) error
	PerformStreamingRequest(serviceName, action string, params map[string]interface{}, w http.ResponseWriter) (string, error)
	ExecuteRequestWithParams(templateID string, params map[string]interface{}) (json.RawMessage, error)

	// Template management
	AddRouteTemplate(serviceName, action string, route template.RouteTemplate)
	SaveTemplates(filepath string) error
	LoadTemplates(filepath string) error

	// Service configuration
	GetServiceURL(serviceName string) string
	SetServiceURL(serviceName, url string)
	GetServiceToken(serviceName string) string

	// Headers management
	SetServiceHeaders(serviceName string, headers map[string]string)
	GetServiceHeaders(serviceName string) map[string]string
	RemoveServiceHeader(serviceName string, headerName string)

	// Parameters management
	SetServiceParams(serviceName string, params map[string]interface{})
	GetServiceParams(serviceName string) map[string]interface{}
	RemoveServiceParam(serviceName string, paramName string)

	// Workflow management
	RegisterWorkflow(wf workflow.Workflow) error
	AddWorkflowStep(workflowName string, step workflow.WorkflowStep) error
	ExecuteWorkflow(name string, params map[string]interface{}, result interface{}) (map[string]interface{}, error)
	GetWorkflow(name string) (workflow.Workflow, bool)
	ListWorkflows() []string
	SaveWorkflows(filepath string) error
	LoadWorkflows(filepath string) error
}

// ModularAPIService implements the Service interface
type ModularAPIService struct {
	config           *config.Config
	templateStore    *template.TemplateStore
	httpClient       *client.Client
	streamClient     *client.StreamingClient
	serviceHeaders   map[string]map[string]string      // Service-level headers
	serviceParams    map[string]map[string]interface{} // Service-level parameters
	workflowExecutor *workflow.WorkflowExecutor        // Workflow executor
}

// NewService creates a new modular API service
func NewService(cfg *config.Config) Service {
	service := &ModularAPIService{
		config:         cfg,
		templateStore:  template.NewTemplateStore(),
		httpClient:     client.NewClient(180 * time.Second), // Default timeout of 3 minutes
		streamClient:   client.NewStreamingClient(),
		serviceHeaders: make(map[string]map[string]string),
		serviceParams:  make(map[string]map[string]interface{}),
	}

	// Initialize workflow executor after the service is created
	service.workflowExecutor = workflow.NewWorkflowExecutor(service)

	return service
}

// PrepareRequest prepares a request using the template and provided parameters
func (s *ModularAPIService) PrepareRequest(serviceName, action string, params map[string]interface{}) (*http.Request, error) {
	tmpl, ok := s.templateStore.GetTemplate(serviceName, action)
	if !ok {
		return nil, fmt.Errorf("no template found for action: %s in service %s", action, serviceName)
	}

	cfg, ok := s.config.GetServiceConfig(serviceName)
	if !ok {
		return nil, fmt.Errorf("no configuration found for service: %s", serviceName)
	}

	log.GlobalLogger.Infof("Preparing request from template: %s %s for action %s.%s\n", tmpl.Method, tmpl.Endpoint, serviceName, action)

	// Prepare all parameters in the correct order of precedence:
	// 1. First add default parameters from service configuration
	mergedParams := make(map[string]interface{})
	if cfg.DefaultParams != nil {
		for key, value := range cfg.DefaultParams {
			mergedParams[key] = value
		}
	}

	// 2. Add global service parameters (which override defaults)
	if globalParams, ok := s.serviceParams[serviceName]; ok {
		for k, v := range globalParams {
			mergedParams[k] = v
		}
	}

	// 3. Finally add request-specific parameters (which override both)
	for k, v := range params {
		mergedParams[k] = v
	}

	// Log the final merged parameters for debugging
	debugParamsJson, _ := json.MarshalIndent(mergedParams, "", "  ")
	log.GlobalLogger.Infof("Merged parameters: %s", string(debugParamsJson))

	// Build the URL with path parameters
	endpoint := tmpl.Endpoint
	for _, pathParam := range tmpl.PathParams {
		// Check for both regular and optional placeholders for this param
		regularPlaceholder := "{{" + pathParam + "}}"
		optionalPlaceholder := "{{" + pathParam + "?}}"

		if value, exists := mergedParams[pathParam]; exists {
			// Replace both regular and optional placeholders with the value
			endpoint = strings.ReplaceAll(endpoint, regularPlaceholder, fmt.Sprintf("%v", value))
			endpoint = strings.ReplaceAll(endpoint, optionalPlaceholder, fmt.Sprintf("%v", value))
		} else if strings.Contains(endpoint, optionalPlaceholder) {
			// Handle optional path parameters that aren't provided
			// We need to remove the entire segment from the URL path
			parts := strings.Split(endpoint, "/")
			for i, part := range parts {
				if part == optionalPlaceholder {
					// Remove this segment
					parts = append(parts[:i], parts[i+1:]...)
					break
				}
			}
			endpoint = strings.Join(parts, "/")
		} else if tmpl.OptionalParams[pathParam] {
			// If parameter is marked as optional in our map, we can skip it
			continue
		} else {
			return nil, fmt.Errorf("missing required path parameter: %s", pathParam)
		}
	}

	url := cfg.ApiURL + endpoint

	// Prepare request body if template has one
	var processedBody map[string]interface{}
	if tmpl.Body != nil {
		// Process body template values
		processedBody = make(map[string]interface{})
		for key, value := range tmpl.Body {
			if processedValue, valid := template.ProcessTemplateValue(value, mergedParams, tmpl.OptionalParams); valid {
				processedBody[key] = processedValue
			} else {
				// Check if this is an optional parameter
				stringValue, isString := value.(string)
				if isString && (strings.HasSuffix(strings.TrimPrefix(strings.TrimSuffix(stringValue, "}}"), "{{"), "?") ||
					tmpl.OptionalParams[key]) {
					// Skip optional parameters that aren't provided
					continue
				}

				return nil, fmt.Errorf("missing required body parameter for key: %s", key)
			}
		}

		// Only include the body if we have parameters to send
		if len(processedBody) > 0 {
			// For debugging purposes only
			debugJson, _ := json.MarshalIndent(processedBody, "", "  ")
			log.GlobalLogger.Infof("Request body (debug): %s", string(debugJson))
		}
	}

	// Create the request with the properly formatted JSON body
	var req *http.Request
	var err error

	if len(processedBody) > 0 {
		// Use json.MarshalIndent to create a clean, formatted JSON string
		formattedJSON, err := json.MarshalIndent(processedBody, "", "  ")
		if err != nil {
			log.GlobalLogger.Errorf("Failed to marshal request body: %v", err)
			return nil, err
		}

		// Log the exact JSON that will be sent
		log.GlobalLogger.Infof("Raw JSON body to be sent: %s", string(formattedJSON))

		// Create the request with the formatted JSON
		req, err = http.NewRequest(tmpl.Method, url, bytes.NewReader(formattedJSON))
	} else {
		// Create request without body
		req, err = http.NewRequest(tmpl.Method, url, nil)
	}

	if err != nil {
		log.GlobalLogger.Errorf("Failed to create request: %v", err)
		return nil, err
	}

	// Add headers in the following order:
	// 1. Global headers for the service
	if globalHeaders, ok := s.serviceHeaders[serviceName]; ok {
		for key, value := range globalHeaders {
			req.Header.Set(key, value)
		}
	}

	// 2. Route-specific headers (can override global headers)
	for key, value := range tmpl.Headers {
		req.Header.Set(key, value)
	}

	// 3. Authorization header if token is provided
	if cfg.ApiToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.ApiToken)
	}

	// Process query parameters from template only
	if tmpl.QueryParams != nil {
		q := req.URL.Query()
		for key, value := range tmpl.QueryParams {
			if processedValue, valid := template.ProcessTemplateValue(value, mergedParams, tmpl.OptionalParams); valid {
				q.Set(key, fmt.Sprintf("%v", processedValue))
			} else {
				// Check if this is an optional parameter
				stringValue, isString := value.(string)
				if isString && (strings.HasSuffix(strings.TrimPrefix(strings.TrimSuffix(stringValue, "}}"), "{{"), "?") ||
					tmpl.OptionalParams[key]) {
					// Skip optional parameters that aren't provided
					continue
				}

				return nil, fmt.Errorf("missing required query parameter: %s", key)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

// MakeRequest performs an HTTP request and unmarshals the response into the result
func (s *ModularAPIService) MakeRequest(req *http.Request, result interface{}) error {
	return s.httpClient.MakeRequest(req, result)
}

// MakeStreamingRequest performs a streaming HTTP request
func (s *ModularAPIService) MakeStreamingRequest(req *http.Request, w http.ResponseWriter) (string, error) {
	return s.streamClient.MakeStreamingRequest(req, w)
}

// PerformRequest combines PrepareRequest and MakeRequest into a single function
func (s *ModularAPIService) PerformRequest(serviceName, action string, params map[string]interface{}, result interface{}) error {
	req, err := s.PrepareRequest(serviceName, action, params)
	if err != nil {
		return fmt.Errorf("failed to prepare request: %w", err)
	}

	err = s.MakeRequest(req, result)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	return nil
}

// PerformStreamingRequest performs a streaming request using the template and parameters
func (s *ModularAPIService) PerformStreamingRequest(serviceName, action string, params map[string]interface{}, w http.ResponseWriter) (string, error) {
	req, err := s.PrepareRequest(serviceName, action, params)
	if err != nil {
		return "", fmt.Errorf("failed to prepare streaming request: %w", err)
	}

	response, err := s.MakeStreamingRequest(req, w)
	if err != nil {
		return "", fmt.Errorf("failed to make streaming request: %w", err)
	}

	return response, nil
}

// AddRouteTemplate adds a route template for a specific service and action
func (s *ModularAPIService) AddRouteTemplate(serviceName, action string, route template.RouteTemplate) {
	s.templateStore.AddTemplate(serviceName, action, route)
}

// SaveTemplates saves the current template configuration to a JSON file
func (s *ModularAPIService) SaveTemplates(filepath string) error {
	return s.templateStore.SaveToFile(filepath)
}

// LoadTemplates loads template configuration from a JSON file and merges it with existing templates
func (s *ModularAPIService) LoadTemplates(filepath string) error {
	return s.templateStore.LoadFromFile(filepath)
}

// GetServiceURL returns the URL for a specific service
func (s *ModularAPIService) GetServiceURL(serviceName string) string {
	if cfg, ok := s.config.GetServiceConfig(serviceName); ok {
		return cfg.ApiURL
	}
	return ""
}

// SetServiceURL sets the URL for a specific service
func (s *ModularAPIService) SetServiceURL(serviceName, url string) {
	if cfg, ok := s.config.GetServiceConfig(serviceName); ok {
		cfg.ApiURL = url
		s.config.SetServiceConfig(serviceName, cfg)
	}
}

// GetServiceToken returns the token for a specific service
func (s *ModularAPIService) GetServiceToken(serviceName string) string {
	if cfg, ok := s.config.GetServiceConfig(serviceName); ok {
		return cfg.ApiToken
	}
	return ""
}

// SetServiceHeaders sets global headers for a specific service
func (s *ModularAPIService) SetServiceHeaders(serviceName string, headers map[string]string) {
	if s.serviceHeaders[serviceName] == nil {
		s.serviceHeaders[serviceName] = make(map[string]string)
	}
	for k, v := range headers {
		s.serviceHeaders[serviceName][k] = v
	}
}

// GetServiceHeaders gets the global headers for a specific service
func (s *ModularAPIService) GetServiceHeaders(serviceName string) map[string]string {
	if headers, ok := s.serviceHeaders[serviceName]; ok {
		// Return a copy to prevent modification of internal state
		result := make(map[string]string)
		for k, v := range headers {
			result[k] = v
		}
		return result
	}
	return nil
}

// RemoveServiceHeader removes a global header from a service
func (s *ModularAPIService) RemoveServiceHeader(serviceName string, headerName string) {
	if headers, ok := s.serviceHeaders[serviceName]; ok {
		delete(headers, headerName)
	}
}

// SetServiceParams sets global parameters for a specific service
func (s *ModularAPIService) SetServiceParams(serviceName string, params map[string]interface{}) {
	if s.serviceParams[serviceName] == nil {
		s.serviceParams[serviceName] = make(map[string]interface{})
	}
	for k, v := range params {
		s.serviceParams[serviceName][k] = v
	}
}

// GetServiceParams gets the global parameters for a specific service
func (s *ModularAPIService) GetServiceParams(serviceName string) map[string]interface{} {
	if params, ok := s.serviceParams[serviceName]; ok {
		// Return a copy to prevent modification of internal state
		result := make(map[string]interface{})
		for k, v := range params {
			result[k] = v
		}
		return result
	}
	return nil
}

// RemoveServiceParam removes a global parameter from a service
func (s *ModularAPIService) RemoveServiceParam(serviceName string, paramName string) {
	if params, ok := s.serviceParams[serviceName]; ok {
		delete(params, paramName)
	}
}

// ExecuteRequestWithParams is a helper method for executing a request with parameters
func (s *ModularAPIService) ExecuteRequestWithParams(templateID string, params map[string]interface{}) (json.RawMessage, error) {
	// Split template ID into service and action
	parts := workflow.SplitTemplateID(templateID)
	if len(parts) != 2 {
		return nil, workflow.ErrInvalidTemplateID
	}

	serviceName, actionName := parts[0], parts[1]

	// Use a map to receive the JSON response
	var result map[string]interface{}

	// Execute the request
	err := s.PerformRequest(serviceName, actionName, params, &result)
	if err != nil {
		return nil, err
	}

	// Convert back to JSON for the raw message
	return json.Marshal(result)
}

// RegisterWorkflow registers a new workflow with the service
func (s *ModularAPIService) RegisterWorkflow(wf workflow.Workflow) error {
	return s.workflowExecutor.RegisterWorkflow(wf)
}

// AddWorkflowStep adds a step to an existing workflow or creates a new workflow if it doesn't exist
func (s *ModularAPIService) AddWorkflowStep(workflowName string, step workflow.WorkflowStep) error {
	// Check if workflow exists
	existingWorkflow, exists := s.GetWorkflow(workflowName)

	if !exists {
		// Create a new workflow with this step
		newWorkflow := workflow.Workflow{
			Name:  workflowName,
			Steps: []workflow.WorkflowStep{step},
		}
		return s.RegisterWorkflow(newWorkflow)
	}

	// Add step to existing workflow
	existingWorkflow.Steps = append(existingWorkflow.Steps, step)
	return s.RegisterWorkflow(existingWorkflow)
}

// ExecuteWorkflow executes a workflow with the given parameters
// If result is not nil, the response from the last step will be unmarshaled into it
func (s *ModularAPIService) ExecuteWorkflow(name string, params map[string]interface{}, result interface{}) (map[string]interface{}, error) {
	return s.workflowExecutor.ExecuteWorkflow(name, params, result)
}

// GetWorkflow returns a workflow by name
func (s *ModularAPIService) GetWorkflow(name string) (workflow.Workflow, bool) {
	return s.workflowExecutor.GetWorkflow(name)
}

// ListWorkflows returns a list of all registered workflow names
func (s *ModularAPIService) ListWorkflows() []string {
	return s.workflowExecutor.ListWorkflows()
}

// SaveWorkflows saves all workflows to a file
func (s *ModularAPIService) SaveWorkflows(filepath string) error {
	return s.workflowExecutor.SaveWorkflows(filepath)
}

// LoadWorkflows loads workflows from a file
func (s *ModularAPIService) LoadWorkflows(filepath string) error {
	return s.workflowExecutor.LoadWorkflows(filepath)
}
