package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/rrodriguez06/modular_api/internal/log"
)

// StreamingClient handles streaming HTTP requests
type StreamingClient struct {
	httpClient HTTPClient
}

// NewStreamingClient creates a new streaming client
func NewStreamingClient() *StreamingClient {
	return &StreamingClient{
		httpClient: &http.Client{
			Timeout: 0, // No timeout for streaming
		},
	}
}

// MakeStreamingRequest performs a streaming HTTP request
func (c *StreamingClient) MakeStreamingRequest(req *http.Request, w http.ResponseWriter) (string, error) {
	log.GlobalLogger.Infof("API Streaming Request to %s: %s\nHeaders: %v", req.URL.String(), req.Method, req.Header)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.GlobalLogger.Errorf("Error performing streaming request: %v", err)
		return "", fmt.Errorf("error performing streaming request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.GlobalLogger.Errorf("Streaming API call error: %s", string(bodyBytes))
		return "", fmt.Errorf("streaming API call error: %s, status code: %d", string(bodyBytes), resp.StatusCode)
	}

	// Set headers on our response to the client to indicate streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.GlobalLogger.Error("Response writer does not support flushing")
		return "", fmt.Errorf("response writer does not support flushing")
	}

	var responseBuffer bytes.Buffer
	buffer := make([]byte, 4096) // Use a fixed-size buffer to read chunks of data

	for {
		// Read a chunk of data
		n, err := resp.Body.Read(buffer)

		// Process any data received, even in case of an error
		if n > 0 {
			chunk := buffer[:n]

			// Write chunk to the client
			if _, writeErr := w.Write(chunk); writeErr != nil {
				log.GlobalLogger.Errorf("Error writing to response: %v", writeErr)
				return responseBuffer.String(), fmt.Errorf("error writing to response: %w", writeErr)
			}

			// Flush to ensure data is sent to the client immediately
			flusher.Flush()

			// Store in our response buffer
			responseBuffer.Write(chunk)
		}

		// Handle any errors after processing data
		if err != nil {
			if err == io.EOF {
				log.GlobalLogger.Info("Streaming request completed")
				break // End of stream
			}
			log.GlobalLogger.Errorf("Error reading from streaming response: %v", err)
			return responseBuffer.String(), fmt.Errorf("error reading from streaming response: %w", err)
		}
	}

	return responseBuffer.String(), nil
}
