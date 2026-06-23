package models

import (
	"encoding/json"
	"fmt"
	"strings"
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

// Normalize maps common API status strings to canonical StepStatus values.
func (s StepStatus) Normalize() StepStatus {
	switch strings.ToLower(string(s)) {
	case "success", "successful", "complete", "completed", "done":
		return StepStatusSuccess
	case "running", "in_progress", "in-progress", "executing":
		return StepStatusRunning
	case "failure", "failed", "error", "errored":
		return StepStatusFailure
	case "pending", "queued", "scheduled":
		return StepStatusPending
	case "skipped", "skip":
		return StepStatusSkipped
	case "canceled", "cancelled", "terminated":
		return StepStatusCanceled
	default:
		return s
	}
}

type RetryConfig struct {
	Limit int `json:"limit"`
}

func (r *RetryConfig) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	// API may return retries as an object or an int.
	if data[0] == '{' {
		type rawRetryConfig RetryConfig
		return json.Unmarshal(data, (*rawRetryConfig)(r))
	}
	var count int
	if err := json.Unmarshal(data, &count); err != nil {
		return err
	}
	r.Limit = count
	return nil
}

func (r RetryConfig) String() string {
	return fmt.Sprintf("%d", r.Limit)
}

type TimeoutConfig struct {
	Seconds int    `json:"seconds"`
	Raw     string `json:"-"`
}

func (t *TimeoutConfig) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	// API may return timeout as an object, int, or string.
	if data[0] == '{' {
		type rawTimeoutConfig TimeoutConfig
		return json.Unmarshal(data, (*rawTimeoutConfig)(t))
	}
	if data[0] == '"' {
		return json.Unmarshal(data, &t.Raw)
	}
	return json.Unmarshal(data, &t.Seconds)
}

func (t TimeoutConfig) String() string {
	if t.Raw != "" {
		return t.Raw
	}
	if t.Seconds > 0 {
		return fmt.Sprintf("%ds", t.Seconds)
	}
	return ""
}

type StepConfig struct {
	Retries RetryConfig   `json:"retries"`
	Timeout TimeoutConfig `json:"timeout"`
}

type StepError struct {
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace"`
}

type Attempt struct {
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Status     string     `json:"status"`
}

type Attempts []Attempt

func (a *Attempts) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Cloudflare API returns attempts as an array of objects.
	// Fallback to int (count) if the API ever changes.
	if data[0] == '[' {
		return json.Unmarshal(data, (*[]Attempt)(a))
	}

	var count int
	if err := json.Unmarshal(data, &count); err != nil {
		return err
	}
	*a = nil
	for i := 0; i < count; i++ {
		*a = append(*a, Attempt{})
	}
	return nil
}

func (a Attempts) Count() int {
	return len(a)
}

type Step struct {
	Name       string           `json:"name"`
	Type       string           `json:"type"`
	Status     StepStatus       `json:"status"`
	StartedAt  time.Time        `json:"started_at"`
	StartedOn  time.Time        `json:"started_on"`
	FinishedAt *time.Time       `json:"finished_at,omitempty"`
	FinishedOn *time.Time       `json:"finished_on,omitempty"`
	Config     *StepConfig      `json:"config,omitempty"`
	Error      *StepError       `json:"error,omitempty"`
	Output     *json.RawMessage `json:"output,omitempty"`
	Attempts   Attempts         `json:"attempts"`
}

func (s Step) DisplayStartedAt() time.Time {
	if !s.StartedOn.IsZero() {
		return s.StartedOn
	}
	return s.StartedAt
}

func (s Step) DisplayFinishedAt() *time.Time {
	if s.FinishedOn != nil && !s.FinishedOn.IsZero() {
		return s.FinishedOn
	}
	return s.FinishedAt
}

type Workflow struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedOn  time.Time `json:"created_on"`
	ModifiedAt time.Time `json:"modified_at"`
	ModifiedOn time.Time `json:"modified_on"`
}

func (w Workflow) DisplayCreatedAt() time.Time {
	if !w.CreatedOn.IsZero() {
		return w.CreatedOn
	}
	return w.CreatedAt
}

func (w Workflow) DisplayModifiedAt() time.Time {
	if !w.ModifiedOn.IsZero() {
		return w.ModifiedOn
	}
	return w.ModifiedAt
}

type Trigger struct {
	Type   string `json:"type"`
	Source string `json:"source,omitempty"`
}

func (t *Trigger) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Cloudflare API returns trigger as an object.
	// Fallback to string if it ever changes.
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		t.Type = s
		return nil
	}

	type rawTrigger Trigger
	return json.Unmarshal(data, (*rawTrigger)(t))
}

func (t Trigger) String() string {
	if t.Type == "" && t.Source == "" {
		return ""
	}
	if t.Type == "" {
		return t.Source
	}
	if t.Source == "" {
		return t.Type
	}
	return fmt.Sprintf("%s / %s", t.Type, t.Source)
}

type Instance struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	CreatedOn time.Time `json:"created_on"`
	Steps     []Step    `json:"steps"`
	Trigger   Trigger   `json:"trigger"`
	VersionID string    `json:"version_id"`
}

func (i Instance) DisplayCreatedAt() time.Time {
	if !i.CreatedOn.IsZero() {
		return i.CreatedOn
	}
	return i.CreatedAt
}

func (i *Instance) CalculateProgress() (float64, string) {
	if len(i.Steps) == 0 {
		return 0, "0% (0/0 steps)"
	}

	successCount := 0
	for _, step := range i.Steps {
		if step.Status.Normalize() == StepStatusSuccess {
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
