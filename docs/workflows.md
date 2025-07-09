# Workflows

Workflows in the Modular API package allow you to chain multiple API requests together, using the results of one request as inputs to the next. This is especially useful for complex operations that require multiple API calls.

## Workflow Basics

A workflow consists of a series of steps, where each step is an API request. The results of each step can be stored in variables that can be used by later steps.

## Creating a Workflow

Workflows are created using the workflow builder in the service builder:

```go
// Create a workflow
builder.WithWorkflow("get_user_by_patient", "Get user associated with a patient").
    WithStep(
        modularapi.NewWorkflowStepTemplate("get_patient", "Get patient details", "API", "GetPatient").
            WithParam("patient_id", "{{patient_id}}").
            WithResultMap("response.owner_user_id", "user_id"),
    ).
    WithStep(
        modularapi.NewWorkflowStepTemplate("get_user", "Get user details", "API", "GetUser").
            WithDynamicParam("user_id", "user_id").
            WithCondition(workflow.ConditionExists, "user_id", nil),
    ).
    Build()
```

## Workflow Steps

Each step in a workflow is defined by:

- A unique ID
- A description
- The service name to use
- The template name to use
- Parameters for the template
- Result mappings to extract values from the response
- Optional conditions for executing the step

## Parameter Types

Workflows support two types of parameters:

1. **Static parameters** - Fixed values provided when the workflow is defined:

```go
WorkflowStep.WithParam("patient_id", "{{patient_id}}")
```

1. **Dynamic parameters** - Values sourced from variables created during workflow execution:

```go
WorkflowStep.WithDynamicParam("user_id", "user_id")
```

## Result Mapping

Result mapping allows you to extract values from a step's response and store them as variables for use in later steps:

```go
WorkflowStep.WithResultMap("response.owner_user_id", "user_id")
```

This extracts the `owner_user_id` field from the `response` object and stores it in a variable called `user_id`.

You can use dot notation to access nested fields:

```go
WorkflowStep.WithResultMap("response.data.user.id", "user_id")
```

## Conditional Steps

You can make a step execute conditionally based on the value of a variable:

```go
WorkflowStep.WithCondition(workflow.ConditionExists, "user_id", nil)
```

Available condition types:

- `ConditionExists` - Checks if a variable exists and is not nil
- `ConditionEquals` - Checks if a variable equals a value
- `ConditionContains` - Checks if a variable contains a value (string or slice)
- `ConditionGreaterThan` - Checks if a variable is greater than a value
- `ConditionLessThan` - Checks if a variable is less than a value

## Executing a Workflow

Workflows are executed using the `ExecuteWorkflow` method:

```go
result, err := service.ExecuteWorkflow("get_user_by_patient", map[string]interface{}{
    "patient_id": "123456",
}, nil)
```

The parameters are:

1. Workflow name - The name of the workflow to execute
2. Initial parameters - The parameters to pass to the workflow
3. Result object - Optional object to receive the result of the final step

## Working with Results

The `ExecuteWorkflow` method returns two values:

1. A map of all variables created during workflow execution
2. An error if something went wrong

You can also pass a third parameter to receive the result of the final step:

```go
var userResponse UserResponse
result, err := service.ExecuteWorkflow("get_user_by_patient", map[string]interface{}{
    "patient_id": "123456",
}, &userResponse)
```

### Result Usage Patterns

There are three main ways to use the `ExecuteWorkflow` result parameter:

1. **Without Result Parameter** - Just get the workflow variables:

```go
result, err := service.ExecuteWorkflow("my_workflow", initialParams, nil)
```

1. **With Generic Map Result** - Get both workflow variables and the response as a generic map:

```go
var genericResponse map[string]interface{}
result, err := service.ExecuteWorkflow("my_workflow", initialParams, &genericResponse)
```

1. **With Typed Struct Result** - Get both workflow variables and the response as a typed struct:

```go
var typedResponse UserResponse
result, err := service.ExecuteWorkflow("my_workflow", initialParams, &typedResponse)
```

## Parallel Execution

Workflows can execute steps in parallel using the `ParallelWith` field:

```go
builder.WithWorkflow("user_dashboard", "Get user dashboard data").
    WithStep(
        modularapi.NewWorkflowStepTemplate("get_user", "Get user details", "API", "GetUser").
            WithParam("user_id", "{{user_id}}"),
    ).
    WithStep(
        modularapi.NewWorkflowStepTemplate("get_user_posts", "Get user posts", "API", "GetUserPosts").
            WithParam("user_id", "{{user_id}}").
            WithParallelWith("get_user_followers"),
    ).
    WithStep(
        modularapi.NewWorkflowStepTemplate("get_user_followers", "Get user followers", "API", "GetUserFollowers").
            WithParam("user_id", "{{user_id}}"),
    ).
    Build()
```

In this example, the `get_user_posts` and `get_user_followers` steps will execute in parallel after the `get_user` step completes.
