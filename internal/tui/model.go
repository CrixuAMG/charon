package tui

import (
	"strings"

	"github.com/crixuamg/charon/internal/browser"
	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/db"
	gitinfo "github.com/crixuamg/charon/internal/git"
	"github.com/crixuamg/charon/internal/kitty"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// gitInfoMsg is sent asynchronously when git info for a project is loaded.
type gitInfoMsg struct {
	projectName string
	info        gitinfo.Info
}

type layoutMode int

const (
	layoutCard layoutMode = iota
	layoutCardCompact
	layoutTable
	layoutTableCompact
)

func (l layoutMode) String() string {
	switch l {
	case layoutCard:
		return "card"
	case layoutCardCompact:
		return "card-compact"
	case layoutTable:
		return "table"
	case layoutTableCompact:
		return "table-compact"
	default:
		return "card"
	}
}

type Model struct {
	config           *config.Config
	db               *db.DB
	cursor           int
	message          string
	isError          bool
	width            int
	height           int
	quitting         bool
	state            viewState
	formInputs       []textinput.Model
	formFocus        int
	editIndex        int
	searchQuery      string
	searchMode       bool
	searchHistory    []string
	searchHistoryIdx int
	currentSort      sortMode
	currentFilter    filterMode
	activeTag        string
	currentLayout    layoutMode
	styles           Styles
	keys             keyMap
	// Form-specific fields
	formPinned       bool
	formExecType     string // "local" or "docker"
	formTasksFrom    string // selected taskset name or empty
	formTasksetNames []string
	scrollOffset     int
	gitInfos         map[string]gitinfo.Info
	// Taskset management fields
	tasksetCursor   int
	editTasksetName string // name of taskset being edited
}

func NewModel(cfg *config.Config) Model {
	theme := getTheme(cfg.Theme)
	database, err := db.Open()
	if err != nil {
		database = nil
	}

	layout := getLayout(cfg.Interface.Layout)

	return Model{
		config:        cfg,
		db:            database,
		cursor:        0,
		state:         stateList,
		currentSort:   sortByCustom,
		currentFilter: filterNone,
		currentLayout: layout,
		styles:        NewStyles(theme),
		keys:          defaultKeyMap(),
	}
}

func getLayout(layoutName string) layoutMode {
	switch strings.ToLower(layoutName) {
	case "card":
		return layoutCard
	case "card-compact":
		return layoutCardCompact
	case "table":
		return layoutTable
	case "table-compact":
		return layoutTableCompact
	default:
		return layoutCard
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
	// Form fields: Name(0), Path(1), Container(2), Tasks(3), Tags(4)
	m.formInputs = make([]textinput.Model, 5)

	inputs := []struct {
		placeholder string
		charLimit   int
		value       string
	}{
		{"project-name", 50, ""},
		{"~/path/to/project", 200, ""},
		{"container-name", 100, ""},
		{"echo 'Hello World'; pwd;", 500, ""},
		{"work, hobby, infra", 200, ""},
	}

	// Initialize form state
	m.formPinned = false
	m.formExecType = "local"
	m.formTasksFrom = ""

	// Build list of available taskset names
	m.formTasksetNames = []string{""}
	for name := range m.config.TaskSets {
		m.formTasksetNames = append(m.formTasksetNames, name)
	}

	if project != nil {
		inputs[0].value = project.Name
		inputs[1].value = project.Path
		inputs[3].value = strings.Join(project.Tasks, ", ")
		inputs[4].value = strings.Join(project.Tags, ", ")

		m.formPinned = project.Pinned
		m.formTasksFrom = project.TasksFrom

		if project.Execution != nil {
			m.formExecType = project.Execution.Type
			inputs[2].value = project.Execution.Container
		}
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

// Fixed line counts that consume vertical space outside the project list.
const (
	viewportHeaderLines  = 2 // title + blank line
	viewportStatusLines  = 2 // status bar + blank line
	viewportHelpLines    = 1
	viewportPaddingLines = 2 // container top + bottom padding
	viewportMessageLines = 2 // reserve for status/error message
	viewportOverhead     = viewportHeaderLines + viewportStatusLines + viewportHelpLines + viewportPaddingLines + viewportMessageLines
)

// viewportSize returns the number of list items that fit in the terminal.
func (m Model) viewportSize() int {
	if m.height == 0 {
		return 20
	}
	available := m.height - viewportOverhead
	if available < 1 {
		return 1
	}
	switch m.currentLayout {
	case layoutCardCompact:
		// 1 content line + 1 separator = 2 lines per item
		if available < 2 {
			return 1
		}
		return available / 2
	case layoutTable, layoutTableCompact:
		// 2 header lines + 1 line per row
		available -= 2
		if available < 1 {
			return 1
		}
		return available
	default: // layoutCard
		// 2 content lines (name+path, tasks) + 1 separator = 3 lines per item
		if available < 3 {
			return 1
		}
		return available / 3
	}
}

// adjustScroll keeps cursor visible within the viewport.
func adjustScroll(cursor, offset, viewSize int) int {
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+viewSize {
		return cursor - viewSize + 1
	}
	return offset
}

func (m Model) getFilteredProjects() []projectWithIndex {
	filtered := filterProjects(m.config.Projects, m.config, m.currentFilter, m.activeTag, m.searchQuery)
	sortProjects(filtered, m.currentSort, m.db)
	return filtered
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, p := range m.config.Projects {
		p := p
		cmds = append(cmds, func() tea.Msg {
			return gitInfoMsg{projectName: p.Name, info: gitinfo.GetInfo(p.Path)}
		})
	}
	return tea.Batch(cmds...)
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

	case gitInfoMsg:
		if m.gitInfos == nil {
			m.gitInfos = make(map[string]gitinfo.Info)
		}
		m.gitInfos[msg.projectName] = msg.info
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateList:
			return m.updateList(msg)
		case stateAdd, stateEdit:
			return m.updateForm(msg)
		case stateDelete:
			return m.updateDelete(msg)
		case stateTasksetList:
			return m.updateTasksetList(msg)
		case stateTasksetAdd, stateTasksetEdit:
			return m.updateTasksetForm(msg)
		case stateTasksetDelete:
			return m.updateTasksetDelete(msg)
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
		m.scrollOffset = adjustScroll(m.cursor, m.scrollOffset, m.viewportSize())
		m.message = ""

	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(filtered)-1 {
			m.cursor++
		}
		m.scrollOffset = adjustScroll(m.cursor, m.scrollOffset, m.viewportSize())
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
		m.currentSort = (m.currentSort + 1) % 4
		m.cursor = 0
		m.scrollOffset = 0
		m.message = ""

	case key.Matches(msg, m.keys.Filter):
		m.currentFilter = (m.currentFilter + 1) % 3
		m.cursor = 0
		m.scrollOffset = 0
		m.message = ""

	case key.Matches(msg, m.keys.Layout):
		m.currentLayout = (m.currentLayout + 1) % 4
		m.message = ""
		// Save layout preference to config
		m.config.Interface.Layout = m.currentLayout.String()
		if err := config.Save(m.config); err != nil {
			m.message = "Failed to save layout preference"
			m.isError = true
		}

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

	case key.Matches(msg, m.keys.TagFilter):
		tags := allTags(m.config.Projects)
		if len(tags) == 0 {
			break
		}
		if m.activeTag == "" {
			m.activeTag = tags[0]
		} else {
			current := -1
			for i, t := range tags {
				if t == m.activeTag {
					current = i
					break
				}
			}
			if current == -1 || current == len(tags)-1 {
				m.activeTag = ""
			} else {
				m.activeTag = tags[current+1]
			}
		}
		m.cursor = 0
		m.scrollOffset = 0
		m.message = ""

	case key.Matches(msg, m.keys.Tasksets):
		m.state = stateTasksetList
		m.tasksetCursor = 0
		m.message = ""

	case key.Matches(msg, m.keys.First):
		m.cursor = 0
		m.scrollOffset = 0
		m.message = ""

	case key.Matches(msg, m.keys.Last):
		if len(filtered) > 0 {
			m.cursor = len(filtered) - 1
			m.scrollOffset = adjustScroll(m.cursor, m.scrollOffset, m.viewportSize())
		}
		m.message = ""

	case key.Matches(msg, m.keys.RevealPath):
		if len(filtered) > 0 {
			path := filtered[m.cursor].project.Path
			if err := browser.Open(path); err != nil {
				m.message = "Failed to open: " + err.Error()
				m.isError = true
			} else {
				m.message = ""
			}
		}
	}

	return m, nil
}

const maxSearchHistory = 20

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searchMode = false
		m.searchQuery = ""
		m.searchHistoryIdx = 0
		m.cursor = 0
		return m, nil

	case tea.KeyEnter:
		if m.searchQuery != "" {
			m.searchHistory = appendUniqueHistory(m.searchHistory, m.searchQuery, maxSearchHistory)
		}
		m.searchMode = false
		m.searchHistoryIdx = 0
		m.cursor = 0
		return m, nil

	case tea.KeyUp:
		if m.searchHistoryIdx < len(m.searchHistory) {
			m.searchHistoryIdx++
			m.searchQuery = m.searchHistory[len(m.searchHistory)-m.searchHistoryIdx]
			m.cursor = 0
			m.scrollOffset = 0
		}

	case tea.KeyDown:
		if m.searchHistoryIdx > 1 {
			m.searchHistoryIdx--
			m.searchQuery = m.searchHistory[len(m.searchHistory)-m.searchHistoryIdx]
		} else if m.searchHistoryIdx == 1 {
			m.searchHistoryIdx = 0
			m.searchQuery = ""
		}
		m.cursor = 0
		m.scrollOffset = 0

	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.searchHistoryIdx = 0
			m.cursor = 0
			m.scrollOffset = 0
		} else {
			m.searchMode = false
		}

	case tea.KeyRunes:
		m.searchQuery += string(msg.Runes)
		m.searchHistoryIdx = 0
		m.cursor = 0
		m.scrollOffset = 0
	}

	return m, nil
}

func appendUniqueHistory(history []string, query string, max int) []string {
	if len(history) > 0 && history[len(history)-1] == query {
		return history
	}
	history = append(history, query)
	if len(history) > max {
		history = history[1:]
	}
	return history
}

// getInputIndex maps logical form field index to formInputs array index.
// Returns -1 if the field is not a text input.
// Field order: Name(0), Path(1), Pinned(2), ExecType(3), Container(4), TasksFrom(5), Tasks(6), Tags(7)
func (m Model) getInputIndex(fieldIndex int) int {
	switch fieldIndex {
	case 0:
		return 0 // Name
	case 1:
		return 1 // Path
	case 4:
		return 2 // Container
	case 6:
		return 3 // Tasks
	case 7:
		return 4 // Tags
	default:
		return -1
	}
}

func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Total fields: Name(0), Path(1), Pinned(2), ExecType(3), Container(4), TasksFrom(5), Tasks(6), Tags(7)
	totalFields := 8

	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.state = stateList
		return m, nil

	case key.Matches(msg, m.keys.NextField):
		// Blur current field if it's a text input
		if inputIdx := m.getInputIndex(m.formFocus); inputIdx >= 0 {
			m.formInputs[inputIdx].Blur()
		}
		m.formFocus = (m.formFocus + 1) % totalFields
		// Skip container field if execution type is local
		if m.formFocus == 4 && m.formExecType != "docker" {
			m.formFocus = (m.formFocus + 1) % totalFields
		}
		// Focus new field if it's a text input
		if inputIdx := m.getInputIndex(m.formFocus); inputIdx >= 0 {
			m.formInputs[inputIdx].Focus()
		}
		return m, textinput.Blink

	case key.Matches(msg, m.keys.PrevField):
		// Blur current field if it's a text input
		if inputIdx := m.getInputIndex(m.formFocus); inputIdx >= 0 {
			m.formInputs[inputIdx].Blur()
		}
		m.formFocus--
		if m.formFocus < 0 {
			m.formFocus = totalFields - 1
		}
		// Skip container field if execution type is local
		if m.formFocus == 4 && m.formExecType != "docker" {
			m.formFocus--
			if m.formFocus < 0 {
				m.formFocus = totalFields - 1
			}
		}
		// Focus new field if it's a text input
		if inputIdx := m.getInputIndex(m.formFocus); inputIdx >= 0 {
			m.formInputs[inputIdx].Focus()
		}
		return m, textinput.Blink

	case key.Matches(msg, m.keys.Save):
		return m.saveProject()
	}

	// Handle field-specific input
	switch m.formFocus {
	case 2: // Pinned (boolean toggle)
		if msg.Type == tea.KeySpace || msg.Type == tea.KeyEnter {
			m.formPinned = !m.formPinned
		}
	case 3: // Execution Type (selection)
		if msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight || msg.Type == tea.KeySpace {
			if m.formExecType == "local" {
				m.formExecType = "docker"
			} else {
				m.formExecType = "local"
			}
		}
	case 5: // TasksFrom (selection)
		if len(m.formTasksetNames) == 0 {
			break
		}
		if msg.Type == tea.KeyLeft {
			// Cycle backwards through taskset names
			currentIdx := -1
			for i, name := range m.formTasksetNames {
				if name == m.formTasksFrom {
					currentIdx = i
					break
				}
			}
			if currentIdx == -1 {
				// Current value not in list, start from beginning
				m.formTasksFrom = m.formTasksetNames[0]
			} else if currentIdx > 0 {
				m.formTasksFrom = m.formTasksetNames[currentIdx-1]
			} else {
				m.formTasksFrom = m.formTasksetNames[len(m.formTasksetNames)-1]
			}
		} else if msg.Type == tea.KeyRight || msg.Type == tea.KeySpace {
			// Cycle forwards through taskset names
			currentIdx := -1
			for i, name := range m.formTasksetNames {
				if name == m.formTasksFrom {
					currentIdx = i
					break
				}
			}
			if currentIdx == -1 {
				// Current value not in list, start from beginning
				m.formTasksFrom = m.formTasksetNames[0]
			} else if currentIdx < len(m.formTasksetNames)-1 {
				m.formTasksFrom = m.formTasksetNames[currentIdx+1]
			} else {
				m.formTasksFrom = m.formTasksetNames[0]
			}
		}
	default:
		// For text input fields
		if inputIdx := m.getInputIndex(m.formFocus); inputIdx >= 0 {
			var cmd tea.Cmd
			m.formInputs[inputIdx], cmd = m.formInputs[inputIdx].Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) saveProject() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.formInputs[0].Value())
	path := strings.TrimSpace(m.formInputs[1].Value())
	container := strings.TrimSpace(m.formInputs[2].Value())
	tasksStr := strings.TrimSpace(m.formInputs[3].Value())
	tagsStr := strings.TrimSpace(m.formInputs[4].Value())

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

	var tags []string
	for _, t := range strings.Split(tagsStr, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}

	project := config.Project{
		Name:      name,
		Path:      path,
		Pinned:    m.formPinned,
		Tasks:     tasks,
		TasksFrom: m.formTasksFrom,
		Tags:      tags,
	}

	// Set execution if not local or if container is specified
	if m.formExecType == "docker" || container != "" {
		project.Execution = &config.Execution{
			Type:      m.formExecType,
			Container: container,
		}
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
	case stateTasksetList:
		mainContent = m.viewTasksetList()
	case stateTasksetAdd:
		mainContent = m.viewTasksetForm("Add New Taskset")
	case stateTasksetEdit:
		mainContent = m.viewTasksetForm("Edit Taskset")
	case stateTasksetDelete:
		mainContent = m.viewTasksetDelete()
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

// Taskset management functions

func (m *Model) initTasksetFormInputs(tasksetName string) {
	// Taskset form has 2 fields: Name and Tasks
	m.formInputs = make([]textinput.Model, 2)

	inputs := []struct {
		placeholder string
		charLimit   int
		value       string
	}{
		{"taskset-name", 50, ""},
		{"lazygit, nvim, yarn dev", 500, ""},
	}

	if tasksetName != "" {
		inputs[0].value = tasksetName
		if tasks, ok := m.config.TaskSets[tasksetName]; ok {
			inputs[1].value = strings.Join(tasks, ", ")
		}
		m.editTasksetName = tasksetName
	} else {
		m.editTasksetName = ""
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

func (m Model) getTasksetNames() []string {
	names := make([]string, 0, len(m.config.TaskSets))
	for name := range m.config.TaskSets {
		names = append(names, name)
	}
	return names
}

func (m Model) updateTasksetList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	tasksetNames := m.getTasksetNames()

	switch {
	case key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.Quit):
		m.state = stateList
		m.message = ""
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.tasksetCursor > 0 {
			m.tasksetCursor--
		}
		m.message = ""

	case key.Matches(msg, m.keys.Down):
		if m.tasksetCursor < len(tasksetNames)-1 {
			m.tasksetCursor++
		}
		m.message = ""

	case key.Matches(msg, m.keys.Add):
		m.state = stateTasksetAdd
		m.initTasksetFormInputs("")
		m.message = ""
		return m, textinput.Blink

	case key.Matches(msg, m.keys.Edit):
		if len(tasksetNames) > 0 && m.tasksetCursor < len(tasksetNames) {
			m.state = stateTasksetEdit
			m.initTasksetFormInputs(tasksetNames[m.tasksetCursor])
			m.message = ""
			return m, textinput.Blink
		}

	case key.Matches(msg, m.keys.Delete):
		if len(tasksetNames) > 0 && m.tasksetCursor < len(tasksetNames) {
			m.state = stateTasksetDelete
			m.message = ""
		}
	}

	return m, nil
}

func (m Model) updateTasksetForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.state = stateTasksetList
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
		return m.saveTaskset()
	}

	var cmd tea.Cmd
	m.formInputs[m.formFocus], cmd = m.formInputs[m.formFocus].Update(msg)
	return m, cmd
}

func (m Model) saveTaskset() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.formInputs[0].Value())
	tasksStr := strings.TrimSpace(m.formInputs[1].Value())

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

	if m.config.TaskSets == nil {
		m.config.TaskSets = make(map[string][]string)
	}

	// If editing and name changed, delete old entry
	if m.state == stateTasksetEdit && m.editTasksetName != "" && m.editTasksetName != name {
		delete(m.config.TaskSets, m.editTasksetName)
	}

	m.config.TaskSets[name] = tasks

	if err := config.Save(m.config); err != nil {
		m.message = "Error saving: " + err.Error()
		m.isError = true
		return m, nil
	}

	m.message = "Taskset saved!"
	m.isError = false
	m.state = stateTasksetList
	return m, nil
}

func (m Model) updateTasksetDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	tasksetNames := m.getTasksetNames()

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("y", "enter"))):
		if m.tasksetCursor < len(tasksetNames) {
			name := tasksetNames[m.tasksetCursor]
			delete(m.config.TaskSets, name)

			if m.tasksetCursor >= len(m.config.TaskSets) && m.tasksetCursor > 0 {
				m.tasksetCursor--
			}

			if err := config.Save(m.config); err != nil {
				m.message = "Error saving: " + err.Error()
				m.isError = true
			} else {
				m.message = "Taskset deleted"
				m.isError = false
			}
		}
		m.state = stateTasksetList

	case key.Matches(msg, key.NewBinding(key.WithKeys("n", "esc", "q"))):
		m.state = stateTasksetList
	}

	return m, nil
}
