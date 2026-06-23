package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GraphQLClient struct {
	httpClient *http.Client
	apiToken   string
	accountID  string
}

func NewGraphQLClient(apiToken, accountID string) *GraphQLClient {
	return &GraphQLClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiToken:  apiToken,
		accountID: accountID,
	}
}

type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type graphqlResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

func (c *GraphQLClient) doQuery(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
	reqBody := graphqlRequest{
		Query:     query,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := "https://api.cloudflare.com/client/v4/graphql"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result graphqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	return result.Data, nil
}

type workflowsAdaptiveGroupsResponse struct {
	Viewer struct {
		Accounts []accountData `json:"accounts"`
	} `json:"viewer"`
}

type accountData struct {
	InstanceRuns []workflowGroup `json:"instanceRuns"`
	WallTime     []workflowGroup `json:"wallTime"`
	Success      []workflowGroup `json:"success"`
	WorkflowFail []workflowGroup `json:"workflowFail"`
}

type workflowGroup struct {
	Count      int64            `json:"count"`
	Sum        sumData          `json:"sum"`
	Dimensions dimensionsData    `json:"dimensions"`
}

type sumData struct {
	WallTime float64 `json:"wallTime"`
}

type dimensionsData struct {
	Date string `json:"date"`
}

func (c *GraphQLClient) FetchInvocationCount(ctx context.Context, workflowName, accountId, datetime_geq, datetime_leq string) (int64, []float64, error) {
	query := `
	query GetInvocationCount($accountId: string!, $workflowName: string!, $datetimeStart: Time!, $datetimeEnd: Time!) {
	  viewer {
	    accounts(filter: { accountTag: $accountId }) {
	      instanceRuns: workflowsAdaptiveGroups(
	        limit: 10000
	        filter: {
	          datetimeHour_geq: $datetimeStart
	          datetimeHour_leq: $datetimeEnd
	          workflowName: $workflowName
	          eventType: "WORKFLOW_START"
	        }
	      ) {
	        count
	        dimensions {
	          date: datetimeHour
	        }
	      }
	    }
	  }
	}
	`

	variables := map[string]interface{}{
		"accountId":     accountId,
		"workflowName":  workflowName,
		"datetimeStart": datetime_geq,
		"datetimeEnd":   datetime_leq,
	}

	data, err := c.doQuery(ctx, query, variables)
	if err != nil {
		return 0, nil, err
	}

	var result workflowsAdaptiveGroupsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return 0, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var total int64
	var buckets []float64
	for _, account := range result.Viewer.Accounts {
		for _, group := range account.InstanceRuns {
			total += group.Count
			buckets = append(buckets, float64(group.Count))
		}
	}

	return total, buckets, nil
}

func (c *GraphQLClient) FetchWallTime(ctx context.Context, workflowName, accountId, datetime_geq, datetime_leq string) (float64, int64, []float64, error) {
	query := `
	query GetWallTime($accountId: string!, $workflowName: string!, $datetimeStart: Time!, $datetimeEnd: Time!) {
	  viewer {
	    accounts(filter: { accountTag: $accountId }) {
	      wallTime: workflowsAdaptiveGroups(
	        limit: 10000
	        filter: {
	          datetimeHour_geq: $datetimeStart
	          datetimeHour_leq: $datetimeEnd
	          workflowName: $workflowName
	        }
	      ) {
	        count
	        sum {
	          wallTime
	        }
	        dimensions {
	          date: datetimeHour
	        }
	      }
	    }
	  }
	}
	`

	variables := map[string]interface{}{
		"accountId":     accountId,
		"workflowName":  workflowName,
		"datetimeStart": datetime_geq,
		"datetimeEnd":   datetime_leq,
	}

	data, err := c.doQuery(ctx, query, variables)
	if err != nil {
		return 0, 0, nil, err
	}

	var result workflowsAdaptiveGroupsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return 0, 0, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var totalSum int64
	var totalCount int
	var buckets []float64

	for _, account := range result.Viewer.Accounts {
		for _, group := range account.WallTime {
			totalSum += int64(group.Sum.WallTime)
			totalCount += int(group.Count)
			buckets = append(buckets, group.Sum.WallTime)
		}
	}

	var avgMs float64
	if totalCount > 0 {
		avgMs = float64(totalSum) / float64(totalCount)
	}

	return avgMs, totalSum, buckets, nil
}

func (c *GraphQLClient) FetchFailureRate(ctx context.Context, workflowName, accountId, datetime_geq, datetime_leq string) (float64, []float64, error) {
	query := `
	query GetFailureRate($accountId: string!, $workflowName: string!, $datetimeStart: Time!, $datetimeEnd: Time!) {
	  viewer {
	    accounts(filter: { accountTag: $accountId }) {
	      success: workflowsAdaptiveGroups(
	        limit: 10000
	        filter: {
	          datetimeHour_geq: $datetimeStart
	          datetimeHour_leq: $datetimeEnd
	          workflowName: $workflowName
	          eventType: "WORKFLOW_SUCCESS"
	        }
	      ) {
	        count
	      }
	      workflowFail: workflowsAdaptiveGroups(
	        limit: 10000
	        filter: {
	          datetimeHour_geq: $datetimeStart
	          datetimeHour_leq: $datetimeEnd
	          workflowName: $workflowName
	          eventType: "WORKFLOW_FAILURE"
	        }
	      ) {
	        count
	        dimensions {
	          date: datetimeHour
	        }
	      }
	    }
	  }
	}
	`

	variables := map[string]interface{}{
		"accountId":     accountId,
		"workflowName":  workflowName,
		"datetimeStart": datetime_geq,
		"datetimeEnd":   datetime_leq,
	}

	data, err := c.doQuery(ctx, query, variables)
	if err != nil {
		return 0, nil, err
	}

	var result workflowsAdaptiveGroupsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return 0, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var successCount, failCount int64
	var buckets []float64

	for _, account := range result.Viewer.Accounts {
		for _, group := range account.Success {
			successCount += group.Count
		}
		for _, group := range account.WorkflowFail {
			failCount += group.Count
			buckets = append(buckets, float64(group.Count))
		}
	}

	total := successCount + failCount
	var failRatio float64
	if total > 0 {
		failRatio = float64(failCount) / float64(total)
	}

	return failRatio, buckets, nil
}
