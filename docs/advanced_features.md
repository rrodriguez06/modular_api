# Advanced Features

This document covers advanced features of the Modular API package that can help you build more sophisticated API interactions.

## Streaming Requests

For APIs that return streaming data, use the `PerformStreamingRequest` method:

```go
http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
    _, err := service.PerformStreamingRequest("MyAPI", "StreamData", map[string]interface{}{
        "user_id": "123",
    }, w)
    
    if err != nil {
        http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
    }
})
```

This method is useful for handling:

- Large data responses that shouldn't be loaded entirely into memory
- Server-sent events (SSE)
- Real-time data streams

## Parameter Substitution

The Modular API package supports template-based parameter substitution in various places:

### In URL Paths

```go
template.NewRouteTemplate("GET", "/api/{{version}}/users/{{user_id}}")
```

### In Query Parameters

```go
template.WithQueryParams(map[string]string{
    "query": "{{search_term}}",
    "limit": "{{page_size}}",
})
```

### In Request Bodies

```go
template.WithBody(map[string]interface{}{
    "name": "{{name}}",
    "email": "{{email}}",
    "preferences": map[string]interface{}{
        "language": "{{language}}",
        "theme": "{{theme}}",
    },
})
```

### In Workflow Step Parameters

```go
step.WithParam("user_id", "{{user_id}}")
```

## Expression Handling

In workflows, you can use expressions to extract values from complex responses:

```go
// Extract a nested value using dot notation
step.WithResultMap("response.data.user.id", "user_id")

// Use in dynamic parameters
step.WithDynamicParam("owner_id", "user.owner_id")
```

## Error Handling

Workflows provide several error handling strategies:

```go
// Continue to the next step even if this step fails
step.WithErrorHandling(workflow.ContinueOnError)

// Abort the workflow if this step fails (default)
step.WithErrorHandling(workflow.AbortOnError)

// Retry the step if it fails (with configurable retry count and delay)
step.WithErrorHandling(workflow.RetryOnError).
    WithMaxRetries(3).
    WithRetryDelay(500) // milliseconds
```

## Typed Responses

While you can use generic `map[string]interface{}` for API responses, you can also define typed structures:

```go
type UserResponse struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    CreatedAt string `json:"created_at"`
}

var user UserResponse
err := service.PerformRequest("MyAPI", "GetUser", map[string]interface{}{
    "user_id": "123",
}, &user)
```

The same works for workflow results:

```go
var user UserResponse
result, err := service.ExecuteWorkflow("get_user", map[string]interface{}{
    "user_id": "123",
}, &user)
```

## Template Persistence

You can save and load templates to/from JSON files:

```go
// Save templates to file
err := service.SaveTemplates("templates.json")

// Load templates from file
err = service.LoadTemplates("templates.json")
```

## Workflow Persistence

Similarly, you can save and load workflows:

```go
// Save workflows to file
err := service.GetWorkflowService().SaveWorkflows("workflows.json")

// Load workflows from file
err = service.GetWorkflowService().LoadWorkflows("workflows.json")
```

This is useful for creating workflows at runtime or storing workflows configured by users.
