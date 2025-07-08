package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rrodriguez06/modular_api/internal/log"
)

// HTTPClient is an interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the HTTP client used by the API service
type Client struct {
	httpClient HTTPClient
	timeout    time.Duration
}

// NewClient creates a new HTTP client with the specified timeout
func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// SetTimeout sets the client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient = &http.Client{
		Timeout: timeout,
	}
}

// MakeRequest performs an HTTP request and unmarshals the response into the result
func (c *Client) MakeRequest(req *http.Request, result interface{}) error {
	// Log request details for debugging purposes
	if req.Body != nil {
		// Read the request body
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			log.GlobalLogger.Errorf("Error reading request body: %v", err)
			return fmt.Errorf("error reading request body: %w", err)
		}

		// Restore the body for the actual request
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Log the request
		log.GlobalLogger.Infof("API Request to %s: %s\nHeaders: %v\nBody: %s",
			req.URL.String(), req.Method, req.Header, string(bodyBytes))
	} else {
		log.GlobalLogger.Infof("API Request to %s: %s\nHeaders: %v\nNo Body",
			req.URL.String(), req.Method, req.Header)
	}

	// Make the actual request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot perform request: %w", err)
	}
	defer resp.Body.Close()

	log.GlobalLogger.Infof("API Response Status: %d %s", resp.StatusCode, resp.Status)
	log.GlobalLogger.Infof("API Response Headers: %v", resp.Header)

	// Read the response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body: %w", err)
	}
	// Put the body back
	resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes))

	// Log response body for all responses to help with debugging
	log.GlobalLogger.Infof("API Response Body (raw): %s", string(respBodyBytes))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.GlobalLogger.Errorf("API call error: %s", string(respBodyBytes))
		return fmt.Errorf("API call error: %s, status code: %d", string(respBodyBytes), resp.StatusCode)
	}

	if result != nil && len(respBodyBytes) > 0 {
		// Put the body back again for decoding
		resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes))

		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			log.GlobalLogger.Errorf("Cannot decode response: %v", err)
			return fmt.Errorf("cannot decode response: %w", err)
		}
	}

	return nil
}
