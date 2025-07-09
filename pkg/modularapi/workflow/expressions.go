package workflow

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// extractValue extracts a value from a nested map using dot notation
// e.g. "user.profile.name" would extract data["user"]["profile"]["name"]
func extractValue(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")

	// Start with the root object
	var current interface{} = data

	// Traverse the path
	for i, part := range parts {
		// Handle array indexing if the part is like "items[0]"
		indexMatch := regexp.MustCompile(`^(.*?)\[(\d+)\]$`).FindStringSubmatch(part)
		if indexMatch != nil {
			// We have an array index
			fieldName := indexMatch[1]
			indexStr := indexMatch[2]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, false
			}

			// First get the field value
			fieldMap, ok := current.(map[string]interface{})
			if !ok {
				log.Printf("Failed to access array field %s: parent is not a map but %T", fieldName, current)
				return nil, false
			}

			arrayField, exists := fieldMap[fieldName]
			if !exists {
				log.Printf("Array field %s not found in map", fieldName)
				return nil, false
			}

			// Then get the array element
			arrayValue, ok := arrayField.([]interface{})
			if !ok {
				log.Printf("Field %s is not an array but %T", fieldName, arrayField)
				return nil, false
			}

			if index < 0 || index >= len(arrayValue) {
				log.Printf("Array index %d is out of bounds for array of length %d", index, len(arrayValue))
				return nil, false
			}

			current = arrayValue[index]
		} else {
			// Regular field access
			currentMap, ok := current.(map[string]interface{})
			if !ok {
				// For debugging, print the current path we're trying to access
				accessedPath := strings.Join(parts[:i], ".")
				log.Printf("Failed to access field %s: parent path %s is not a map but %T",
					part, accessedPath, current)
				return nil, false
			}

			value, exists := currentMap[part]
			if !exists {
				log.Printf("Field %s not found in map with keys: %v", part, getMapKeys(currentMap))
				return nil, false
			}

			current = value
		}
	}

	return current, true
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// evaluateCondition checks if a condition is met based on the variables
func evaluateCondition(condition *StepCondition, variables map[string]interface{}) (bool, error) {
	if condition == nil {
		return true, nil
	}

	// Get the source value
	sourceValue, exists := variables[condition.SourceVariable]

	// For exists condition, we only need to check if the variable exists
	if condition.Type == ConditionExists {
		return exists && sourceValue != nil, nil
	}

	// For other conditions, we need both the source variable and the comparison value
	if !exists {
		return false, nil
	}

	// Evaluate based on condition type
	switch condition.Type {
	case ConditionEquals:
		return reflect.DeepEqual(sourceValue, condition.Value), nil

	case ConditionContains:
		return evaluateContains(sourceValue, condition.Value)

	case ConditionGreaterThan:
		return evaluateGreaterThan(sourceValue, condition.Value)

	case ConditionLessThan:
		return evaluateLessThan(sourceValue, condition.Value)

	default:
		return false, fmt.Errorf("unsupported condition type: %s", condition.Type)
	}
}

// evaluateContains checks if a value contains another value (for strings, slices, maps)
func evaluateContains(source, target interface{}) (bool, error) {
	// Handle string contains
	sourceStr, isSourceStr := source.(string)
	targetStr, isTargetStr := target.(string)
	if isSourceStr && isTargetStr {
		return strings.Contains(sourceStr, targetStr), nil
	}

	// Handle slice contains
	sourceVal := reflect.ValueOf(source)
	if sourceVal.Kind() == reflect.Slice || sourceVal.Kind() == reflect.Array {
		for i := 0; i < sourceVal.Len(); i++ {
			if reflect.DeepEqual(sourceVal.Index(i).Interface(), target) {
				return true, nil
			}
		}
		return false, nil
	}

	// Handle map contains (checks keys)
	if sourceVal.Kind() == reflect.Map {
		for _, key := range sourceVal.MapKeys() {
			if reflect.DeepEqual(key.Interface(), target) {
				return true, nil
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("contains condition not supported for type %T", source)
}

// evaluateGreaterThan checks if a value is greater than another value
func evaluateGreaterThan(source, target interface{}) (bool, error) {
	// Convert to float64 for numeric comparison
	sourceFloat, sourceErr := toFloat64(source)
	targetFloat, targetErr := toFloat64(target)

	if sourceErr == nil && targetErr == nil {
		return sourceFloat > targetFloat, nil
	}

	// Compare strings lexicographically
	sourceStr, isSourceStr := source.(string)
	targetStr, isTargetStr := target.(string)
	if isSourceStr && isTargetStr {
		return sourceStr > targetStr, nil
	}

	return false, fmt.Errorf("greater than condition not supported for types %T and %T", source, target)
}

// evaluateLessThan checks if a value is less than another value
func evaluateLessThan(source, target interface{}) (bool, error) {
	// Convert to float64 for numeric comparison
	sourceFloat, sourceErr := toFloat64(source)
	targetFloat, targetErr := toFloat64(target)

	if sourceErr == nil && targetErr == nil {
		return sourceFloat < targetFloat, nil
	}

	// Compare strings lexicographically
	sourceStr, isSourceStr := source.(string)
	targetStr, isTargetStr := target.(string)
	if isSourceStr && isTargetStr {
		return sourceStr < targetStr, nil
	}

	return false, fmt.Errorf("less than condition not supported for types %T and %T", source, target)
}

// toFloat64 converts various types to float64 for comparison
func toFloat64(v interface{}) (float64, error) {
	switch value := v.(type) {
	case float64:
		return value, nil
	case float32:
		return float64(value), nil
	case int:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case uint:
		return float64(value), nil
	case uint64:
		return float64(value), nil
	case uint32:
		return float64(value), nil
	case string:
		return strconv.ParseFloat(value, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// expressionPattern is a simple regex to detect expressions
var expressionPattern = regexp.MustCompile(`\{\{(.+?)\}\}`)

// isExpression checks if a string is an expression
func isExpression(s string) bool {
	return expressionPattern.MatchString(s)
}

// evaluateExpression evaluates an expression and returns the result
// For now, this is a simple implementation that handles variable substitution
// In the future, this could be expanded to handle more complex expressions
func evaluateExpression(expr string, variables map[string]interface{}) (interface{}, error) {
	// Simple variable substitution
	matches := expressionPattern.FindAllStringSubmatch(expr, -1)
	if len(matches) == 0 {
		return expr, nil
	}

	// If the entire string is an expression like "{{variable}}"
	if len(matches) == 1 && matches[0][0] == expr {
		varName := matches[0][1]

		// Check for ternary operation
		if strings.Contains(varName, "?") {
			return evaluateTernary(varName, variables)
		}

		// Direct variable reference
		if value, exists := variables[varName]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("variable %s not found", varName)
	}

	// Handle multiple expressions within a string
	result := expr
	for _, match := range matches {
		fullMatch := match[0]
		varName := match[1]

		// Get the variable value
		var replaceValue string
		if value, exists := variables[varName]; exists {
			replaceValue = fmt.Sprintf("%v", value)
		} else {
			return nil, fmt.Errorf("variable %s not found", varName)
		}

		// Replace in the result
		result = strings.Replace(result, fullMatch, replaceValue, 1)
	}

	return result, nil
}

// evaluateTernary handles simple ternary operations like "condition ? trueValue : falseValue"
func evaluateTernary(expr string, variables map[string]interface{}) (interface{}, error) {
	parts := strings.Split(expr, "?")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ternary expression: %s", expr)
	}

	condition := strings.TrimSpace(parts[0])
	choiceParts := strings.Split(parts[1], ":")
	if len(choiceParts) != 2 {
		return nil, fmt.Errorf("invalid ternary expression: %s", expr)
	}

	trueValue := strings.TrimSpace(choiceParts[0])
	falseValue := strings.TrimSpace(choiceParts[1])

	// Evaluate condition
	// Handle simple equality condition (a == b)
	if strings.Contains(condition, "==") {
		eqParts := strings.Split(condition, "==")
		if len(eqParts) != 2 {
			return nil, fmt.Errorf("invalid equality condition: %s", condition)
		}

		leftVal := getValueForExpression(strings.TrimSpace(eqParts[0]), variables)
		rightVal := getValueForExpression(strings.TrimSpace(eqParts[1]), variables)

		if reflect.DeepEqual(leftVal, rightVal) {
			return getValueForExpression(trueValue, variables), nil
		} else {
			return getValueForExpression(falseValue, variables), nil
		}
	}

	// Handle simple inequality condition (a != b)
	if strings.Contains(condition, "!=") {
		eqParts := strings.Split(condition, "!=")
		if len(eqParts) != 2 {
			return nil, fmt.Errorf("invalid inequality condition: %s", condition)
		}

		leftVal := getValueForExpression(strings.TrimSpace(eqParts[0]), variables)
		rightVal := getValueForExpression(strings.TrimSpace(eqParts[1]), variables)

		if !reflect.DeepEqual(leftVal, rightVal) {
			return getValueForExpression(trueValue, variables), nil
		} else {
			return getValueForExpression(falseValue, variables), nil
		}
	}

	// Handle simple variable check (just the variable name means check if it's truthy)
	condValue := getValueForExpression(condition, variables)
	if isTruthy(condValue) {
		return getValueForExpression(trueValue, variables), nil
	} else {
		return getValueForExpression(falseValue, variables), nil
	}
}

// getValueForExpression gets the value for a variable or literal expression
func getValueForExpression(expr string, variables map[string]interface{}) interface{} {
	// Check if it's a quoted string
	if (strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'")) ||
		(strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"")) {
		// Remove quotes
		return expr[1 : len(expr)-1]
	}

	// Check if it's a number
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return num
	}

	// Check if it's a boolean
	if expr == "true" {
		return true
	} else if expr == "false" {
		return false
	}

	// Check if it's a variable
	if value, exists := variables[expr]; exists {
		return value
	}

	// Default to nil
	return nil
}

// isTruthy checks if a value is truthy (not false, nil, zero, or empty)
func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}

	switch value := v.(type) {
	case bool:
		return value
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(value).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(value).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(value).Float() != 0
	case string:
		return value != ""
	case []interface{}:
		return len(value) > 0
	case map[string]interface{}:
		return len(value) > 0
	default:
		return true
	}
}
