package common

import (
	"github.com/c1f/c1f/pkg/models"
)

type WorkflowSelectedMsg struct {
	Workflow models.Workflow
}

type ErrorMsg struct {
	Err error
}
