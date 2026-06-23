package help

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Binding struct {
	Key         string
	Description string
}

type Section struct {
	Title    string
	Bindings []Binding
}

type Model struct {
	visible  bool
	sections []Section
	width    int
	height   int
}

func New(sections []Section) Model {
	return Model{
		sections: sections,
	}
}

func (m *Model) Toggle() {
	m.visible = !m.visible
}

func (m *Model) Show() {
	m.visible = true
}

func (m *Model) Hide() {
	m.visible = false
}

func (m Model) Visible() bool {
	return m.visible
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if !m.visible {
			return m, nil
		}
		switch msg.String() {
		case "?", "esc", "q":
			m.visible = false
		}
	}
	return m, nil
}

func (m Model) View() string {
	if !m.visible {
		return ""
	}

	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF87D7"))

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A3A3A3"))

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB2D7")).
		Bold(true)

	sb.WriteString(titleStyle.Render("Help") + "\n\n")

	for _, section := range m.sections {
		sb.WriteString(sectionStyle.Render(section.Title) + "\n")
		for _, b := range section.Bindings {
			sb.WriteString("  " + keyStyle.Render(padKey(b.Key)) + "  " + descStyle.Render(b.Description) + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(descStyle.Render("Press ? or Esc to close") + "\n")

	content := sb.String()
	contentWidth := lipgloss.Width(content)
	contentHeight := lipgloss.Height(content)

	if m.width > 0 && m.height > 0 {
		// Center the help panel
		leftPadding := (m.width - contentWidth) / 2
		if leftPadding < 0 {
			leftPadding = 0
		}
		topPadding := (m.height - contentHeight) / 2
		if topPadding < 0 {
			topPadding = 0
		}

		panel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF87D7")).
			Padding(1, 2).
			Render(content)

		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			panel,
		)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF87D7")).
		Padding(1, 2).
		Render(content)
}

func padKey(k string) string {
	// Add a binding-like display using brackets for special keys.
	if k == "?" || k == "/" || k == "esc" || k == "enter" || k == "tab" {
		return strings.ToUpper(k)
	}
	return k
}

// WithBindings builds sections from bubbletea key.Binding slices for consistency.
func WithBindings(title string, bindings []key.Binding) Section {
	var b []Binding
	for _, kb := range bindings {
		b = append(b, Binding{
			Key:         kb.Help().Key,
			Description: kb.Help().Desc,
		})
	}
	return Section{Title: title, Bindings: b}
}
