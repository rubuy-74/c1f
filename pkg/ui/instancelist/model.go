package instancelist

import (
	"context"
	"fmt"
	"sort"

	"github.com/c1f/c1f/pkg/api"
	"github.com/c1f/c1f/pkg/models"
	"github.com/c1f/c1f/pkg/ui/common"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	instance models.Instance
}

func (i item) Title() string {
	id := i.instance.ID
	if len(id) > 8 {
		id = id[:8]
	}
	status := i.instance.Status
	statusStyle := lipgloss.NewStyle()
	if status == "success" || status == "running" {
		statusStyle = common.StatusGreen
	} else if status == "failure" {
		statusStyle = common.StatusRed
	}

	return fmt.Sprintf("%s [%s]", id, statusStyle.Render(status))
}

func (i item) Description() string {
	started := common.RelativeTime(i.instance.CreatedAt)
	return fmt.Sprintf("Started: %s | Trigger: %s", started, i.instance.Trigger)
}

func (i item) FilterValue() string { return i.instance.ID }

type Model struct {
	list     list.Model
	client   *api.Client
	workflow models.Workflow
	loaded   bool
}

func New(client *api.Client) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return Model{
		list:   l,
		client: client,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetWorkflow(w models.Workflow) tea.Cmd {
	m.workflow = w
	m.list.Title = fmt.Sprintf("Instances for %s", w.Name)
	m.loaded = false
	return m.FetchInstances
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	case instancesMsg:
		// Custom Sort: Running first, then Most Recent
		sort.Slice(msg.instances, func(i, j int) bool {
			if msg.instances[i].Status == "running" && msg.instances[j].Status != "running" {
				return true
			}
			if msg.instances[i].Status != "running" && msg.instances[j].Status == "running" {
				return false
			}
			return msg.instances[i].CreatedAt.After(msg.instances[j].CreatedAt)
		})

		items := make([]list.Item, len(msg.instances))
		for i, inst := range msg.instances {
			items[i] = item{instance: inst}
		}
		m.list.SetItems(items)
		m.loaded = true
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.loaded {
		return fmt.Sprintf("\n  Loading instances for %s...", m.workflow.Name)
	}
	return m.list.View()
}

type instancesMsg struct {
	instances []models.Instance
}

func (m Model) FetchInstances() tea.Msg {
	instances, err := m.client.ListInstances(context.Background(), m.workflow.Name)
	if err != nil {
		return common.ErrorMsg{Err: err}
	}
	return instancesMsg{instances: instances}
}
