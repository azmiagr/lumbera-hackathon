package mcsapi

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
	apiURL string
	client *http.Client
}

func Init() Interface {
	return &Client{
		apiURL: os.Getenv("MCS_SCORING_API_URL"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SendMCSRequest(payload any) error {
	if c.apiURL == "" {
		return fmt.Errorf("MCS_SCORING_API_URL is required")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ngrok-skip-browser-warning", "true")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("MCS scoring API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
