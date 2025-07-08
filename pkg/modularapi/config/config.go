package config

// ApiConfig holds the configuration for an API service
type ApiConfig struct {
	ApiURL        string                 `json:"apiURL"`
	ApiToken      string                 `json:"apiToken,omitempty"`
	DefaultParams map[string]interface{} `json:"defaultParams,omitempty"`
}

// Config holds the configuration for the modular API service
type Config struct {
	Services map[string]ApiConfig `json:"services"`
}

// NewConfig creates a new empty configuration
func NewConfig() *Config {
	return &Config{
		Services: make(map[string]ApiConfig),
	}
}

// SetServiceConfig sets the configuration for a specific service
func (c *Config) SetServiceConfig(serviceName string, config ApiConfig) {
	c.Services[serviceName] = config
}

// GetServiceConfig returns the configuration for a specific service
func (c *Config) GetServiceConfig(serviceName string) (ApiConfig, bool) {
	cfg, ok := c.Services[serviceName]
	return cfg, ok
}
