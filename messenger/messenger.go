// messenger/messenger.go
package messenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"matterfeed/logger"
)

type Message struct {
	Text string `json:"text"`
}

func SendMessage(url, message string) error {
	msg := Message{Text: message}
	payload, err := json.Marshal(msg)
	if err != nil {
		return logger.LogAndReturnError(err, "failed to marshal message")
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok {
			if opErr.Timeout() {
				return logger.LogAndReturnError(err, "network timeout error")
			}
			return logger.LogAndReturnError(err, "network error")
		}
		return logger.LogAndReturnError(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return logger.LogAndReturnError(fmt.Errorf("failed to send message, status: %d, response: %s", resp.StatusCode, body), "sending message")
	}

	return nil
}
