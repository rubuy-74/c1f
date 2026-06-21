package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type StepStatus string

const (
	StepStatusSuccess  StepStatus = "success"
	StepStatusRunning  StepStatus = "running"
	StepStatusFailure  StepStatus = "failure"
	StepStatusPending  StepStatus = "pending"
	StepStatusSkipped  StepStatus = "skipped"
	StepStatusCanceled StepStatus = "canceled"
)

type Step struct {
	Name   string     `json:"name"`
	Status StepStatus `json:"status"`
}

type Workflow struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type Instance struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Steps     []Step    `json:"steps"`
	Trigger   string    `json:"trigger"`
}

func (i *Instance) CalculateProgress() (float64, string) {
	if len(i.Steps) == 0 {
		return 0, "0% (0/0 steps)"
	}

	successCount := 0
	for _, step := range i.Steps {
		if step.Status == StepStatusSuccess {
			successCount++
		}
	}

	ratio := float64(successCount) / float64(len(i.Steps))
	percentage := ratio * 100
	return percentage, fmt.Sprintf("%.0f%% (%d/%d steps)", percentage, successCount, len(i.Steps))
}

type APIResponse struct {
	Success  bool            `json:"success"`
	Errors   []APIError      `json:"errors"`
	Messages []string        `json:"messages"`
	Result   json.RawMessage `json:"result"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s (Code %d)", e.Message, e.Code)
}
