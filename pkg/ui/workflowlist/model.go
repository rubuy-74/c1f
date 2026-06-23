package workflowlist

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/c1f/c1f/pkg/api"
	"github.com/c1f/c1f/pkg/models"
	"github.com/c1f/c1f/pkg/ui/common"
	"github.com/c1f/c1f/pkg/ui/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	workflow models.Workflow
}

func (i item) Title() string { return i.workflow.Name }
func (i item) Description() string {
	createdAt := i.workflow.DisplayCreatedAt()
	created := createdAt.Format("2006-01-02 15:04")
	if createdAt.IsZero() {
		created = "—"
	}
	id := i.workflow.ID
	if id == "" {
		id = "—"
	}
	return fmt.Sprintf("ID: %s | Created: %s", id, created)
}
func (i item) FilterValue() string { return i.workflow.Name }

type Model struct {
	list                        list.Model
	help                        help.Model
	client                      *api.Client
	loaded                      bool
	showAnalytics                 bool
	analyticsLoading              bool
	analyticsError                error
	analyticsData                 *api.AnalyticsData
	selectedWorkflowForAnalytics  models.Workflow
	graphqlClient                 *api.GraphQLClient
	analyticsTimeRange            string
	width                         int
	height                        int
}

func New(client *api.Client) Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#A3A3A3"))
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF87D7")).Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB2D7"))

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Workflows"
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
				{Key: "/", Description: "Filter workflows"},
			},
		},
		{
			Title: "Actions",
			Bindings: []help.Binding{
				{Key: "enter", Description: "View instances"},
				{Key: "a", Description: "Toggle analytics"},
				{Key: "r", Description: "Refresh analytics"},
				{Key: "t", Description: "Cycle time range"},
			},
		},
	})

	return Model{
		list:                  l,
		help:                  h,
		client:                client,
		graphqlClient:         api.NewGraphQLClient(client.APIToken(), client.AccountID()),
		analyticsTimeRange:    "24h",
		showAnalytics:         true,
	}
}

func (m Model) Init() tea.Cmd {
	return m.FetchWorkflows
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.help.SetSize(msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
	case workflowsMsg:
		items := make([]list.Item, len(msg.workflows))
		for i, w := range msg.workflows {
			items[i] = list.Item(item{workflow: w})
		}
		m.list.SetItems(items)
		m.loaded = true
		if m.showAnalytics && len(items) > 0 {
			first := items[0].(item)
			m.SetAnalyticsWorkflow(first.workflow)
			m.analyticsLoading = true
			m.analyticsError = nil
			m.analyticsData = nil
			datetime_geq, datetime_leq := m.timeRangeToDatetimes(m.analyticsTimeRange)
			return m, m.fetchAnalyticsCmd(first.workflow.Name, datetime_geq, datetime_leq)
		}
	case analyticsMsg:
		if msg.err != nil {
			m.analyticsError = msg.err
			m.analyticsLoading = false
		} else {
			m.analyticsData = msg.data
			m.analyticsError = nil
			m.analyticsLoading = false
		}
	}

	if m.help.Visible() {
		var helpCmd tea.Cmd
		m.help, helpCmd = m.help.Update(msg)
		return m, helpCmd
	}

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
						return common.WorkflowSelectedMsg{Workflow: i.workflow}
					}
				}
			case "a":
				return m, m.ToggleAnalytics()
			case "r":
				if m.showAnalytics {
					workflowName := m.selectedWorkflowForAnalytics.Name
					if workflowName == "" {
						if i, ok := m.list.SelectedItem().(item); ok {
							workflowName = i.workflow.Name
							m.SetAnalyticsWorkflow(i.workflow)
						}
					}
					if workflowName != "" {
						m.analyticsLoading = true
						datetime_geq, datetime_leq := m.timeRangeToDatetimes(m.analyticsTimeRange)
						return m, m.fetchAnalyticsCmd(workflowName, datetime_geq, datetime_leq)
					}
				}
			case "t":
				if m.showAnalytics {
					m.analyticsTimeRange = m.cycleTimeRange()
					workflowName := m.selectedWorkflowForAnalytics.Name
					if workflowName == "" {
						if i, ok := m.list.SelectedItem().(item); ok {
							workflowName = i.workflow.Name
							m.SetAnalyticsWorkflow(i.workflow)
						}
					}
					m.analyticsLoading = true
					datetime_geq, datetime_leq := m.timeRangeToDatetimes(m.analyticsTimeRange)
					return m, m.fetchAnalyticsCmd(workflowName, datetime_geq, datetime_leq)
				}
			case "esc":
				if m.showAnalytics {
					m.showAnalytics = false
					return m, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	prevSelectedName := m.selectedWorkflowForAnalytics.Name
	m.list, cmd = m.list.Update(msg)
	if m.showAnalytics {
		if i, ok := m.list.SelectedItem().(item); ok {
			if i.workflow.Name != prevSelectedName {
				m.SetAnalyticsWorkflow(i.workflow)
				m.analyticsLoading = true
				m.analyticsError = nil
				m.analyticsData = nil
				datetime_geq, datetime_leq := m.timeRangeToDatetimes(m.analyticsTimeRange)
				return m, m.fetchAnalyticsCmd(i.workflow.Name, datetime_geq, datetime_leq)
			}
		}
	}
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
		return "\n  Loading workflows..."
	}

	if m.help.Visible() {
		return m.help.View()
	}

	listView := m.list.View()

	if m.showAnalytics {
		analyticsPanel := m.renderAnalyticsPanel()
		listView = listView + "\n\n" + analyticsPanel
	}

	return listView
}

func (m Model) renderAnalyticsPanel() string {
	borderColor := lipgloss.Color("#FF87D7")
	panelWidth := m.width - 4
	if panelWidth < 20 {
		panelWidth = 40
	}

	workflowName := m.selectedWorkflowForAnalytics.Name
	if workflowName == "" {
		workflowName = "—"
	}

	timeRangeLabel := m.analyticsTimeRange

	var content string
	if m.analyticsLoading {
		content = "Loading..."
	} else if m.analyticsError != nil {
		content = "Failed to load — press r to retry"
	} else if m.analyticsData != nil {
		data := m.analyticsData
		sparkWidth := 30
		if panelWidth > 50 {
			sparkWidth = panelWidth - 30
		}
		invSpark := common.RenderSparkline(data.InvocationBuckets, sparkWidth)
		wallSpark := common.RenderSparkline(data.WallTimeBuckets, sparkWidth)
		failSpark := common.RenderSparkline(data.FailBuckets, sparkWidth)

		invCount := formatIntWithCommas(data.InvocationCount)
		wallTime := common.FormatDurationMs(data.AvgWallTimeMs)
		failPct := fmt.Sprintf("%.1f%%", data.FailRatio*100)

		content = fmt.Sprintf("Invocations [%s]   %s   %s\n", timeRangeLabel, invCount, invSpark)
		content += fmt.Sprintf("Avg Wall-time       %s   %s\n", wallTime, wallSpark)
		content += fmt.Sprintf("Failure Rate        %s   %s\n", failPct, failSpark)
		content += "\n"
		content += fmt.Sprintf("CPU Time: %sms total", formatIntWithCommas(data.SumWallTimeMs))
		content += fmt.Sprintf("  |  ~%s estimated cost", m.formatCost(data.SumWallTimeMs))
		content += "\n"
	} else {
		content = "No data available"
	}

	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB2D7")).Bold(true)
	timeRangeBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#F25D94")).
		Padding(0, 1).
		Render("[" + timeRangeLabel + "]")
	header := headerStyle.Render("analytics: "+workflowName) + "  " + timeRangeBadge

	contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A3A3A3"))

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(panelWidth)

	panelContent := header + "\n\n" + contentStyle.Render(content) + "\n" + hintStyle.Render("[t] cycle range  [r] refresh  [Esc] close")

	return panel.Render(panelContent)
}

func formatIntWithCommas(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var result strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(r)
	}
	return result.String()
}

func (m Model) formatCost(sumWallTimeMs int64) string {
	costPer100Ms := 0.000001
	if envCost := os.Getenv("CLOUDFLARE_CPU_COST_PER_100MS"); envCost != "" {
		var parsed float64
		fmt.Sscanf(envCost, "%f", &parsed)
		if parsed > 0 {
			costPer100Ms = parsed
		}
	}
	cost := (float64(sumWallTimeMs) / 100.0) * costPer100Ms
	if cost < 0.005 && cost > 0 {
		return "$0.00"
	}
	return fmt.Sprintf("$%.2f", cost)
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

type analyticsMsg struct {
	data *api.AnalyticsData
	err  error
}

func (m Model) SetAnalyticsWorkflow(workflow models.Workflow) {
	m.selectedWorkflowForAnalytics = workflow
}

func (m Model) ToggleAnalytics() tea.Cmd {
	m.showAnalytics = !m.showAnalytics
	if m.showAnalytics {
		m.analyticsLoading = true
		m.analyticsError = nil
		m.analyticsData = nil
		if i, ok := m.list.SelectedItem().(item); ok {
			m.SetAnalyticsWorkflow(i.workflow)
			datetime_geq, datetime_leq := m.timeRangeToDatetimes(m.analyticsTimeRange)
			return m.fetchAnalyticsCmd(i.workflow.Name, datetime_geq, datetime_leq)
		}
	}
	return nil
}

func (m Model) cycleTimeRange() string {
	switch m.analyticsTimeRange {
	case "24h":
		return "7d"
	case "7d":
		return "30d"
	default:
		return "24h"
	}
}

func (m Model) timeRangeToDatetimes(timeRange string) (string, string) {
	now := time.Now()
	var datetime_geq string
	switch timeRange {
	case "7d":
		datetime_geq = now.Add(-7 * 24 * time.Hour).Format(time.RFC3339)
	case "30d":
		datetime_geq = now.Add(-30 * 24 * time.Hour).Format(time.RFC3339)
	default:
		datetime_geq = now.Add(-24 * time.Hour).Format(time.RFC3339)
	}
	datetime_leq := now.Format(time.RFC3339)
	return datetime_geq, datetime_leq
}

func (m Model) fetchAnalyticsCmd(workflowName string, datetime_geq, datetime_leq string) tea.Cmd {
	return func() tea.Msg {
		accountId := m.client.AccountID()
		var wg sync.WaitGroup
		var mu sync.Mutex

		invCount, invBuckets := int64(0), []float64(nil)
		wallAvg, wallSum := float64(0), int64(0)
		wallBuckets := []float64(nil)
		failRatio, failBuckets := float64(0), []float64(nil)

		wg.Add(3)
		go func() {
			defer wg.Done()
			count, buckets, err := m.graphqlClient.FetchInvocationCount(context.Background(), workflowName, accountId, datetime_geq, datetime_leq)
			mu.Lock()
			invCount = count
			invBuckets = buckets
			if err != nil {
				invCount = 0
				invBuckets = nil
			}
			mu.Unlock()
		}()
		go func() {
			defer wg.Done()
			avgMs, sumMs, buckets, err := m.graphqlClient.FetchWallTime(context.Background(), workflowName, accountId, datetime_geq, datetime_leq)
			mu.Lock()
			wallAvg, wallSum, wallBuckets = avgMs, sumMs, buckets
			if err != nil {
				wallAvg = 0
				wallSum = 0
				wallBuckets = nil
			}
			mu.Unlock()
		}()
		go func() {
			defer wg.Done()
			ratio, buckets, err := m.graphqlClient.FetchFailureRate(context.Background(), workflowName, accountId, datetime_geq, datetime_leq)
			mu.Lock()
			failRatio, failBuckets = ratio, buckets
			if err != nil {
				failRatio = 0
				failBuckets = nil
			}
			mu.Unlock()
		}()
		wg.Wait()

		return analyticsMsg{data: &api.AnalyticsData{
			InvocationCount:   invCount,
			InvocationBuckets: invBuckets,
			AvgWallTimeMs:     wallAvg,
			SumWallTimeMs:     wallSum,
			WallTimeBuckets:   wallBuckets,
			FailRatio:         failRatio,
			FailBuckets:       failBuckets,
			TimeRange:         m.analyticsTimeRange,
		}}
	}
}