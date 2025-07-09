# Modular API

A flexible and extensible Go package for handling API requests with modular templating and workflow capabilities.

## Features

- **Template-based API requests**: Define your API endpoints once, use them everywhere
- **Parameter validation**: Automatic validation of required and optional parameters
- **Multiple request types**: Support for standard requests and streaming requests
- **Service-level configuration**: Configure headers, parameters, and other settings at the service level
- **Template persistence**: Save and load templates from JSON files
- **Flexible parameter handling**: Support for path parameters, query parameters, and body parameters
- **Optional parameters**: Mark parameters as optional with the `?` suffix
- **Workflows**: Chain multiple API calls with dependencies between them
- **Parallel execution**: Execute workflow steps in parallel
- **Conditional execution**: Execute workflow steps based on conditions
- **Result mapping**: Map response fields to variables for use in subsequent steps

## Installation

```bash
go get github.com/romainrodriguez/modular_api
```

## Quick Start

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

## Documentation

For more detailed documentation, please refer to the [documentation index](docs/index.md) or the following guides:

- [Getting Started](docs/getting_started.md) - Basic setup and usage
- [Templates](docs/templates.md) - Working with API templates
- [Services](docs/services.md) - Configuring and using services
- [Workflows](docs/workflows.md) - Creating and executing workflows
- [Advanced Features](docs/advanced_features.md) - Advanced features and techniques
- [Examples](docs/examples.md) - Practical examples of common use cases

## License

MIT License
