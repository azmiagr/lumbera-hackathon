package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Interface interface {
	SendMessage(target, message string) error
}

type FonnteClient struct {
	token  string
	apiURL string
	client *http.Client
}

type FonnteResponse struct {
	Status bool   `json:"status"`
	Reason string `json:"reason"`
	Detail any    `json:"detail"`
}

func Init() Interface {
	return &FonnteClient{
		token:  os.Getenv("FONNTE_TOKEN"),
		apiURL: "https://api.fonnte.com/send",
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *FonnteClient) SendMessage(target, message string) error {
	if c.token == "" {
		return fmt.Errorf("FONNTE_TOKEN is required")
	}

	form := url.Values{}
	form.Set("target", target)
	form.Set("message", message)

	req, err := http.NewRequest(http.MethodPost, c.apiURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fonnte API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result FonnteResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return fmt.Errorf("failed to parse Fonnte response: %w", err)
	}

	if !result.Status {
		return fmt.Errorf("fonnte failed: %s", result.Reason)
	}

	return nil
}
