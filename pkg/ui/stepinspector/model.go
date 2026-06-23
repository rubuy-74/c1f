package stepinspector

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/c1f/c1f/pkg/api"
	"github.com/c1f/c1f/pkg/models"
	"github.com/c1f/c1f/pkg/ui/common"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pane int

const (
	listPane pane = iota
	detailPane
)

type filter int

const (
	filterAll filter = iota
	filterFailed
	filterRunning
	filterSuccess
)

type Model struct {
	client      *api.Client
	workflow    models.Workflow
	instance    models.Instance
	loaded      bool
	width       int
	height      int
	cursor      int
	viewport    viewport.Model
	activePane  pane
	wrapping    bool
	filter      filter
	showRaw     bool
	rawViewport viewport.Model
	error       error
	errTimer    *time.Timer
}

func New(client *api.Client) Model {
	return Model{
		client:      client,
		rawViewport: viewport.New(0, 0),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetInstance(w models.Workflow, i models.Instance) tea.Cmd {
	m.workflow = w
	m.instance = i
	m.loaded = false
	m.cursor = 0
	m.activePane = listPane
	m.viewport = viewport.New(0, 0)
	m.rawViewport = viewport.New(0, 0)
	m.filter = filterAll
	m.showRaw = false
	m.error = nil
	return m.FetchInstance
}

type clearErrorMsg struct{}

func (m *Model) filteredSteps() []models.Step {
	if m.filter == filterAll {
		return m.instance.Steps
	}

	var filtered []models.Step
	for _, step := range m.instance.Steps {
		normalized := step.Status.Normalize()
		switch m.filter {
		case filterFailed:
			if normalized == models.StepStatusFailure {
				filtered = append(filtered, step)
			}
		case filterRunning:
			if normalized == models.StepStatusRunning {
				filtered = append(filtered, step)
			}
		case filterSuccess:
			if normalized == models.StepStatusSuccess {
				filtered = append(filtered, step)
			}
		}
	}
	return filtered
}

func (m *Model) updateViewport() {
	steps := m.filteredSteps()
	if len(steps) == 0 {
		m.viewport.SetContent("No steps matching filter.")
		return
	}

	if m.cursor >= len(steps) {
		m.cursor = 0
	}

	step := steps[m.cursor]
	var sb strings.Builder

	startedAt := step.DisplayStartedAt()
	finishedAt := step.DisplayFinishedAt()
	normalizedStatus := step.Status.Normalize()

	sb.WriteString(fmt.Sprintf("Step: %s\n", step.Name))
	sb.WriteString(fmt.Sprintf("Status: %s\n", normalizedStatus))

	if !startedAt.IsZero() {
		sb.WriteString(fmt.Sprintf("Started: %s\n", startedAt.Format("2006-01-02 15:04:05")))
	} else {
		sb.WriteString("Started: —\n")
	}

	if finishedAt != nil && !finishedAt.IsZero() {
		sb.WriteString(fmt.Sprintf("Finished: %s\n", finishedAt.Format("2006-01-02 15:04:05")))
		duration := finishedAt.Sub(startedAt)
		if startedAt.IsZero() {
			duration = 0
		}
		sb.WriteString(fmt.Sprintf("Duration: %s\n", common.FormatDuration(duration)))
	} else if normalizedStatus == models.StepStatusRunning {
		duration := time.Since(startedAt)
		if startedAt.IsZero() {
			duration = 0
		}
		sb.WriteString(fmt.Sprintf("Duration: %s (running)\n", common.FormatDuration(duration)))
	}

	sb.WriteString(fmt.Sprintf("Attempts: %d\n", step.Attempts.Count()))
	if step.Config != nil {
		retries := step.Config.Retries.String()
		if retries == "" {
			retries = "-"
		}
		timeout := step.Config.Timeout.String()
		if timeout == "" {
			timeout = "-"
		}
		sb.WriteString(fmt.Sprintf("Config: Retries=%s, Timeout=%s\n", retries, timeout))
	}

	if step.Output != nil {
		sb.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("OUTPUT:") + "\n")
		sb.WriteString(common.FormatStepOutput(*step.Output) + "\n")
	}

	if step.Error != nil {
		sb.WriteString("\n" + common.StatusRed.Render("ERROR:") + "\n")
		sb.WriteString(step.Error.Message + "\n")
		if step.Error.StackTrace != "" {
			sb.WriteString("\n" + common.StatusRed.Render("STACK TRACE:") + "\n")
			sb.WriteString(step.Error.StackTrace)
		}
	}

	content := sb.String()
	if m.wrapping {
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(content))
	} else {
		m.viewport.SetContent(content)
	}
}

func (m *Model) updateRawViewport() {
	data, err := json.MarshalIndent(m.instance, "", "  ")
	if err != nil {
		m.rawViewport.SetContent(fmt.Sprintf("Error marshaling raw data: %v", err))
		return
	}
	m.rawViewport.SetContent(string(data))
}

func (m Model) IsRawMode() bool {
	return m.showRaw
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		leftWidth := m.width / 3
		rightWidth := m.width - leftWidth - 2
		
		m.viewport.Width = rightWidth - 2
		m.viewport.Height = m.height - 4
		
		m.rawViewport.Width = m.width - 2
		m.rawViewport.Height = m.height - 2

		if !m.loaded {
			m.viewport.SetContent("Loading...")
		} else {
			m.updateViewport()
			m.updateRawViewport()
		}

	case instanceMsg:
		m.instance = msg.instance
		m.loaded = true
		m.updateViewport()
		m.updateRawViewport()

	case common.ErrorMsg:
		m.error = msg.Err
		if m.errTimer != nil {
			m.errTimer.Stop()
		}
		m.errTimer = time.AfterFunc(3*time.Second, func() {
			// We can't directly return a command from here easily without a channel or similar
			// but we can send a message back to the program if we had access to it.
			// Instead, we'll handle it in the next Update tick if we check for it, 
			// or just use a message.
		})
		// A better way with Bubble Tea is to return a command that sleeps and then returns a clearErrorMsg
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})

	case clearErrorMsg:
		m.error = nil
		return m, nil

	case tea.KeyMsg:
		if m.showRaw {
			switch msg.String() {
			case "v", "esc":
				m.showRaw = false
			default:
				m.rawViewport, cmd = m.rawViewport.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.loaded = false
			return m, m.FetchInstance
		case "f":
			m.filter = (m.filter + 1) % 4
			m.cursor = 0
			m.updateViewport()
			return m, nil
		case "v":
			m.showRaw = true
			m.updateRawViewport()
			return m, nil
		}

		if m.activePane == listPane {
			steps := m.filteredSteps()
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					m.updateViewport()
				}
			case "down", "j":
				if m.cursor < len(steps)-1 {
					m.cursor++
					m.updateViewport()
				}
			case "tab":
				m.activePane = detailPane
			}
		} else {
			switch msg.String() {
			case "tab":
				m.activePane = listPane
				return m, nil
			case "w":
				m.wrapping = !m.wrapping
				m.updateViewport()
				return m, nil
			}
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.loaded && m.error == nil {
		return fmt.Sprintf("\n  Loading instance %s...", m.instance.ID)
	}

	if m.showRaw {
		return lipgloss.NewStyle().
			Width(m.width - 2).
			Height(m.height - 2).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("205")).
			Render(m.rawViewport.View())
	}

	// Split pane layout
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 2 // -2 for borders/padding

	// Left pane: Step list
	var leftContent string
	steps := m.filteredSteps()
	if len(steps) == 0 {
		leftContent = "No steps matching filter."
	} else {
		for i, step := range steps {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			
			typeIcon := ""
			switch strings.ToLower(step.Type) {
			case "step.do", "do", "step":
				typeIcon = "[DO] "
			case "step.sleep", "sleep":
				typeIcon = "[SLP] "
			case "step.wait_for_event", "wait_for_event", "waitforevent":
				typeIcon = "[WFE] "
			default:
				if step.Type != "" {
					typeIcon = fmt.Sprintf("[%s] ", strings.ToUpper(step.Type))
				}
			}

			status := string(step.Status.Normalize())
			statusStyle := lipgloss.NewStyle()
			switch step.Status.Normalize() {
			case models.StepStatusSuccess:
				statusStyle = common.StatusGreen
			case models.StepStatusFailure:
				statusStyle = common.StatusRed
			case models.StepStatusRunning:
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			}

			if status == "" {
				status = "unknown"
			}

			leftContent += fmt.Sprintf("%s %s%s [%s]\n", cursor, typeIcon, step.Name, statusStyle.Render(status))
		}
	}

	// Filter indicator
	filterText := "Filter: All"
	switch m.filter {
	case filterFailed:
		filterText = "Filter: Failed"
	case filterRunning:
		filterText = "Filter: Running"
	case filterSuccess:
		filterText = "Filter: Success"
	}
	leftContent = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(filterText) + "\n\n" + leftContent

	// Error message (toast)
	if m.error != nil {
		errorToast := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("160")).
			Padding(0, 1).
			Render(fmt.Sprintf(" Error: %v ", m.error))
		leftContent = errorToast + "\n" + leftContent
	}

	leftBorderStyle := lipgloss.NormalBorder()
	leftBorderColor := lipgloss.Color("240")
	if m.activePane == listPane {
		leftBorderColor = lipgloss.Color("205")
	}

	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Height(m.height - 2).
		Border(leftBorderStyle, false, true, false, false).
		BorderForeground(leftBorderColor).
		Render(leftContent)

	// Right pane: Details viewport
	rightBorderColor := lipgloss.Color("240")
	if m.activePane == detailPane {
		rightBorderColor = lipgloss.Color("205")
	}

	header := m.renderHeader(rightWidth)
	headerHeight := lipgloss.Height(header)
	
	// Update viewport size based on header
	m.viewport.Width = rightWidth - 2
	m.viewport.Height = m.height - 2 - headerHeight - 2 // -2 for outer border

	rightPane := lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.height - 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(rightBorderColor).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View()))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

func (m Model) renderHeader(width int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240")).
		Width(width - 2)

	instanceID := fmt.Sprintf("Instance: %s", m.instance.ID)
	versionID := fmt.Sprintf("Version:  %s", m.instance.VersionID)
	trigger := fmt.Sprintf("Trigger:  %s", m.instance.Trigger.String())

	var totalWallTime string
	if len(m.instance.Steps) > 0 {
		start := m.instance.Steps[0].DisplayStartedAt()
		if start.IsZero() {
			start = m.instance.DisplayCreatedAt()
		}

		var end time.Time
		lastStep := m.instance.Steps[len(m.instance.Steps)-1]
		if finished := lastStep.DisplayFinishedAt(); finished != nil && !finished.IsZero() {
			end = *finished
		} else {
			end = time.Now()
		}

		if !start.IsZero() {
			totalWallTime = fmt.Sprintf("Wall Time: %s", common.FormatDuration(end.Sub(start)))
		} else {
			totalWallTime = "Wall Time: —"
		}
	} else {
		totalWallTime = "Wall Time: —"
	}

	return style.Render(fmt.Sprintf("%s\n%s\n%s\n%s", instanceID, versionID, trigger, totalWallTime))
}

type instanceMsg struct {
	instance models.Instance
}

func (m Model) FetchInstance() tea.Msg {
	instance, err := m.client.GetWorkflowInstance(context.Background(), m.workflow.Name, m.instance.ID)
	if err != nil {
		return common.ErrorMsg{Err: err}
	}
	return instanceMsg{instance: instance}
}
