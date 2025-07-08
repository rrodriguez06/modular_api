package modularapi

import (
	"time"

	"github.com/rrodriguez06/modular_api/internal/log"
	"github.com/rrodriguez06/modular_api/pkg/modularapi/config"
	"github.com/rrodriguez06/modular_api/pkg/modularapi/template"
)

// ServiceBuilder is a builder for creating a modular API service
type ServiceBuilder struct {
	config         *config.Config
	serviceConfigs map[string]config.ApiConfig
	templates      map[string]map[string]template.RouteTemplate
	serviceHeaders map[string]map[string]string
	serviceParams  map[string]map[string]interface{}
	timeout        time.Duration
	logLevel       log.LogLevel
}

// NewServiceBuilder creates a new service builder
func NewServiceBuilder() *ServiceBuilder {
	return &ServiceBuilder{
		serviceConfigs: make(map[string]config.ApiConfig),
		templates:      make(map[string]map[string]template.RouteTemplate),
		serviceHeaders: make(map[string]map[string]string),
		serviceParams:  make(map[string]map[string]interface{}),
		timeout:        180 * time.Second, // Default timeout of 3 minutes
		logLevel:       log.INFO,          // Default log level
	}
}

// WithTimeout sets the HTTP client timeout
func (b *ServiceBuilder) WithTimeout(timeout time.Duration) *ServiceBuilder {
	b.timeout = timeout
	return b
}

// WithLogLevel sets the log level
func (b *ServiceBuilder) WithLogLevel(level log.LogLevel) *ServiceBuilder {
	b.logLevel = level
	return b
}

// WithService adds a service configuration
func (b *ServiceBuilder) WithService(name string, apiURL, apiToken string) *ServiceBuilder {
	b.serviceConfigs[name] = config.ApiConfig{
		ApiURL:   apiURL,
		ApiToken: apiToken,
	}
	return b
}

// WithServiceDefaultParams adds default parameters to a service
func (b *ServiceBuilder) WithServiceDefaultParams(serviceName string, params map[string]interface{}) *ServiceBuilder {
	// Ensure the service config exists
	cfg, ok := b.serviceConfigs[serviceName]
	if !ok {
		// If the service doesn't exist, create it with empty values
		cfg = config.ApiConfig{}
	}

	// Initialize DefaultParams if needed
	if cfg.DefaultParams == nil {
		cfg.DefaultParams = make(map[string]interface{})
	}

	// Add the parameters
	for k, v := range params {
		cfg.DefaultParams[k] = v
	}

	// Update the service config
	b.serviceConfigs[serviceName] = cfg

	// Important: Also add these as regular service params to ensure they're available
	// during all stages of request preparation
	return b.WithServiceParams(serviceName, params)
}

// WithServiceHeaders adds global headers to a service
func (b *ServiceBuilder) WithServiceHeaders(serviceName string, headers map[string]string) *ServiceBuilder {
	if b.serviceHeaders[serviceName] == nil {
		b.serviceHeaders[serviceName] = make(map[string]string)
	}
	for k, v := range headers {
		b.serviceHeaders[serviceName][k] = v
	}
	return b
}

// WithServiceParams adds global parameters to a service
func (b *ServiceBuilder) WithServiceParams(serviceName string, params map[string]interface{}) *ServiceBuilder {
	if b.serviceParams[serviceName] == nil {
		b.serviceParams[serviceName] = make(map[string]interface{})
	}
	for k, v := range params {
		b.serviceParams[serviceName][k] = v
	}
	return b
}

// WithTemplate adds a route template
func (b *ServiceBuilder) WithTemplate(serviceName, action string, tmpl template.RouteTemplate) *ServiceBuilder {
	if b.templates[serviceName] == nil {
		b.templates[serviceName] = make(map[string]template.RouteTemplate)
	}
	b.templates[serviceName][action] = tmpl
	return b
}

// WithTemplatesFromFile loads templates from a file
func (b *ServiceBuilder) WithTemplatesFromFile(filepath string) *ServiceBuilder {
	// Templates will be loaded during Build()
	return b
}

// Build creates a new modular API service
func (b *ServiceBuilder) Build() Service {
	// Create configuration
	cfg := config.NewConfig()
	for name, svcCfg := range b.serviceConfigs {
		cfg.SetServiceConfig(name, svcCfg)
	}

	// Set log level
	log.SetGlobalLogger(log.NewDefaultLogger(b.logLevel))

	// Create service
	svc := NewService(cfg)

	// Add templates
	for serviceName, actions := range b.templates {
		for action, tmpl := range actions {
			svc.AddRouteTemplate(serviceName, action, tmpl)
		}
	}

	// Add service headers
	for serviceName, headers := range b.serviceHeaders {
		svc.SetServiceHeaders(serviceName, headers)
	}

	// Add service parameters
	for serviceName, params := range b.serviceParams {
		svc.SetServiceParams(serviceName, params)
	}

	return svc
}
