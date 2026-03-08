package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/crixuamg/charon/internal/browser"
	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/db"
	gitinfo "github.com/crixuamg/charon/internal/git"
	"github.com/crixuamg/charon/internal/kitty"
	"github.com/crixuamg/charon/internal/tasks"
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
	layoutDetail
	layoutGrouped // projects grouped under tag headers
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
	case layoutDetail:
		return "detail"
	case layoutGrouped:
		return "grouped"
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
	showArchived     bool
	// Taskset management fields
	tasksetCursor   int
	editTasksetName string // name of taskset being edited
	// Input prompt fields (task template parameters)
	inputLabels    []string
	inputValues    map[string]string
	inputFocus     int
	pendingProject *config.Project
	// Multi-select / bulk operations
	selected      map[int]bool // keyed by original config.Projects index
	pendingBulkOp bulkAction
	// Mouse support
	lastClickY    int
	lastClickTime time.Time
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
	case "detail":
		return layoutDetail
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

// Fixed line counts that consume vertical space outside the project list.
const (
	viewportHeaderLines  = 2 // title + blank line
	viewportStatusLines  = 2 // status bar + blank line
	viewportHelpLines    = 1
	viewportPaddingLines = 2 // container top + bottom padding
	viewportMessageLines = 2 // reserve for status/error message
	viewportOverhead     = viewportHeaderLines + viewportStatusLines + viewportHelpLines + viewportPaddingLines + viewportMessageLines
)

// viewportSize returns the number of list items (or rows for grouped layout)
// that fit in the terminal.
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
	case layoutDetail:
		// ~8 lines per item
		if available < 8 {
			return 1
		}
		return available / 8
	case layoutGrouped:
		// 1 line per row (headers and projects alike)
		return available
	default: // layoutCard
		// 2 content lines (name+path, tasks) + 1 separator = 3 lines per item
		if available < 3 {
			return 1
		}
		return available / 3
	}
}

// adjustScrollForLayout computes the new scroll offset after the cursor moves.
// For the grouped layout scrollOffset tracks row indices (including header rows);
// for all other layouts it tracks project indices.
func (m Model) adjustScrollForLayout(cursor int, filtered []projectWithIndex) int {
	if m.currentLayout == layoutGrouped && len(filtered) > 0 {
		rows := buildGroupedRows(filtered)
		originalIdx := 0
		if cursor < len(filtered) {
			originalIdx = filtered[cursor].index
		}
		cursorRow := findGroupedCursorRow(originalIdx, rows)
		return adjustScroll(cursorRow, m.scrollOffset, m.viewportSize())
	}
	return adjustScroll(cursor, m.scrollOffset, m.viewportSize())
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
	filtered := filterProjects(m.config.Projects, m.config, m.currentFilter, m.activeTag, m.showArchived, m.searchQuery)
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

	case noteEditDoneMsg:
		return handleNoteEditDone(m, msg), nil

	case tea.MouseMsg:
		if m.state == stateList && !m.searchMode {
			return m.updateMouse(msg)
		}

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
		case stateInput:
			return m.updateInput(msg)
		case stateBulkConfirm:
			return m.updateBulkConfirm(msg)
		case stateNote:
			return m.updateNote(msg)
		case stateDashboard:
			return m.updateDashboard(msg)
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
		m.scrollOffset = m.adjustScrollForLayout(m.cursor, filtered)
		m.message = ""

	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(filtered)-1 {
			m.cursor++
		}
		m.scrollOffset = m.adjustScrollForLayout(m.cursor, filtered)
		m.message = ""

	case key.Matches(msg, m.keys.Enter):
		if len(filtered) > 0 {
			project := filtered[m.cursor].project
			taskList := tasks.EffectiveTasks(project, m.config)
			labels := tasks.Placeholders(taskList)
			if len(labels) > 0 {
				// Task has ${label} placeholders — collect input first.
				m.pendingProject = &project
				m.inputLabels = labels
				m.inputValues = make(map[string]string)
				m.inputFocus = 0
				m.formInputs = make([]textinput.Model, len(labels))
				for i, label := range labels {
					m.formInputs[i] = textinput.New()
					m.formInputs[i].Placeholder = label
					m.formInputs[i].CharLimit = 200
					m.formInputs[i].Width = 40
				}
				m.formInputs[0].Focus()
				m.state = stateInput
				m.message = ""
				return m, textinput.Blink
			}
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
		m.currentLayout = (m.currentLayout + 1) % 6
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
			m.scrollOffset = m.adjustScrollForLayout(m.cursor, filtered)
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

	case key.Matches(msg, m.keys.Archive):
		if len(filtered) > 0 {
			idx := getOriginalIndex(filtered, m.cursor)
			m.config.Projects[idx].Archived = !m.config.Projects[idx].Archived
			archived := m.config.Projects[idx].Archived
			if err := config.Save(m.config); err != nil {
				m.message = "Error saving: " + err.Error()
				m.isError = true
			} else if archived {
				m.message = "Project archived"
				m.isError = false
			} else {
				m.message = "Project unarchived"
				m.isError = false
			}
			// Move cursor if the project just disappeared from view.
			filtered = m.getFilteredProjects()
			if m.cursor >= len(filtered) && m.cursor > 0 {
				m.cursor--
			}
		}

	case key.Matches(msg, m.keys.ToggleArchive):
		m.showArchived = !m.showArchived
		m.cursor = 0
		m.scrollOffset = 0
		m.message = ""

	case key.Matches(msg, m.keys.ThemeCycle):
		themes := []string{"default", "gruvbox", "tokyonight"}
		current := m.config.Theme
		next := themes[0]
		for i, t := range themes {
			if t == current {
				next = themes[(i+1)%len(themes)]
				break
			}
		}
		m.config.Theme = next
		m.styles = NewStyles(getTheme(next))
		if err := config.Save(m.config); err != nil {
			m.message = "Failed to save theme"
			m.isError = true
		} else {
			m.message = "Theme: " + next
			m.isError = false
		}

	case key.Matches(msg, m.keys.Select):
		if len(filtered) > 0 {
			idx := getOriginalIndex(filtered, m.cursor)
			if m.selected == nil {
				m.selected = make(map[int]bool)
			}
			if m.selected[idx] {
				delete(m.selected, idx)
			} else {
				m.selected[idx] = true
			}
			// Advance cursor after toggling.
			if m.cursor < len(filtered)-1 {
				m.cursor++
				m.scrollOffset = m.adjustScrollForLayout(m.cursor, filtered)
			}
			m.message = ""
		}

	case key.Matches(msg, m.keys.BulkDelete):
		if len(m.selected) > 0 {
			m.pendingBulkOp = bulkDelete
			m.state = stateBulkConfirm
			m.message = ""
		}

	case key.Matches(msg, m.keys.BulkArchive):
		if len(m.selected) > 0 {
			m.pendingBulkOp = bulkArchive
			m.state = stateBulkConfirm
			m.message = ""
		}

	case key.Matches(msg, m.keys.Note):
		if len(filtered) > 0 {
			m.state = stateNote
			m.message = ""
		}

	case key.Matches(msg, m.keys.EditNote):
		if len(filtered) > 0 {
			return m.openNoteEditor(filtered)
		}

	case key.Matches(msg, m.keys.Dashboard):
		m.state = stateDashboard
		m.message = ""
	}

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
	case stateInput:
		mainContent = m.viewInput()
	case stateBulkConfirm:
		mainContent = m.viewBulkConfirm()
	case stateNote:
		mainContent = m.viewNote()
	case stateDashboard:
		mainContent = m.viewDashboard()
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
