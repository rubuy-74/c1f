package instancelist

import (
	"context"
	"fmt"
	"sort"

	"github.com/c1f/c1f/pkg/api"
	"github.com/c1f/c1f/pkg/models"
	"github.com/c1f/c1f/pkg/ui/common"
	"github.com/c1f/c1f/pkg/ui/help"
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
	createdAt := i.instance.DisplayCreatedAt()
	started := common.RelativeTime(createdAt)
	if createdAt.IsZero() {
		started = "—"
	}
	trigger := i.instance.Trigger.String()
	if trigger == "" {
		trigger = "—"
	}
	return fmt.Sprintf("Started: %s | Trigger: %s", started, trigger)
}

func (i item) FilterValue() string { return i.instance.ID }

type Model struct {
	list     list.Model
	help     help.Model
	client   *api.Client
	workflow models.Workflow
	loaded   bool
}

func New(client *api.Client) Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#A3A3A3"))
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF87D7")).Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB2D7"))

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()

	h := help.New([]help.Section{
		{
			Title: "Global",
			Bindings: []help.Binding{
				{Key: "?", Description: "Toggle help"},
				{Key: "q", Description: "Quit"},
				{Key: "esc / b", Description: "Go back / dismiss"},
			},
		},
		{
			Title: "Navigation",
			Bindings: []help.Binding{
				{Key: "j / k", Description: "Move down / up"},
				{Key: "gg / G", Description: "Jump to top / bottom"},
				{Key: "/", Description: "Filter instances"},
			},
		},
		{
			Title: "Actions",
			Bindings: []help.Binding{
				{Key: "enter", Description: "View instance steps"},
			},
		},
	})

	return Model{
		list: l,
		help: h,
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
		m.help.SetSize(msg.Width, msg.Height)
	case instancesMsg:
		// Custom Sort: Running first, then Most Recent
		sort.Slice(msg.instances, func(i, j int) bool {
			if msg.instances[i].Status == "running" && msg.instances[j].Status != "running" {
				return true
			}
			if msg.instances[i].Status != "running" && msg.instances[j].Status == "running" {
				return false
			}
			return msg.instances[i].DisplayCreatedAt().After(msg.instances[j].DisplayCreatedAt())
		})

		items := make([]list.Item, len(msg.instances))
		for i, inst := range msg.instances {
			items[i] = item{instance: inst}
		}
		m.list.SetItems(items)
		m.loaded = true
	}

	// Help overlay takes precedence.
	if m.help.Visible() {
		var helpCmd tea.Cmd
		m.help, helpCmd = m.help.Update(msg)
		return m, helpCmd
	}

	// Normal key handling.
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.list.FilterState() != list.Filtering {
			switch keyMsg.String() {
			case "?":
				m.help.Show()
				return m, nil
			case "gg":
				m.list.GoToStart()
				return m, nil
			case "G":
				m.list.GoToEnd()
				return m, nil
			case "/":
				m.list.SetFilterState(list.Filtering)
				return m, nil
			case "enter":
				if i, ok := m.list.SelectedItem().(item); ok {
					return m, func() tea.Msg {
						return common.InstanceSelectedMsg{
							Workflow: m.workflow,
							Instance: i.instance,
						}
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) IsHelpVisible() bool {
	return m.help.Visible()
}

func (m Model) IsFiltering() bool {
	return m.list.FilterState() == list.Filtering
}

func (m Model) View() string {
	if !m.loaded {
		return fmt.Sprintf("\n  Loading instances for %s...", m.workflow.Name)
	}

	if m.help.Visible() {
		return m.help.View()
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
