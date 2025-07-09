# Services

Services in the Modular API package represent external APIs that you interact with. This document explains how to configure and use services.

## Creating Services

Services are created using the service builder:

```go
builder := modularapi.NewServiceBuilder().
    WithTimeout(30 * time.Second).
    WithService("MyAPI", "https://api.example.com", "YOUR_API_TOKEN")
```

The `WithService` method takes three parameters:

1. Service name - A unique identifier for the service
2. Base URL - The base URL of the API
3. API key (optional) - An API key to authenticate requests

## Service Configuration

### Headers

You can set headers that will be applied to all requests to a service:

```go
builder.WithServiceHeaders("MyAPI", map[string]string{
    "Content-Type": "application/json",
    "Accept": "application/json",
    "X-API-Version": "1.0",
})
```

Or after the service has been built:

```go
service.SetServiceHeaders("MyAPI", map[string]string{
    "Content-Type": "application/json",
    "Accept": "application/json",
})
```

### Default Parameters

You can set default parameters that will be applied to all requests to a service:

```go
builder.WithServiceDefaultParams("MyAPI", map[string]interface{}{
    "version": "v1",
    "language": "en",
})
```

Or after the service has been built:

```go
service.SetServiceParams("MyAPI", map[string]interface{}{
    "version": "v1",
    "language": "en",
})
```

### Timeout

You can set a timeout for all requests:

```go
builder.WithTimeout(30 * time.Second)
```

## Making API Requests

Once a service is configured, you can make requests using the templates you've defined:

```go
var result map[string]interface{}
err := service.PerformRequest("MyAPI", "GetUser", map[string]interface{}{
    "user_id": "123",
}, &result)
```

The parameters are:

1. Service name - The name of the service to use
2. Template name - The name of the template to use
3. Parameters - The parameters to apply to the template
4. Result - A pointer to where the result should be stored

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

This is useful for APIs that return large amounts of data or for real-time data feeds.
