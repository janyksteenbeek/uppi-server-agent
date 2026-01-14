package reporter

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/janyksteenbeek/uppi-server-agent/internal/collector"
	"github.com/janyksteenbeek/uppi-server-agent/internal/config"
)

// SendMetrics collects and sends metrics to the Uppi server
func SendMetrics(cfg config.Config) error {
	metrics, err := collector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Create HMAC signature
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := createHMACSignature(timestamp, string(payload), cfg.Secret)

	// Create request
	url := fmt.Sprintf("%s/api/server/%s/report", cfg.Instance, cfg.ServerId)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Timestamp", timestamp)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func createHMACSignature(timestamp, payload, secret string) string {
	message := timestamp + payload
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return fmt.Sprintf("%x", h.Sum(nil))
}
