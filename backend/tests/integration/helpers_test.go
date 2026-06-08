//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const defaultAPI = "http://localhost:8080"

func apiBase() string {
	if u := os.Getenv("API_URL"); u != "" {
		return u
	}
	return defaultAPI
}

type apiClient struct {
	base    string
	token   string
	agency  string
	client  *http.Client
}

func newClient() *apiClient {
	return &apiClient{
		base:   apiBase() + "/api/v1",
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *apiClient) do(ctx context.Context, method, path string, body interface{}, out interface{}) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.agency != "" {
		req.Header.Set("X-Agency-ID", c.agency)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if out != nil && resp.StatusCode < 300 {
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(out)
	}
	return resp, nil
}

func (c *apiClient) login(ctx context.Context, email, password string) error {
	var result struct {
		AccessToken string `json:"access_token"`
		AgencyID    string `json:"agency_id"`
		User        struct {
			AgencyID string `json:"agency_id"`
		} `json:"user"`
	}
	resp, err := c.do(ctx, http.MethodPost, "/auth/login", map[string]string{
		"email": email, "password": password,
	}, &result)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: %d", resp.StatusCode)
	}
	c.token = result.AccessToken
	c.agency = result.AgencyID
	if c.agency == "" {
		c.agency = result.User.AgencyID
	}
	return nil
}

func skipIfNoAPI(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiBase()+"/health", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skip("API not available at " + apiBase())
	}
	resp.Body.Close()
}
