package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/crixuamg/charon/internal/kitty"
	"github.com/crixuamg/charon/internal/tasks"
)

func (m Model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.state = stateList
		m.pendingProject = nil
		return m, nil

	case key.Matches(msg, m.keys.Save):
		return m.openPendingProject()

	case key.Matches(msg, m.keys.NextField):
		m.formInputs[m.inputFocus].Blur()
		m.inputFocus = (m.inputFocus + 1) % len(m.inputLabels)
		m.formInputs[m.inputFocus].Focus()
		return m, textinput.Blink

	case key.Matches(msg, m.keys.PrevField):
		m.formInputs[m.inputFocus].Blur()
		m.inputFocus--
		if m.inputFocus < 0 {
			m.inputFocus = len(m.inputLabels) - 1
		}
		m.formInputs[m.inputFocus].Focus()
		return m, textinput.Blink
	}

	// If Enter is pressed on last field, open the project.
	if msg.Type == tea.KeyEnter && m.inputFocus == len(m.inputLabels)-1 {
		return m.openPendingProject()
	}

	var cmd tea.Cmd
	m.formInputs[m.inputFocus], cmd = m.formInputs[m.inputFocus].Update(msg)
	return m, cmd
}

func (m Model) openPendingProject() (tea.Model, tea.Cmd) {
	if m.pendingProject == nil {
		m.state = stateList
		return m, nil
	}

	for i, label := range m.inputLabels {
		m.inputValues[label] = strings.TrimSpace(m.formInputs[i].Value())
	}

	taskList := tasks.EffectiveTasks(*m.pendingProject, m.config)
	taskList = tasks.ApplyInputs(taskList, m.inputValues)

	// Temporarily patch the project tasks for opening.
	patched := *m.pendingProject
	patched.Tasks = taskList
	patched.TasksFrom = ""

	if err := kitty.OpenProject(patched, m.config); err != nil {
		m.message = "Error: " + err.Error()
		m.isError = true
		m.state = stateList
		m.pendingProject = nil
		return m, nil
	}

	if m.db != nil {
		_ = m.db.RecordAccess(m.pendingProject.Name)
	}

	m.quitting = true
	return m, tea.Quit
}
