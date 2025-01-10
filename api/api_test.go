// api/api_test.go
package api_test

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"matterfeed/api"
	"matterfeed/config"
)

func TestStartAPIServer(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			Port: 8080,
		},
	}

	stopCh := make(chan struct{})

	go api.StartAPIServer(cfg, stopCh)

	time.Sleep(100 * time.Millisecond)

	if !isPortInUse(cfg.API.Port) {
		t.Errorf("Expected port %d to be in use before shutdown", cfg.API.Port)
	}

	client := &http.Client{}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", cfg.API.Port))
	if err != nil {
		t.Fatalf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	if resp.ContentLength > 0 {
		t.Errorf("Expected empty response body, got length %d", resp.ContentLength)
	}

	close(stopCh)

	time.Sleep(500 * time.Millisecond)

	if isPortInUse(cfg.API.Port) {
		t.Errorf("Expected port %d to be free after shutdown", cfg.API.Port)
	}
}

func isPortInUse(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return true
	}
	if err := listener.Close(); err != nil {
		panic(err)
	}
	return false
}
