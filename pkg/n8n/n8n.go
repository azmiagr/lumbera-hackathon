package n8n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Interface interface {
	SendMCSRequest(payload any) error
}

type Client struct {
	webhookURL string
	client     *http.Client
}

func Init() Interface {
	return &Client{
		webhookURL: os.Getenv("N8N_MCS_WEBHOOK_URL"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SendMCSRequest(payload any) error {
	if c.webhookURL == "" {
		return fmt.Errorf("N8N_MCS_WEBHOOK_URL is required")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("n8n webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
