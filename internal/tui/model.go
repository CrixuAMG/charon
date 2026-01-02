package tui

import (
	"strings"

	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/db"
	"github.com/crixuamg/charon/internal/kitty"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	config        *config.Config
	db            *db.DB
	cursor        int
	message       string
	isError       bool
	width         int
	height        int
	quitting      bool
	state         viewState
	formInputs    []textinput.Model
	formFocus     int
	editIndex     int
	searchQuery   string
	searchMode    bool
	currentSort   sortMode
	currentFilter filterMode
	styles        Styles
	keys          keyMap
}

func NewModel(cfg *config.Config) Model {
	theme := getTheme(cfg.Theme)
	database, err := db.Open()
	if err != nil {
		database = nil
	}

	return Model{
		config:        cfg,
		db:            database,
		cursor:        0,
		state:         stateList,
		currentSort:   sortByCustom,
		currentFilter: filterNone,
		styles:        NewStyles(theme),
		keys:          defaultKeyMap(),
	}
}

func getTheme(themeName string) Theme {
	switch strings.ToLower(themeName) {
	case "gruvbox":
		return GruvboxTheme()
	case "tokyonight", "tokyo-night":
		return TokyoNightTheme()
	default:
		return DefaultTheme()
	}
}

func (m *Model) initFormInputs(project *config.Project) {
	m.formInputs = make([]textinput.Model, 4)

	inputs := []struct {
		placeholder string
		charLimit   int
		value       string
	}{
		{"project-name", 50, ""},
		{"~/path/to/project", 200, ""},
		{"/var/www/html (leave empty for local)", 200, ""},
		{"echo 'Hello World'; pwd;", 500, ""},
	}

	if project != nil {
		inputs[0].value = project.Name
		inputs[1].value = project.Path
		inputs[2].value = strings.Join(project.Tasks, ", ")
	}

	for i, cfg := range inputs {
		m.formInputs[i] = textinput.New()
		m.formInputs[i].Placeholder = cfg.placeholder
		m.formInputs[i].CharLimit = cfg.charLimit
		m.formInputs[i].Width = 40
		if cfg.value != "" {
			m.formInputs[i].SetValue(cfg.value)
		}
	}

	m.formFocus = 0
	m.formInputs[0].Focus()
}

func (m Model) getFilteredProjects() []projectWithIndex {
	filtered := filterProjects(m.config.Projects, m.config, m.currentFilter, m.searchQuery)
	sortProjects(filtered, m.currentSort, m.db)
	return filtered
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Cleanup() {
	if m.db != nil {
		m.db.Close()
	}
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
	if m.searchMode {
		return m.updateSearch(msg)
	}

	filtered := m.getFilteredProjects()

	switch {
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		m.message = ""

	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(filtered)-1 {
			m.cursor++
		}
		m.message = ""

	case key.Matches(msg, m.keys.Enter):
		if len(filtered) > 0 {
			project := filtered[m.cursor].project
			if err := kitty.OpenProject(project, m.config); err != nil {
				m.message = "Error: " + err.Error()
				m.isError = true
			} else {
				if m.db != nil {
					_ = m.db.RecordAccess(project.Name)
				}
				m.quitting = true
				return m, tea.Quit
			}
		}

	case key.Matches(msg, m.keys.Search):
		m.searchMode = true
		m.searchQuery = ""
		m.cursor = 0
		m.message = ""

	case key.Matches(msg, m.keys.Sort):
		m.currentSort = (m.currentSort + 1) % 3
		m.cursor = 0
		m.message = ""

	case key.Matches(msg, m.keys.Filter):
		m.currentFilter = (m.currentFilter + 1) % 3
		m.cursor = 0
		m.message = ""

	case key.Matches(msg, m.keys.Add):
		m.state = stateAdd
		m.initFormInputs(nil)
		m.message = ""
		return m, textinput.Blink

	case key.Matches(msg, m.keys.Edit):
		if len(filtered) > 0 {
			m.state = stateEdit
			m.editIndex = getOriginalIndex(filtered, m.cursor)
			m.initFormInputs(&m.config.Projects[m.editIndex])
			m.message = ""
			return m, textinput.Blink
		}

	case key.Matches(msg, m.keys.Delete):
		if len(filtered) > 0 {
			m.state = stateDelete
			m.message = ""
		}

	case key.Matches(msg, m.keys.First):
		m.cursor = 0
		m.message = ""

	case key.Matches(msg, m.keys.Last):
		if len(filtered) > 0 {
			m.cursor = len(filtered) - 1
		}
		m.message = ""
	}

	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searchMode = false
		m.searchQuery = ""
		m.cursor = 0
		return m, nil

	case tea.KeyEnter:
		m.searchMode = false
		m.cursor = 0
		return m, nil

	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.cursor = 0
		} else {
			m.searchMode = false
		}

	case tea.KeyRunes:
		m.searchQuery += string(msg.Runes)
		m.cursor = 0
	}

	return m, nil
}

func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.state = stateList
		return m, nil

	case key.Matches(msg, m.keys.NextField):
		m.formInputs[m.formFocus].Blur()
		m.formFocus = (m.formFocus + 1) % len(m.formInputs)
		m.formInputs[m.formFocus].Focus()
		return m, textinput.Blink

	case key.Matches(msg, m.keys.PrevField):
		m.formInputs[m.formFocus].Blur()
		m.formFocus--
		if m.formFocus < 0 {
			m.formFocus = len(m.formInputs) - 1
		}
		m.formInputs[m.formFocus].Focus()
		return m, textinput.Blink

	case key.Matches(msg, m.keys.Save):
		return m.saveProject()
	}

	var cmd tea.Cmd
	m.formInputs[m.formFocus], cmd = m.formInputs[m.formFocus].Update(msg)
	return m, cmd
}

func (m Model) saveProject() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.formInputs[0].Value())
	path := strings.TrimSpace(m.formInputs[1].Value())
	tasksStr := strings.TrimSpace(m.formInputs[3].Value())

	if name == "" {
		m.message = "Name is required"
		m.isError = true
		return m, nil
	}

	var tasks []string
	if tasksStr != "" {
		for _, t := range strings.Split(tasksStr, ",") {
			if t = strings.TrimSpace(t); t != "" {
				tasks = append(tasks, t)
			}
		}
	}

	project := config.Project{
		Name:  name,
		Path:  path,
		Tasks: tasks,
	}

	if m.state == stateAdd {
		m.config.Projects = append(m.config.Projects, project)
		m.cursor = len(m.config.Projects) - 1
	} else {
		m.config.Projects[m.editIndex] = project
	}

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
	filtered := m.getFilteredProjects()

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("y", "enter"))):
		if m.cursor < len(filtered) {
			originalIdx := getOriginalIndex(filtered, m.cursor)
			m.config.Projects = append(
				m.config.Projects[:originalIdx],
				m.config.Projects[originalIdx+1:]...,
			)
			if m.cursor >= len(m.config.Projects) && m.cursor > 0 {
				m.cursor--
			}

			if err := config.Save(m.config); err != nil {
				m.message = "Error saving: " + err.Error()
				m.isError = true
			} else {
				m.message = "Project deleted"
				m.isError = false
			}
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

	var content strings.Builder

	content.WriteString(m.styles.Title.Render("⚡ Charon"))
	content.WriteString(" ")
	content.WriteString(m.styles.Subtitle.Render("Project Navigator"))
	content.WriteString("\n\n")

	var mainContent string
	switch m.state {
	case stateList:
		mainContent = m.viewList()
	case stateAdd:
		mainContent = m.viewForm("Add New Project")
	case stateEdit:
		mainContent = m.viewForm("Edit Project")
	case stateDelete:
		mainContent = m.viewDelete()
	}

	content.WriteString(mainContent)

	helpBar := m.renderHelpBar()

	// Calculate available height for content
	headerHeight := 2 // Title + spacing
	helpHeight := lipgloss.Height(helpBar)
	containerPadding := 2 // top + bottom padding
	availableHeight := m.height - headerHeight - helpHeight - containerPadding

	// Add spacing to push help bar to bottom
	currentHeight := lipgloss.Height(mainContent)
	if currentHeight < availableHeight {
		content.WriteString(strings.Repeat("\n", availableHeight-currentHeight))
	}

	content.WriteString("\n")
	content.WriteString(helpBar)

	return m.styles.Container.Render(content.String())
}
