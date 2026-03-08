package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/crixuamg/charon/internal/config"
)

func (m *Model) initTasksetFormInputs(tasksetName string) {
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

	var taskList []string
	if tasksStr != "" {
		for _, t := range strings.Split(tasksStr, ",") {
			if t = strings.TrimSpace(t); t != "" {
				taskList = append(taskList, t)
			}
		}
	}

	if m.config.TaskSets == nil {
		m.config.TaskSets = make(map[string][]string)
	}

	if m.state == stateTasksetEdit && m.editTasksetName != "" && m.editTasksetName != name {
		delete(m.config.TaskSets, m.editTasksetName)
	}

	m.config.TaskSets[name] = taskList

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
