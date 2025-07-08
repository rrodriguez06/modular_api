package modularapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rrodriguez06/modular_api/pkg/modularapi"
	"github.com/rrodriguez06/modular_api/pkg/modularapi/config"
	"github.com/rrodriguez06/modular_api/pkg/modularapi/template"
)

func TestModularAPIService(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the request has the correct path
		if r.URL.Path != "/api/v1/users/123" {
			t.Errorf("Expected request to '/api/v1/users/123', got: %s", r.URL.Path)
		}

		// Check that the request has the correct headers
		if r.Header.Get("X-Test-Header") != "test-value" {
			t.Errorf("Expected X-Test-Header: test-value, got: %s", r.Header.Get("X-Test-Header"))
		}

		// Return a test response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    123,
			"name":  "Test User",
			"email": "test@example.com",
		})
	}))
	defer server.Close()

	// Create a configuration
	cfg := config.NewConfig()
	cfg.SetServiceConfig("TestAPI", config.ApiConfig{
		ApiURL:   server.URL,
		ApiToken: "test-token",
	})

	// Create a service
	service := modularapi.NewService(cfg)

	// Create a template
	tmpl := template.NewRouteTemplate("GET", "/api/{{version}}/users/{{user_id}}")
	tmpl.Headers = map[string]string{
		"X-Test-Header": "test-value",
	}

	// Add the template to the service
	service.AddRouteTemplate("TestAPI", "GetUser", *tmpl)

	// Make a request
	var result map[string]interface{}
	err := service.PerformRequest("TestAPI", "GetUser", map[string]interface{}{
		"version": "v1",
		"user_id": "123",
	}, &result)

	// Check that the request succeeded
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check that the response was parsed correctly
	if result["id"] != float64(123) {
		t.Errorf("Expected id: 123, got: %v", result["id"])
	}

	if result["name"] != "Test User" {
		t.Errorf("Expected name: Test User, got: %v", result["name"])
	}

	if result["email"] != "test@example.com" {
		t.Errorf("Expected email: test@example.com, got: %v", result["email"])
	}
}
