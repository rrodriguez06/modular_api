# Examples

This document provides practical examples of using the Modular API package for common scenarios.

## Basic API Request

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/romainrodriguez/modular_api/pkg/modularapi"
    "github.com/romainrodriguez/modular_api/pkg/modularapi/template"
)

func main() {
    // Create a new service builder
    builder := modularapi.NewServiceBuilder().
        WithTimeout(30 * time.Second).
        WithService("MyAPI", "https://api.example.com", "API_KEY").
        WithServiceHeaders("MyAPI", map[string]string{
            "Content-Type": "application/json",
            "Accept": "application/json",
        })

    // Add a template for getting user data
    builder.WithTemplate("MyAPI", "GetUser", *template.NewRouteTemplate(
        "GET", 
        "/users/{{user_id}}",
    ))

    // Build the service
    service := builder.Build()

    // Make the API request
    var result map[string]interface{}
    err := service.PerformRequest("MyAPI", "GetUser", map[string]interface{}{
        "user_id": "123",
    }, &result)
    
    if err != nil {
        log.Fatalf("Error making API request: %v", err)
    }
    
    fmt.Printf("User data: %+v\n", result)
}
```

## POST Request with Body

```go
// Add a template for creating a new user
builder.WithTemplate("MyAPI", "CreateUser", *template.NewRouteTemplate(
    "POST", 
    "/users",
).WithBody(map[string]interface{}{
    "name": "{{name}}",
    "email": "{{email}}",
    "age": "{{age?}}",
}))

// Make the API request
var result map[string]interface{}
err := service.PerformRequest("MyAPI", "CreateUser", map[string]interface{}{
    "name": "John Doe",
    "email": "john.doe@example.com",
    // age is optional, so we can omit it
}, &result)
```

## Basic Workflow Example

```go
// Define templates
builder.WithTemplate("API", "GetPatient", *template.NewRouteTemplate(
    "GET",
    "/patients/{{patient_id}}",
))

builder.WithTemplate("API", "GetUser", *template.NewRouteTemplate(
    "GET",
    "/users/{{user_id}}",
))

// Create workflow
patientStep := modularapi.NewWorkflowStepTemplate(
    "get_patient", 
    "Get patient details", 
    "API", 
    "GetPatient",
).WithParam("patient_id", "{{patient_id}}").
  WithResultMap("response.owner_user_id", "user_id")

userStep := modularapi.NewWorkflowStepTemplate(
    "get_user", 
    "Get user details", 
    "API", 
    "GetUser",
).WithDynamicParam("user_id", "user_id").
  WithCondition(workflow.ConditionExists, "user_id", nil)

builder.WithWorkflow("get_patient_and_doctor", "Get patient and their doctor").
    WithStep(patientStep).
    WithStep(userStep).
    Build()

// Execute workflow
var userResponse UserResponse
result, err := service.ExecuteWorkflow("get_patient_and_doctor", map[string]interface{}{
    "patient_id": "123456",
}, &userResponse)
```

## Parallel Workflow Steps

```go
// Define templates
builder.WithTemplate("API", "GetUser", *template.NewRouteTemplate(
    "GET", "/users/{{user_id}}",
))
builder.WithTemplate("API", "GetPosts", *template.NewRouteTemplate(
    "GET", "/users/{{user_id}}/posts",
))
builder.WithTemplate("API", "GetFollowers", *template.NewRouteTemplate(
    "GET", "/users/{{user_id}}/followers",
))

// Create workflow with parallel steps
builder.WithWorkflow("user_dashboard", "Get user dashboard data").
    WithStep(
        modularapi.NewWorkflowStepTemplate("get_user", "Get user details", "API", "GetUser").
            WithParam("user_id", "{{user_id}}"),
    ).
    WithStep(
        modularapi.NewWorkflowStepTemplate("get_posts", "Get user posts", "API", "GetPosts").
            WithParam("user_id", "{{user_id}}").
            WithParallelWith("get_followers"), // Execute in parallel with get_followers
    ).
    WithStep(
        modularapi.NewWorkflowStepTemplate("get_followers", "Get user followers", "API", "GetFollowers").
            WithParam("user_id", "{{user_id}}"),
    ).
    Build()

// Execute the workflow
result, err := service.ExecuteWorkflow("user_dashboard", map[string]interface{}{
    "user_id": "123",
}, nil)
```

## Streaming Response Example

```go
// Create an HTTP handler that streams API responses
http.HandleFunc("/stream-weather", func(w http.ResponseWriter, r *http.Request) {
    // Set headers for streaming response
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // Stream the API response directly to the client
    _, err := service.PerformStreamingRequest("WeatherAPI", "StreamWeather", map[string]interface{}{
        "location": "New York",
        "interval": "1m",
    }, w)
    
    if err != nil {
        log.Printf("Error streaming weather data: %v", err)
    }
})
```

## Result Parameter Example with Typed Response

```go
// Define a type for the user response
type UserResponse struct {
    Response struct {
        ID     string `json:"id"`
        Name   string `json:"name"`
        Email  string `json:"email"`
        Phone  string `json:"phone"`
    } `json:"response"`
}

// Execute workflow with typed result
var userResponse UserResponse
result, err := service.ExecuteWorkflow("get_user_workflow", map[string]interface{}{
    "user_id": "123",
}, &userResponse)

if err != nil {
    log.Fatalf("Error executing workflow: %v", err)
}

// Access the typed response
fmt.Printf("User ID: %s\n", userResponse.Response.ID)
fmt.Printf("User Name: %s\n", userResponse.Response.Name)
fmt.Printf("User Email: %s\n", userResponse.Response.Email)
```
