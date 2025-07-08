# Modular API

A flexible and extensible Go package for handling API requests with modular templating.

## Features

- **Template-based API requests**: Define your API endpoints once, use them everywhere
- **Parameter validation**: Automatic validation of required and optional parameters
- **Multiple request types**: Support for standard requests and streaming requests
- **Service-level configuration**: Configure headers, parameters, and other settings at the service level
- **Template persistence**: Save and load templates from JSON files
- **Flexible parameter handling**: Support for path parameters, query parameters, and body parameters
- **Optional parameters**: Mark parameters as optional with the `?` suffix
- **Clean separation of concerns**: Each component is separated into its own package for better maintainability

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
        WithTimeout(180 * time.Second).
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

## Template Syntax

Templates are defined using a simple syntax:

- **Path parameters**: `{{parameter_name}}` for required parameters, `{{parameter_name?}}` for optional parameters
- **Query parameters**: Define in the `QueryParams` map
- **Body parameters**: Define in the `Body` map

Example template:

```go
template.NewRouteTemplate(
    "POST",
    "/api/{{version}}/users/",
).WithBody(map[string]interface{}{
    "name": "{{name}}",
    "email": "{{email}}",
    "age": "{{age?}}", // Optional parameter
    "address": "{{address?}}", // Optional parameter
})
```

## Loading and Saving Templates

You can save templates to a JSON file and load them later:

```go
// Save templates to file
err := service.SaveTemplates("templates.json")
if err != nil {
    log.Fatalf("Error saving templates: %v", err)
}

// Load templates from file
err = service.LoadTemplates("templates.json")
if err != nil {
    log.Fatalf("Error loading templates: %v", err)
}
```

## Service-level Configuration

You can configure headers and parameters at the service level:

```go
// Set headers for a service
service.SetServiceHeaders("MyAPI", map[string]string{
    "Content-Type": "application/json",
    "X-API-Version": "1.0",
})

// Set parameters for a service
service.SetServiceParams("MyAPI", map[string]interface{}{
    "version": "v1",
    "language": "en",
})
```

## Streaming Requests

Handle streaming requests with the `PerformStreamingRequest` method:

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

## Future Enhancements

- API workflows: Chain multiple API calls with the results of previous calls
- Custom request/response processors
- Authentication mechanisms
- Rate limiting and retries
- Testing utilities

## License

MIT License
