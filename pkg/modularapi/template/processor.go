package template

import (
	"reflect"
	"strings"
)

// processTemplateValue processes a template value, replacing any placeholders with actual values
func ProcessTemplateValue(value interface{}, params map[string]interface{}, optionalParams map[string]bool) (interface{}, bool) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
			// Extract parameter name and check if it's optional
			paramWithBraces := strings.TrimPrefix(strings.TrimSuffix(v, "}}"), "{{")
			isOptional := strings.HasSuffix(paramWithBraces, "?")

			// Get clean parameter name (without ? suffix if present)
			paramName := paramWithBraces
			if isOptional {
				paramName = strings.TrimSuffix(paramWithBraces, "?")
			}

			// Check if the parameter is in the params map
			if paramValue, exists := params[paramName]; exists {
				// For empty string or nil values in optional params, treat as not provided
				if (paramValue == "" || paramValue == nil) && (isOptional || optionalParams[paramName]) {
					return nil, false
				}

				// Handle arrays properly to prevent double encoding
				switch typedValue := paramValue.(type) {
				case []string:
					// Convert []string to []interface{} to ensure proper JSON marshaling
					result := make([]interface{}, len(typedValue))
					for i, s := range typedValue {
						result[i] = s
					}
					return result, true
				case []interface{}:
					// Already an []interface{}, just return it directly
					return typedValue, true
				case []int, []int64, []float64, []bool:
					// For other array types, use reflection to convert to []interface{}
					v := reflect.ValueOf(typedValue)
					if v.Kind() != reflect.Slice {
						// This should never happen since we're in a type switch
						return paramValue, true
					}

					length := v.Len()
					result := make([]interface{}, length)
					for i := 0; i < length; i++ {
						result[i] = v.Index(i).Interface()
					}
					return result, true
				default:
					return paramValue, true
				}
			}

			// If parameter is not found but is optional, return false to indicate it should be omitted
			if isOptional || optionalParams[paramName] {
				return nil, false
			}

			// Required parameter not found
			return nil, false
		}
		return v, true
	case map[string]interface{}:
		processed := make(map[string]interface{})
		for key, val := range v {
			if processedVal, valid := ProcessTemplateValue(val, params, optionalParams); valid {
				processed[key] = processedVal
			}
		}
		return processed, len(processed) > 0
	case []interface{}:
		processed := make([]interface{}, 0, len(v))
		for _, val := range v {
			if processedVal, valid := ProcessTemplateValue(val, params, optionalParams); valid {
				processed = append(processed, processedVal)
			}
		}
		return processed, len(processed) > 0
	default:
		return v, true
	}
}
