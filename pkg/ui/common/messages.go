package common

import (
	"github.com/c1f/c1f/pkg/models"
)

type WorkflowSelectedMsg struct {
	Workflow models.Workflow
}

type InstanceSelectedMsg struct {
	Workflow models.Workflow
	Instance models.Instance
}

type ErrorMsg struct {
	Err error
}
