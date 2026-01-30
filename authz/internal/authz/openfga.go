package authz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Client provides authorization checks using OpenFGA
type Client struct {
	baseURL string
	storeID string
	modelID string
	client  *http.Client
	mu      sync.RWMutex
	devMode bool
}

// NewClient creates a new OpenFGA authorization client
func NewClient(baseURL, storeID string, devMode bool) *Client {
	return &Client{
		baseURL: baseURL,
		storeID: storeID,
		devMode: devMode,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Initialize fetches the latest authorization model ID
func (c *Client) Initialize(ctx context.Context) error {
	if c.devMode {
		return nil
	}

	if c.storeID == "" {
		return fmt.Errorf("store ID not configured")
	}

	url := fmt.Sprintf("%s/stores/%s/authorization-models", c.baseURL, c.storeID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to OpenFGA: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get models: %s - %s", resp.Status, string(body))
	}

	var result struct {
		AuthorizationModels []struct {
			ID string `json:"id"`
		} `json:"authorization_models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if len(result.AuthorizationModels) == 0 {
		return fmt.Errorf("no authorization models found")
	}

	c.mu.Lock()
	c.modelID = result.AuthorizationModels[0].ID
	c.mu.Unlock()

	return nil
}

// Check performs an authorization check
func (c *Client) Check(ctx context.Context, userID, workspaceID, permission, path string) (bool, error) {
	if c.devMode {
		return true, nil
	}

	c.mu.RLock()
	storeID := c.storeID
	modelID := c.modelID
	c.mu.RUnlock()

	if storeID == "" || modelID == "" {
		return false, fmt.Errorf("authz client not initialized")
	}

	user := fmt.Sprintf("user:%s", userID)
	object := fmt.Sprintf("workspace:%s", workspaceID)

	reqBody := map[string]interface{}{
		"tuple_key": map[string]string{
			"user":     user,
			"relation": permission,
			"object":   object,
		},
		"authorization_model_id": modelID,
	}

	body, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/stores/%s/check", c.baseURL, storeID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("check failed: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Allowed, nil
}

// WriteTuple writes an authorization tuple
func (c *Client) WriteTuple(ctx context.Context, user, relation, object string) error {
	if c.devMode {
		return nil
	}

	c.mu.RLock()
	storeID := c.storeID
	modelID := c.modelID
	c.mu.RUnlock()

	reqBody := map[string]interface{}{
		"writes": map[string]interface{}{
			"tuple_keys": []map[string]string{
				{
					"user":     user,
					"relation": relation,
					"object":   object,
				},
			},
		},
		"authorization_model_id": modelID,
	}

	body, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/stores/%s/write", c.baseURL, storeID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("write failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

// DeleteTuple deletes an authorization tuple
func (c *Client) DeleteTuple(ctx context.Context, user, relation, object string) error {
	if c.devMode {
		return nil
	}

	c.mu.RLock()
	storeID := c.storeID
	modelID := c.modelID
	c.mu.RUnlock()

	reqBody := map[string]interface{}{
		"deletes": map[string]interface{}{
			"tuple_keys": []map[string]string{
				{
					"user":     user,
					"relation": relation,
					"object":   object,
				},
			},
		},
		"authorization_model_id": modelID,
	}

	body, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/stores/%s/write", c.baseURL, storeID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s - %s", resp.Status, string(body))
	}

	return nil
}
