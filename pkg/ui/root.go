package ui

import (
	"github.com/c1f/c1f/pkg/api"
	"github.com/c1f/c1f/pkg/ui/common"
	"github.com/c1f/c1f/pkg/ui/instancelist"
	"github.com/c1f/c1f/pkg/ui/stepinspector"
	"github.com/c1f/c1f/pkg/ui/workflowlist"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	workflowListState state = iota
	instanceListState
	stepInspectorState
)

type RootModel struct {
	state         state
	workflowList  workflowlist.Model
	instanceList  instancelist.Model
	stepInspector stepinspector.Model
	client        *api.Client
	error         error
	lastWidth     int
	lastHeight    int
}

func NewRootModel(client *api.Client) RootModel {
	return RootModel{
		state:         workflowListState,
		workflowList:  workflowlist.New(client),
		instanceList:  instancelist.New(client),
		stepInspector: stepinspector.New(client),
		client:        client,
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.workflowList.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.lastWidth = msg.Width
		m.lastHeight = msg.Height
		// Pass to all models to ensure they are ready
		m.workflowList, _ = m.workflowList.Update(msg)
		m.instanceList, _ = m.instanceList.Update(msg)
		m.stepInspector, _ = m.stepInspector.Update(msg)

	case tea.KeyMsg:
		if m.error != nil {
			switch msg.String() {
			case "enter":
				m.error = nil
				if m.state == workflowListState {
					return m, m.workflowList.Init()
				} else if m.state == instanceListState {
					return m, func() tea.Msg { return m.instanceList.FetchInstances() }
				} else {
					return m, m.stepInspector.FetchInstance
				}
			case "esc":
				m.error = nil
				return m, nil
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			return m, nil
		}

		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Let StepInspector handle keys first if it's in showRaw mode
		if m.state == stepInspectorState && m.stepInspector.IsRawMode() {
			var subCmd tea.Cmd
			m.stepInspector, subCmd = m.stepInspector.Update(msg)
			return m, subCmd
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc", "b":
			if m.state == instanceListState {
				m.state = workflowListState
				return m, nil
			} else if m.state == stepInspectorState {
				m.state = instanceListState
				return m, nil
			}
		}

	case common.ErrorMsg:
		if m.state == stepInspectorState {
			m.stepInspector, cmd = m.stepInspector.Update(msg)
			return m, cmd
		}
		m.error = msg.Err
		return m, nil

	case common.WorkflowSelectedMsg:
		m.state = instanceListState
		cmd = m.instanceList.SetWorkflow(msg.Workflow)
		// Ensure the new view knows the size
		m.instanceList, _ = m.instanceList.Update(tea.WindowSizeMsg{Width: m.lastWidth, Height: m.lastHeight})
		return m, cmd

	case common.InstanceSelectedMsg:
		m.state = stepInspectorState
		cmd = m.stepInspector.SetInstance(msg.Workflow, msg.Instance)
		// Ensure the new view knows the size
		m.stepInspector, _ = m.stepInspector.Update(tea.WindowSizeMsg{Width: m.lastWidth, Height: m.lastHeight})
		return m, cmd
	}

	switch m.state {
	case workflowListState:
		m.workflowList, cmd = m.workflowList.Update(msg)
	case instanceListState:
		m.instanceList, cmd = m.instanceList.Update(msg)
	case stepInspectorState:
		m.stepInspector, cmd = m.stepInspector.Update(msg)
	}

	return m, cmd
}

func (m RootModel) View() string {
	if m.error != nil {
		return common.TitleStyle.Render("Error") + "\n\n  " + m.error.Error() + "\n\n  Press Enter to retry, Esc to dismiss, or q to quit"
	}

	switch m.state {
	case workflowListState:
		return m.workflowList.View()
	case instanceListState:
		return m.instanceList.View()
	case stepInspectorState:
		return m.stepInspector.View()
	default:
		return "Unknown state"
	}
}
