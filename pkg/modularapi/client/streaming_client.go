package client

import (
	"bufio"
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
	w.Header().Set("Access-Control-Allow-Origin", "*") // optional CORS header

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.GlobalLogger.Error("Response writer does not support flushing")
		return "", fmt.Errorf("response writer does not support flushing")
	}

	var responseBuffer bytes.Buffer
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.GlobalLogger.Info("Streaming request completed")
				break // End of stream
			}
			log.GlobalLogger.Errorf("Error reading from streaming response: %v", err)
			return "", fmt.Errorf("error reading from streaming response: %w", err)
		}

		if _, err := w.Write(line); err != nil {
			log.GlobalLogger.Errorf("Error writing to response: %v", err)
			return "", fmt.Errorf("error writing to response: %w", err)
		}

		flusher.Flush()
		responseBuffer.Write(line)
	}

	return responseBuffer.String(), nil
}
