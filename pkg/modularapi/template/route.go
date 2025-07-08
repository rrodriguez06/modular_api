package template

// RouteTemplate defines a template for an API route
type RouteTemplate struct {
	Method         string                 `json:"method"`
	Endpoint       string                 `json:"endpoint"`
	Headers        map[string]string      `json:"headers"`
	PathParams     []string               `json:"pathParams,omitempty"`
	QueryParams    map[string]interface{} `json:"queryParams,omitempty"`
	Body           map[string]interface{} `json:"body,omitempty"`
	OptionalParams map[string]bool        `json:"-"` // Tracks which parameters are optional
}

// NewRouteTemplate creates a new route template with initialized maps
func NewRouteTemplate(method, endpoint string) *RouteTemplate {
	return &RouteTemplate{
		Method:         method,
		Endpoint:       endpoint,
		Headers:        make(map[string]string),
		PathParams:     []string{},
		QueryParams:    make(map[string]interface{}),
		Body:           make(map[string]interface{}),
		OptionalParams: make(map[string]bool),
	}
}

// WithHeaders adds headers to the route template
func (rt *RouteTemplate) WithHeaders(headers map[string]string) *RouteTemplate {
	for k, v := range headers {
		rt.Headers[k] = v
	}
	return rt
}

// WithQueryParams adds query parameters to the route template
func (rt *RouteTemplate) WithQueryParams(params map[string]interface{}) *RouteTemplate {
	for k, v := range params {
		rt.QueryParams[k] = v
	}
	return rt
}

// WithBody adds body parameters to the route template
func (rt *RouteTemplate) WithBody(body map[string]interface{}) *RouteTemplate {
	for k, v := range body {
		rt.Body[k] = v
	}
	return rt
}

// Clone creates a deep copy of the route template
func (rt *RouteTemplate) Clone() *RouteTemplate {
	clone := NewRouteTemplate(rt.Method, rt.Endpoint)

	// Copy headers
	for k, v := range rt.Headers {
		clone.Headers[k] = v
	}

	// Copy path parameters
	clone.PathParams = make([]string, len(rt.PathParams))
	copy(clone.PathParams, rt.PathParams)

	// Copy query parameters
	for k, v := range rt.QueryParams {
		clone.QueryParams[k] = v
	}

	// Copy body
	for k, v := range rt.Body {
		clone.Body[k] = v
	}

	// Copy optional parameters
	for k, v := range rt.OptionalParams {
		clone.OptionalParams[k] = v
	}

	return clone
}
