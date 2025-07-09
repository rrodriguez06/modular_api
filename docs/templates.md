# Templates

Templates are a core concept in the Modular API package. They define how API requests are constructed and how parameters are handled.

## Template Basics

A template in the Modular API package represents a specific API endpoint with:

- An HTTP method (GET, POST, PUT, DELETE, etc.)
- A URL path pattern with parameter placeholders
- Optional query parameters
- Optional body parameters
- Optional headers

## Creating Templates

Templates are created using the `template.NewRouteTemplate` function:

```go
// Create a simple GET template
getTemplate := template.NewRouteTemplate(
    "GET",
    "/api/{{version}}/users/{{user_id}}",
)

// Create a POST template with a body
postTemplate := template.NewRouteTemplate(
    "POST",
    "/api/{{version}}/users",
).WithBody(map[string]interface{}{
    "name": "{{name}}",
    "email": "{{email}}",
    "age": "{{age?}}",  // Optional parameter
})

// Create a template with query parameters
searchTemplate := template.NewRouteTemplate(
    "GET",
    "/api/{{version}}/search",
).WithQueryParams(map[string]string{
    "q": "{{query}}",
    "limit": "{{limit?}}",  // Optional parameter
    "offset": "{{offset?}}",  // Optional parameter
})
```

## Parameter Syntax

Templates use a simple syntax for parameters:

- Required parameters: `{{parameter_name}}`
- Optional parameters: `{{parameter_name?}}`

Parameters can appear in:

- The URL path
- Query parameters
- Body parameters
- Headers

## Optional Parameters

Optional parameters are marked with a `?` suffix. If an optional parameter is not provided, it will be:

- Omitted from the URL path
- Omitted from query parameters
- Omitted from the request body

## Adding Templates to a Service

Templates are added to a service using the `WithTemplate` method of the service builder:

```go
builder.WithTemplate("MyAPI", "GetUser", *template.NewRouteTemplate(
    "GET", 
    "/api/{{version}}/users/{{user_id}}/",
))
```

Here, "MyAPI" is the service name and "GetUser" is the template name.

## Saving and Loading Templates

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

## Template Expansion

When you make an API request using a template, the template parameters are expanded using the provided parameter values:

```go
// The template: /api/{{version}}/users/{{user_id}}/
// With parameters: {"version": "v1", "user_id": "123"}
// Expands to: /api/v1/users/123/
```

Parameter validation ensures that all required parameters are provided before making the request.
