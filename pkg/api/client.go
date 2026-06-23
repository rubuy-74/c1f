package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/c1f/c1f/pkg/models"
)

type Client struct {
	httpClient *http.Client
	apiToken   string
	accountID  string
	Debug      bool
}

func NewClient(apiToken, accountID string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiToken:  apiToken,
		accountID: accountID,
	}
}

func (c *Client) doRequest(ctx context.Context, method, urlPath string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/%s", c.accountID, urlPath)

	var lastErr error
	for i := 0; i <= 3; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i) * time.Second):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiToken)
		req.Header.Set("Content-Type", "application/json")

		if c.Debug {
			fmt.Fprintf(os.Stderr, "> %s %s\n", method, url)
			for k, v := range req.Header {
				fmt.Fprintf(os.Stderr, "> %s: %v\n", k, v)
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		if c.Debug {
			fmt.Fprintf(os.Stderr, "< HTTP %d\n", resp.StatusCode)
			for k, v := range resp.Header {
				fmt.Fprintf(os.Stderr, "< %s: %v\n", k, v)
			}
			fmt.Fprintf(os.Stderr, "< %s\n", string(body))
		}

		if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode <= 599) {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		var apiResp models.APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if !apiResp.Success {
			if len(apiResp.Errors) > 0 {
				return body, apiResp.Errors[0]
			}
			return body, fmt.Errorf("API error")
		}

		return apiResp.Result, nil
	}

	return nil, lastErr
}

func (c *Client) ListWorkflows(ctx context.Context) ([]models.Workflow, error) {
	res, err := c.doRequest(ctx, http.MethodGet, "workflows")
	if err != nil {
		return nil, err
	}

	var workflows []models.Workflow
	if err := json.Unmarshal(res, &workflows); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflows: %w", err)
	}

	return workflows, nil
}

func (c *Client) ListInstances(ctx context.Context, workflowName string) ([]models.Instance, error) {
	res, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("workflows/%s/instances", workflowName))
	if err != nil {
		return nil, err
	}

	var instances []models.Instance
	if err := json.Unmarshal(res, &instances); err != nil {
		return nil, fmt.Errorf("failed to unmarshal instances: %w", err)
	}

	return instances, nil
}

func (c *Client) GetWorkflowInstance(ctx context.Context, workflowName, instanceID string) (models.Instance, error) {
	res, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("workflows/%s/instances/%s", workflowName, instanceID))
	if err != nil {
		return models.Instance{}, err
	}

	var instance models.Instance
	if err := json.Unmarshal(res, &instance); err != nil {
		return models.Instance{}, fmt.Errorf("failed to unmarshal instance: %w", err)
	}

	return instance, nil
}
