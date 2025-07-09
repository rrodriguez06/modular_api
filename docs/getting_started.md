# Getting Started with Modular API

This guide will help you set up and use the Modular API package in your Go projects.

## Installation

Install the Modular API package using the Go package manager:

```bash
go get github.com/romainrodriguez/modular_api
```

## Basic Usage

The Modular API package is designed to simplify API interactions through a template-based approach. Here's a simple example:

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
        WithService("MyAPI", "https://api.example.com", "YOUR_API_TOKEN").
        WithServiceDefaultParams("MyAPI", map[string]interface{}{
            "version": "v1",
        })

    // Add a template
    builder.WithTemplate("MyAPI", "GetUser", *template.NewRouteTemplate(
        "GET", 
        "/api/{{version}}/users/{{user_id}}/",
    ))

    // Build the service
    service := builder.Build()

    // Use the service to make an API call
    var result map[string]interface{}
    err := service.PerformRequest("MyAPI", "GetUser", map[string]interface{}{
        "user_id": "123",
    }, &result)
    
    if err != nil {
        log.Fatalf("Error making API request: %v", err)
    }
    
    fmt.Printf("API Result: %+v\n", result)
}
```

## Next Steps

Once you're comfortable with the basics, you can explore:

- [Templates](templates.md) - Learn how to define and use API templates
- [Services](services.md) - Configure services with headers, parameters, and more
- [Workflows](workflows.md) - Chain multiple API calls with interdependent parameters
- [Advanced Features](advanced_features.md) - Explore streaming requests, parameter validation, and more
