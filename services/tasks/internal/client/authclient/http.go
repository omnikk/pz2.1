package authclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type verifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject"`
	Error   string `json:"error"`
}

// Verify проверяет токен через Auth service.
// Возвращает subject и nil при успехе, или ошибку.
func (c *Client) Verify(ctx context.Context, token, requestID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/auth/verify", nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth returned %d", resp.StatusCode)
	}

	var body verifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return body.Subject, nil
}

// ErrUnauthorized — токен невалиден.
var ErrUnauthorized = fmt.Errorf("unauthorized")
