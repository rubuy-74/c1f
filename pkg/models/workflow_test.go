package models

import (
	"testing"
)

func TestInstance_CalculateProgress(t *testing.T) {
	tests := []struct {
		name       string
		steps      []Step
		wantRatio  float64
		wantString string
	}{
		{
			name:       "empty steps",
			steps:      []Step{},
			wantRatio:  0,
			wantString: "0% (0/0 steps)",
		},
		{
			name: "all success",
			steps: []Step{
				{Status: StepStatusSuccess},
				{Status: StepStatusSuccess},
			},
			wantRatio:  100.0,
			wantString: "100% (2/2 steps)",
		},
		{
			name: "partial success",
			steps: []Step{
				{Status: StepStatusSuccess},
				{Status: StepStatusRunning},
				{Status: StepStatusPending},
				{Status: StepStatusSuccess},
				{Status: StepStatusFailure},
			},
			wantRatio:  40.0,
			wantString: "40% (2/5 steps)",
		},
		{
			name: "zero success",
			steps: []Step{
				{Status: StepStatusRunning},
				{Status: StepStatusPending},
			},
			wantRatio:  0,
			wantString: "0% (0/2 steps)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Instance{
				Steps: tt.steps,
			}
			gotRatio, gotString := i.CalculateProgress()
			if gotRatio != tt.wantRatio {
				t.Errorf("CalculateProgress() gotRatio = %v, want %v", gotRatio, tt.wantRatio)
			}
			if gotString != tt.wantString {
				t.Errorf("CalculateProgress() gotString = %v, want %v", gotString, tt.wantString)
			}
		})
	}
}
