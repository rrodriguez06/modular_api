package modularapi

import (
	"encoding/json"
	"log"
)

// ProcessResponse is a helper function for the workflow executor to process responses
func (s *ModularAPIService) ProcessResponse(response json.RawMessage, result interface{}) error {
	return json.Unmarshal(response, result)
}

// Implement the workflow.APIServiceExecutor interface for the ModularAPIService
func (s *ModularAPIService) ExecuteServiceAction(serviceName, actionName string, params map[string]interface{}, result interface{}) error {
	// Convert any string parameters that look like they should be template values
	// This fixes the issue where workflow parameters aren't properly processed as templates
	processedParams := make(map[string]interface{})

	// Copy all parameters first
	for k, v := range params {
		processedParams[k] = v
	}

	// Log the parameters we're using for debugging
	log.Printf("Executing service action: %s.%s with params: %+v", serviceName, actionName, processedParams)

	// Use our standard PerformRequest method, but with a compatibility wrapper
	// for the workflow executor which expects serviceName and actionName separately
	return s.PerformRequest(serviceName, actionName, processedParams, result)
}

// ExecuteServiceActionWithOptions is an extended version that allows passing request options
func (s *ModularAPIService) ExecuteServiceActionWithOptions(serviceName, actionName string, params map[string]interface{}, result interface{}, opts ...RequestOption) error {
	// Convert any string parameters that look like they should be template values
	// This fixes the issue where workflow parameters aren't properly processed as templates
	processedParams := make(map[string]interface{})

	// Copy all parameters first
	for k, v := range params {
		processedParams[k] = v
	}

	// Log the parameters we're using for debugging
	log.Printf("Executing service action with options: %s.%s with params: %+v", serviceName, actionName, processedParams)

	// Use our standard PerformRequest method with options
	return s.PerformRequest(serviceName, actionName, processedParams, result, opts...)
}
