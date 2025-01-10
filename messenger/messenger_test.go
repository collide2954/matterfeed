// messenger/messenger_test.go
package messenger_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"matterfeed/messenger"
)

func TestSendMessageSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	url := server.URL + "/hooks/test"
	message := "Test message"

	err := messenger.SendMessage(url, message)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestSendMessageFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Not Found")
	}))
	defer server.Close()

	url := server.URL + "/hooks/test"
	message := "Test message"

	err := messenger.SendMessage(url, message)
	if err == nil {
		t.Errorf("Expected an error, got none")
	}
}

func TestSendMessageNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		conn.Close()
	}))
	defer server.Close()

	url := server.URL + "/hooks/test"
	message := "Test message"

	err := messenger.SendMessage(url, message)
	if err == nil {
		t.Errorf("Expected an error, got none")
	}
}

func TestSendMessageTimeout(t *testing.T) {
	originalDefaultTransport := http.DefaultTransport
	http.DefaultTransport = &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return nil, &net.DNSError{IsTimeout: true}
		},
	}
	defer func() { http.DefaultTransport = originalDefaultTransport }()

	url := "https://example.com/hooks/test"
	message := "Test message"

	err := messenger.SendMessage(url, message)
	if err == nil {
		t.Errorf("Expected an error, got none")
	}
}

func TestSendMessageMarshalError(t *testing.T) {
	url := "https://example.com/hooks/test"
	message := "\x00\x01\x02" // Invalid UTF-8 sequence

	err := messenger.SendMessage(url, message)
	if err == nil {
		t.Errorf("Expected an error, got none")
	}
}

func TestSendMessageResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad Request: Invalid payload")
	}))
	defer server.Close()

	url := server.URL + "/hooks/test"
	message := "Test message"

	err := messenger.SendMessage(url, message)
	if err == nil {
		t.Errorf("Expected an error, got none")
	}
	if !strings.Contains(fmt.Sprintf("%v", err), "response: Bad Request: Invalid payload") {
		t.Errorf("Expected error to contain response body, but got: %v", err)
	}
}
