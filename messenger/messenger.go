// messenger/messenger.go
package messenger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
				return fmt.Errorf("network timeout error: %w", err)
			}
			return fmt.Errorf("network error: %w", err)
		}
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send message, status: %d, response: %s", resp.StatusCode, body)
	}

	return nil
}
