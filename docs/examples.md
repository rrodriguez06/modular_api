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

## Loop Workflow Example

This example demonstrates how to use the loop functionality in a workflow:

```go
// Define templates
builder.WithTemplate("API", "GetUser", *template.NewRouteTemplate(
    "GET", "/users/{{user_id}}",
))
builder.WithTemplate("API", "GetUserProjects", *template.NewRouteTemplate(
    "GET", "/users/{{user_id}}/projects",
))
builder.WithTemplate("API", "GetProject", *template.NewRouteTemplate(
    "GET", "/projects/{{project_id}}",
))

// Step 1: Get user details
userStep := modularapi.NewWorkflowStepTemplate("get_user", "Get user details", "API", "GetUser").
    WithParam("user_id", "{{user_id}}").
    WithResultMap("response", "user_data")

// Step 2: Get list of project IDs for the user
projectsStep := modularapi.NewWorkflowStepTemplate("get_projects", "Get project IDs", "API", "GetUserProjects").
    WithParam("user_id", "{{user_id}}").
    WithResultMap("response.project_ids", "project_id_list")

// Step 3: Loop over each project ID to get details
projectDetailsStep := modularapi.NewWorkflowStepTemplate("get_project_details", "Get project details", "API", "GetProject").
    WithDynamicParam("project_id", "current_project").           // Use current item in the loop
    WithLoopOver("project_id_list", "current_project").          // Configure as a loop step
    WithResultMap("response", "project_details_collection")      // Results collected into an array

// Create the workflow with all steps
builder.WithWorkflow("user_projects", "Get all projects for a user").
    WithStep(userStep).
    WithStep(projectsStep).
    WithStep(projectDetailsStep).
    Build()

// Execute the workflow
var result map[string]interface{}
workflowVars, err := service.ExecuteWorkflow("user_projects", map[string]interface{}{
    "user_id": "123",
}, &result)

if err != nil {
    log.Fatalf("Error executing workflow: %v", err)
}

// Access the collected results from all loop iterations
projectDetails := workflowVars["project_details_collection"].([]interface{})
fmt.Printf("Found %d projects for user\n", len(projectDetails))
```

## Workflow with Aggregator Example

This example demonstrates how to use the aggregator functionality to structure workflow results:

```go
// Using the same workflow steps as in the loop example, but adding an aggregator
builder.WithWorkflow("user_with_projects", "Get user with all projects").
    WithStep(userStep).
    WithStep(projectsStep).
    WithStep(projectDetailsStep).
    WithAggregator(map[string]string{
        // Define the structure of the final result
        "user": "user_data",                               // Include user data
        "projects": "project_details_collection",          // Include all project details
        "project_count": "project_details_collection.length", // Count of projects
        "user_id": "input.user_id",                        // Include original input
    }).
    Build()

// Define a type for the structured result
type DashboardResult struct {
    User         map[string]interface{}   `json:"user"`
    Projects     []map[string]interface{} `json:"projects"`
    ProjectCount int                      `json:"project_count"`
    UserID       string                   `json:"user_id"`
}

// Execute the workflow with the structured result
var dashboardResult DashboardResult
_, err := service.ExecuteWorkflow("user_with_projects", map[string]interface{}{
    "user_id": "123",
}, &dashboardResult)

if err != nil {
    log.Fatalf("Error executing workflow: %v", err)
}

// Access the structured data
fmt.Printf("User: %s\n", dashboardResult.User["name"])
fmt.Printf("Project Count: %d\n", dashboardResult.ProjectCount)
fmt.Printf("First Project Name: %s\n", dashboardResult.Projects[0]["name"])
```
