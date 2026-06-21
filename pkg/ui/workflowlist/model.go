package workflowlist

import (
	"context"
	"fmt"

	"github.com/c1f/c1f/pkg/api"
	"github.com/c1f/c1f/pkg/models"
	"github.com/c1f/c1f/pkg/ui/common"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type item struct {
	workflow models.Workflow
}

func (i item) Title() string       { return i.workflow.Name }
func (i item) Description() string { return fmt.Sprintf("ID: %s | Created: %s", i.workflow.ID, i.workflow.CreatedAt.Format("2006-01-02 15:04")) }
func (i item) FilterValue() string { return i.workflow.Name }

type Model struct {
	list   list.Model
	client *api.Client
	loaded bool
}

func New(client *api.Client) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Workflows"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return Model{
		list:   l,
		client: client,
	}
}

func (m Model) Init() tea.Cmd {
	return m.FetchWorkflows
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	case workflowsMsg:
		items := make([]list.Item, len(msg.workflows))
		for i, w := range msg.workflows {
			items[i] = item{workflow: w}
		}
		m.list.SetItems(items)
		m.loaded = true
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(item); ok {
				return m, func() tea.Msg {
					return common.WorkflowSelectedMsg{Workflow: i.workflow}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.loaded {
		return "\n  Loading workflows..."
	}
	return m.list.View()
}

type workflowsMsg struct {
	workflows []models.Workflow
}

func (m Model) FetchWorkflows() tea.Msg {
	workflows, err := m.client.ListWorkflows(context.Background())
	if err != nil {
		return common.ErrorMsg{Err: err}
	}
	return workflowsMsg{workflows: workflows}
}
