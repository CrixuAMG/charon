package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/crixuamg/charon/internal/config"
)

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
			currentIdx := -1
			for i, name := range m.formTasksetNames {
				if name == m.formTasksFrom {
					currentIdx = i
					break
				}
			}
			if currentIdx == -1 {
				m.formTasksFrom = m.formTasksetNames[0]
			} else if currentIdx > 0 {
				m.formTasksFrom = m.formTasksetNames[currentIdx-1]
			} else {
				m.formTasksFrom = m.formTasksetNames[len(m.formTasksetNames)-1]
			}
		} else if msg.Type == tea.KeyRight || msg.Type == tea.KeySpace {
			currentIdx := -1
			for i, name := range m.formTasksetNames {
				if name == m.formTasksFrom {
					currentIdx = i
					break
				}
			}
			if currentIdx == -1 {
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

	var taskList []string
	if tasksStr != "" {
		for _, t := range strings.Split(tasksStr, ",") {
			if t = strings.TrimSpace(t); t != "" {
				taskList = append(taskList, t)
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
		Tasks:     taskList,
		TasksFrom: m.formTasksFrom,
		Tags:      tags,
	}

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
