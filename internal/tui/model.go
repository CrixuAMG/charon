package tui

import (
	"strings"

	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/kitty"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState int

const (
	stateList viewState = iota
	stateAdd
	stateEdit
	stateDelete
)

// Colors - Catppuccin Mocha inspired
var (
	colorSurface  = lipgloss.Color("#313244")
	colorOverlay  = lipgloss.Color("#45475a")
	colorText     = lipgloss.Color("#cdd6f4")
	colorSubtext  = lipgloss.Color("#a6adc8")
	colorLavender = lipgloss.Color("#b4befe")
	colorBlue     = lipgloss.Color("#89b4fa")
	colorGreen    = lipgloss.Color("#a6e3a1")
	colorPeach    = lipgloss.Color("#fab387")
	colorRed      = lipgloss.Color("#f38ba8")
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorLavender).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSubtext).
			MarginBottom(2)

	projectStyle = lipgloss.NewStyle().
			Padding(1, 2).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorOverlay)

	selectedProjectStyle = lipgloss.NewStyle().
				Padding(1, 2).
				MarginBottom(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorLavender).
				Background(colorSurface)

	projectNameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorText)

	selectedProjectNameStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(colorLavender)

	pathStyle = lipgloss.NewStyle().
			Foreground(colorSubtext)

	dockerBadgeStyle = lipgloss.NewStyle().
				Foreground(colorBlue).
				Bold(true)

	localBadgeStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	taskStyle = lipgloss.NewStyle().
			Foreground(colorPeach)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorSubtext).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true).
			MarginTop(1)

	successStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true).
			MarginTop(1)

	containerStyle = lipgloss.NewStyle().
			Padding(2, 4)

	formLabelStyle = lipgloss.NewStyle().
			Foreground(colorLavender).
			Bold(true).
			Width(12)

	formStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorLavender).
			Padding(1, 2).
			MarginTop(1)

	deleteBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorRed).
			Padding(1, 2).
			MarginTop(1)
)

type Model struct {
	config     *config.Config
	cursor     int
	message    string
	isError    bool
	width      int
	height     int
	quitting   bool
	state      viewState
	formInputs []textinput.Model
	formFocus  int
	editIndex  int
}

func NewModel(cfg *config.Config) Model {
	return Model{
		config: cfg,
		cursor: 0,
		state:  stateList,
	}
}

func (m *Model) initFormInputs(project *config.Project) {
	m.formInputs = make([]textinput.Model, 4)

	// Name input
	m.formInputs[0] = textinput.New()
	m.formInputs[0].Placeholder = "project-name"
	m.formInputs[0].CharLimit = 50
	m.formInputs[0].Width = 40

	// Path input
	m.formInputs[1] = textinput.New()
	m.formInputs[1].Placeholder = "~/path/to/project"
	m.formInputs[1].CharLimit = 200
	m.formInputs[1].Width = 40

	// Docker path input
	m.formInputs[2] = textinput.New()
	m.formInputs[2].Placeholder = "/var/www/html (leave empty for local)"
	m.formInputs[2].CharLimit = 200
	m.formInputs[2].Width = 40

	// Tasks input
	m.formInputs[3] = textinput.New()
	m.formInputs[3].Placeholder = "echo 'Hello World'; pwd;"
	m.formInputs[3].CharLimit = 500
	m.formInputs[3].Width = 40

	if project != nil {
		m.formInputs[0].SetValue(project.Name)
		m.formInputs[1].SetValue(project.Path)
		m.formInputs[2].SetValue(project.DockerPath)
		m.formInputs[3].SetValue(strings.Join(project.Tasks, ", "))
	}

	m.formFocus = 0
	m.formInputs[0].Focus()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateList:
			return m.updateList(msg)
		case stateAdd, stateEdit:
			return m.updateForm(msg)
		case stateDelete:
			return m.updateDelete(msg)
		}
	}

	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.cursor > 0 {
			m.cursor--
		}
		m.message = ""

	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.cursor < len(m.config.Projects)-1 {
			m.cursor++
		}
		m.message = ""

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
		if len(m.config.Projects) > 0 {
			project := m.config.Projects[m.cursor]
			err := kitty.OpenProject(project, m.config)
			if err != nil {
				m.message = "Error: " + err.Error()
				m.isError = true
			} else {
				m.quitting = true
				return m, tea.Quit
			}
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
		m.state = stateAdd
		m.initFormInputs(nil)
		m.message = ""
		return m, textinput.Blink

	case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
		if len(m.config.Projects) > 0 {
			m.state = stateEdit
			m.editIndex = m.cursor
			m.initFormInputs(&m.config.Projects[m.cursor])
			m.message = ""
			return m, textinput.Blink
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("d", "x"))):
		if len(m.config.Projects) > 0 {
			m.state = stateDelete
			m.message = ""
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("g", "home"))):
		m.cursor = 0
		m.message = ""

	case key.Matches(msg, key.NewBinding(key.WithKeys("G", "end"))):
		if len(m.config.Projects) > 0 {
			m.cursor = len(m.config.Projects) - 1
		}
		m.message = ""
	}

	return m, nil
}

func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.state = stateList
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "down"))):
		m.formInputs[m.formFocus].Blur()
		m.formFocus = (m.formFocus + 1) % len(m.formInputs)
		m.formInputs[m.formFocus].Focus()
		return m, textinput.Blink

	case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab", "up"))):
		m.formInputs[m.formFocus].Blur()
		m.formFocus--
		if m.formFocus < 0 {
			m.formFocus = len(m.formInputs) - 1
		}
		m.formInputs[m.formFocus].Focus()
		return m, textinput.Blink

	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
		return m.saveProject()
	}

	// Update the focused input
	var cmd tea.Cmd
	m.formInputs[m.formFocus], cmd = m.formInputs[m.formFocus].Update(msg)
	return m, cmd
}

func (m Model) saveProject() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.formInputs[0].Value())
	path := strings.TrimSpace(m.formInputs[1].Value())
	dockerPath := strings.TrimSpace(m.formInputs[2].Value())
	tasksStr := strings.TrimSpace(m.formInputs[3].Value())

	if name == "" {
		m.message = "Name is required"
		m.isError = true
		return m, nil
	}

	// Parse tasks
	var tasks []string
	if tasksStr != "" {
		for _, t := range strings.Split(tasksStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tasks = append(tasks, t)
			}
		}
	}

	project := config.Project{
		Name:       name,
		Path:       path,
		DockerPath: dockerPath,
		Tasks:      tasks,
	}

	if m.state == stateAdd {
		m.config.Projects = append(m.config.Projects, project)
		m.cursor = len(m.config.Projects) - 1
	} else {
		m.config.Projects[m.editIndex] = project
	}

	// Save to file
	if err := config.Save(m.config); err != nil {
		m.message = "Error saving: " + err.Error()
		m.isError = true
		return m, nil
	}

	m.message = "Project saved!"
	m.isError = false
	m.state = stateList
	return m, nil
}

func (m Model) updateDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("y", "enter"))):
		// Delete the project
		m.config.Projects = append(
			m.config.Projects[:m.cursor],
			m.config.Projects[m.cursor+1:]...,
		)
		if m.cursor >= len(m.config.Projects) && m.cursor > 0 {
			m.cursor--
		}

		// Save to file
		if err := config.Save(m.config); err != nil {
			m.message = "Error saving: " + err.Error()
			m.isError = true
		} else {
			m.message = "Project deleted"
			m.isError = false
		}
		m.state = stateList

	case key.Matches(msg, key.NewBinding(key.WithKeys("n", "esc", "q"))):
		m.state = stateList
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render(">>> Charon"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Navigate to your projects"))
	b.WriteString("\n")

	switch m.state {
	case stateList:
		b.WriteString(m.viewList())
	case stateAdd:
		b.WriteString(m.viewForm("Add New Project"))
	case stateEdit:
		b.WriteString(m.viewForm("Edit Project"))
	case stateDelete:
		b.WriteString(m.viewDelete())
	}

	return containerStyle.Render(b.String())
}

func (m Model) viewList() string {
	var b strings.Builder

	if len(m.config.Projects) == 0 {
		b.WriteString(pathStyle.Render("No projects configured. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		for i, project := range m.config.Projects {
			isSelected := i == m.cursor
			b.WriteString(m.renderProject(project, isSelected))
			b.WriteString("\n")
		}
	}

	// Message
	if m.message != "" {
		if m.isError {
			b.WriteString(errorStyle.Render(m.message))
		} else {
			b.WriteString(successStyle.Render(m.message))
		}
		b.WriteString("\n")
	}

	// Help at bottom
	help := []string{
		"j/k navigate",
		"enter open",
		"a add",
		"e edit",
		"d delete",
		"q quit",
	}
	b.WriteString(helpStyle.Render(strings.Join(help, " | ")))

	return b.String()
}

func (m Model) viewForm(title string) string {
	var b strings.Builder

	b.WriteString(projectNameStyle.Render(title))
	b.WriteString("\n\n")

	labels := []string{"Name:", "Path:", "Docker:", "Tasks:"}
	for i, input := range m.formInputs {
		label := formLabelStyle.Render(labels[i])
		b.WriteString(label + input.View())
		b.WriteString("\n")
	}

	formContent := formStyle.Render(b.String())

	var footer strings.Builder
	footer.WriteString(formContent)
	footer.WriteString("\n")

	// Message
	if m.message != "" {
		if m.isError {
			footer.WriteString(errorStyle.Render(m.message))
		}
		footer.WriteString("\n")
	}

	// Help at bottom
	help := helpStyle.Render("tab next | shift+tab prev | ctrl+s save | esc cancel")
	footer.WriteString(help)

	return footer.String()
}

func (m Model) viewDelete() string {
	var b strings.Builder

	if m.cursor < len(m.config.Projects) {
		project := m.config.Projects[m.cursor]
		b.WriteString(deleteBoxStyle.Render(
			"Delete project '" + project.Name + "'?\n\n" +
				"This cannot be undone.",
		))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("y confirm | n cancel"))

	return b.String()
}

func (m Model) renderProject(project config.Project, selected bool) string {
	var content strings.Builder

	// Project name with icon
	nameStyle := projectNameStyle
	if selected {
		nameStyle = selectedProjectNameStyle
	}

	icon := "[L] "
	if project.DockerPath != "" || m.config.DockerPath != "" {
		icon = "[D] "
	}

	name := nameStyle.Render(icon + project.Name)
	content.WriteString(name)
	content.WriteString("\n")

	// Path info
	var pathInfo string
	if project.DockerPath != "" {
		pathInfo = dockerBadgeStyle.Render("docker") + " " + pathStyle.Render(project.DockerPath+"/"+project.Name)
	} else if m.config.DockerPath != "" {
		pathInfo = dockerBadgeStyle.Render("docker") + " " + pathStyle.Render(m.config.DockerPath+"/"+project.Name)
	} else {
		pathInfo = localBadgeStyle.Render("local") + " " + pathStyle.Render(project.Path)
	}
	content.WriteString(pathInfo)
	content.WriteString("\n")

	// Tasks
	if len(project.Tasks) > 0 {
		tasks := make([]string, len(project.Tasks))
		for i, t := range project.Tasks {
			if len(t) > 20 {
				t = t[:17] + "..."
			}
			tasks[i] = t
		}
		taskLine := taskStyle.Render(">> " + strings.Join(tasks, " | "))
		content.WriteString(taskLine)
	}

	// Apply container style
	style := projectStyle
	if selected {
		style = selectedProjectStyle
	}

	return style.Render(content.String())
}
