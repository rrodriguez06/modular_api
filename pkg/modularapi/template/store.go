package template

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// TemplateStore manages a collection of route templates
type TemplateStore struct {
	templates map[string]map[string]RouteTemplate
}

// NewTemplateStore creates a new template store
func NewTemplateStore() *TemplateStore {
	return &TemplateStore{
		templates: make(map[string]map[string]RouteTemplate),
	}
}

// AddTemplate adds a route template for a specific service and action
func (ts *TemplateStore) AddTemplate(serviceName, action string, route RouteTemplate) {
	// Initialize the OptionalParams map if it doesn't exist
	if route.OptionalParams == nil {
		route.OptionalParams = make(map[string]bool)
	}

	// Extract path parameters from endpoint placeholders and identify optional params
	route.PathParams = extractPathParams(route.Endpoint)

	// Scan the template for optional parameters and populate the OptionalParams map
	scanTemplateForOptionalParams(&route)

	if ts.templates[serviceName] == nil {
		ts.templates[serviceName] = make(map[string]RouteTemplate)
	}
	ts.templates[serviceName][action] = route
}

// GetTemplate returns a route template for a specific service and action
func (ts *TemplateStore) GetTemplate(serviceName, action string) (RouteTemplate, bool) {
	if serviceTemplates, ok := ts.templates[serviceName]; ok {
		if template, ok := serviceTemplates[action]; ok {
			return template, true
		}
	}
	return RouteTemplate{}, false
}

// HasTemplate checks if a template exists for a specific service and action
func (ts *TemplateStore) HasTemplate(serviceName, action string) bool {
	if serviceTemplates, ok := ts.templates[serviceName]; ok {
		_, ok := serviceTemplates[action]
		return ok
	}
	return false
}

// SaveToFile saves all templates to a JSON file
func (ts *TemplateStore) SaveToFile(filepath string) error {
	data, err := json.MarshalIndent(ts.templates, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal templates: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write templates to file: %w", err)
	}

	return nil
}

// LoadFromFile loads templates from a JSON file and merges them with existing templates
func (ts *TemplateStore) LoadFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read templates file: %w", err)
	}

	templates := make(map[string]map[string]RouteTemplate)
	if err := json.Unmarshal(data, &templates); err != nil {
		return fmt.Errorf("failed to unmarshal templates: %w", err)
	}

	// Merge with existing templates
	for service, routes := range templates {
		if ts.templates[service] == nil {
			ts.templates[service] = make(map[string]RouteTemplate)
		}
		for action, template := range routes {
			// Ensure OptionalParams is initialized
			if template.OptionalParams == nil {
				template.OptionalParams = make(map[string]bool)
			}

			// Re-scan for optional parameters
			scanTemplateForOptionalParams(&template)

			// Update the template
			ts.templates[service][action] = template
		}
	}

	return nil
}

// extractPathParams extracts parameter names from placeholders in the endpoint
func extractPathParams(endpoint string) []string {
	var params []string
	parts := strings.Split(endpoint, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{{") && strings.HasSuffix(part, "}}") {
			param := strings.TrimPrefix(strings.TrimSuffix(part, "}}"), "{{")

			// If the parameter is marked as optional with ? suffix, remove the suffix
			param = strings.TrimSuffix(param, "?")

			params = append(params, param)
		}
	}
	return params
}

// scanTemplateForOptionalParams scans all parts of the template for optional parameters
func scanTemplateForOptionalParams(route *RouteTemplate) {
	// Scan endpoint for optional parameters (marked with {{param?}})
	scanEndpointForOptionalParams(route)

	// Scan body parameters
	if route.Body != nil {
		scanMapForOptionalParams(route.Body, route.OptionalParams)
	}

	// Scan query parameters
	if route.QueryParams != nil {
		scanMapForOptionalParams(route.QueryParams, route.OptionalParams)
	}
}

// scanEndpointForOptionalParams scans the endpoint URL for optional parameters
func scanEndpointForOptionalParams(route *RouteTemplate) {
	parts := strings.Split(route.Endpoint, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{{") && strings.HasSuffix(part, "}}") {
			paramWithBraces := strings.TrimPrefix(strings.TrimSuffix(part, "}}"), "{{")

			// Check if the parameter is marked as optional with ? suffix
			if strings.HasSuffix(paramWithBraces, "?") {
				// Extract the parameter name without the ? suffix
				paramName := strings.TrimSuffix(paramWithBraces, "?")
				// Mark as optional
				route.OptionalParams[paramName] = true
			}
		}
	}
}

// scanMapForOptionalParams recursively scans map values for optional parameters
func scanMapForOptionalParams(data map[string]interface{}, optionalParamsMap map[string]bool) {
	for _, value := range data {
		switch v := value.(type) {
		case string:
			if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
				paramWithBraces := strings.TrimPrefix(strings.TrimSuffix(v, "}}"), "{{")
				if strings.HasSuffix(paramWithBraces, "?") {
					// Extract parameter name without the ? suffix
					paramName := strings.TrimSuffix(paramWithBraces, "?")
					// Mark as optional
					optionalParamsMap[paramName] = true
				}
			}
		case map[string]interface{}:
			scanMapForOptionalParams(v, optionalParamsMap)
		case []interface{}:
			for _, item := range v {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					scanMapForOptionalParams(nestedMap, optionalParamsMap)
				}
			}
		}
	}
}
