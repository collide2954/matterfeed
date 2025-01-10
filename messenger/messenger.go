// messenger/messenger.go
package messenger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

type Message struct {
	Text string `json:"text"`
}

func SendMessage(url, message string) error {
	msg := Message{Text: message}
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			if opErr.Timeout() {
				log.Printf("Network timeout error: %v", err)
			}
			log.Printf("Network error: %v", err)
		}
		log.Printf("Failed to send HTTP request: %v", err)
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to send message, status: %d, response: %s", resp.StatusCode, body)
	}

	return nil
}
