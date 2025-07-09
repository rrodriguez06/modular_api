package workflow_test

import (
	"testing"

	"github.com/rrodriguez06/modular_api/pkg/modularapi/workflow"
)

func TestWorkflowLoopAndAggregator(t *testing.T) {
	// Create mock API service
	mockService := NewMockAPIService()

	// Add mock responses
	mockService.AddMockResponse("users", "get", map[string]interface{}{
		"id":    "user123",
		"name":  "John Doe",
		"email": "john@example.com",
	})

	mockService.AddMockResponse("users", "getItems", map[string]interface{}{
		"user_id": "user123",
		"items": []interface{}{
			"item1",
			"item2",
			"item3",
		},
	})

	mockService.AddMockResponse("items", "getDetails", map[string]interface{}{
		"id":    "{{_params.item_id}}",
		"name":  "Item {{_params.item_id}}",
		"price": 10.99,
	})

	// Create workflow executor
	executor := workflow.NewWorkflowExecutor(mockService)

	// Create test workflow with loop
	loopWorkflow := workflow.Workflow{
		Name:        "loop_workflow",
		Description: "Test workflow with loop",
		Steps: []workflow.WorkflowStep{
			{
				ID:          "get_user",
				Description: "Get user details",
				ServiceName: "users",
				ActionName:  "get",
				Parameters: map[string]interface{}{
					"id": "{{user_id}}",
				},
				ResultMapping: map[string]string{
					"id":    "user_id_result",
					"name":  "user_name",
					"email": "user_email",
				},
			},
			{
				ID:          "get_items",
				Description: "Get user items",
				ServiceName: "users",
				ActionName:  "getItems",
				DynamicParams: map[string]string{
					"user_id": "user_id_result",
				},
				ResultMapping: map[string]string{
					"items": "item_ids",
				},
			},
			{
				ID:          "get_item_details",
				Description: "Get details for each item",
				ServiceName: "items",
				ActionName:  "getDetails",
				DynamicParams: map[string]string{
					"item_id": "current_item",
				},
				ResultMapping: map[string]string{
					"name": "item_details",
				},
				LoopOver: "item_ids",
				LoopAs:   "current_item",
			},
		},
		// Define an aggregator for the workflow
		Aggregator: map[string]string{
			"user":       "user_id_result",
			"user_name":  "user_name",
			"items":      "item_details",
			"item_count": "item_details.length",
		},
	}

	// Register workflow
	err := executor.RegisterWorkflow(loopWorkflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	// Execute workflow
	var aggregatedResult map[string]interface{}
	workflowVars, err := executor.ExecuteWorkflow("loop_workflow", map[string]interface{}{
		"user_id": "user123",
	}, &aggregatedResult)

	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	// Test loop execution results
	itemDetails, ok := workflowVars["item_details"]
	if !ok {
		t.Errorf("Expected item_details to be present in workflow variables")
	}

	// Should be an array from loop step
	itemDetailsArray, ok := itemDetails.([]interface{})
	if !ok {
		t.Errorf("Expected item_details to be an array, got %T", itemDetails)
	} else if len(itemDetailsArray) != 3 {
		t.Errorf("Expected 3 items in item_details, got %d", len(itemDetailsArray))
	}

	// Test aggregator results
	if aggregatedResult["user"] != "user123" {
		t.Errorf("Expected aggregated user to be 'user123', got %v", aggregatedResult["user"])
	}

	if aggregatedResult["user_name"] != "John Doe" {
		t.Errorf("Expected aggregated user_name to be 'John Doe', got %v", aggregatedResult["user_name"])
	}

	if aggregatedResult["item_count"] != float64(3) {
		t.Errorf("Expected aggregated item_count to be 3, got %v", aggregatedResult["item_count"])
	}

	// Check items array in aggregated result
	items, ok := aggregatedResult["items"].([]interface{})
	if !ok {
		t.Errorf("Expected aggregated items to be an array, got %T", aggregatedResult["items"])
	} else if len(items) != 3 {
		t.Errorf("Expected 3 items in aggregated items, got %d", len(items))
	}
}
