package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/crixuamg/charon/internal/kitty"
	"github.com/crixuamg/charon/internal/tasks"
)

// linesPerItem returns the number of terminal rows each project occupies.
func (m Model) linesPerItem() int {
	switch m.currentLayout {
	case layoutCardCompact:
		return 2
	case layoutTable, layoutTableCompact:
		return 1
	case layoutDetail:
		return 8
	case layoutGrouped:
		return 1
	default: // layoutCard
		return 3
	}
}

// listStartY returns the approximate terminal row where the project list begins.
// Accounts for the header (title + blank), status bar (1-2 lines + blank),
// and container top padding (1).
func (m Model) listStartY() int {
	return 5
}

// doubleClickThreshold is the maximum duration between two clicks to be
// considered a double-click.
const doubleClickThreshold = 400 * time.Millisecond

func (m Model) updateMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionRelease {
		return m, nil
	}

	filtered := m.getFilteredProjects()
	if len(filtered) == 0 {
		return m, nil
	}

	startY := m.listStartY()
	lines := m.linesPerItem()

	// For table layouts, account for the 2-line header.
	tableHeaderOffset := 0
	if m.currentLayout == layoutTable || m.currentLayout == layoutTableCompact {
		tableHeaderOffset = 2
	}

	relY := msg.Y - startY - tableHeaderOffset
	if relY < 0 {
		return m, nil
	}

	itemIdx := m.scrollOffset + relY/lines
	if itemIdx < 0 || itemIdx >= len(filtered) {
		return m, nil
	}

	now := time.Now()
	isDoubleClick := msg.Y == m.lastClickY && now.Sub(m.lastClickTime) < doubleClickThreshold

	m.lastClickY = msg.Y
	m.lastClickTime = now

	if isDoubleClick && m.cursor == itemIdx {
		// Open the project on double-click.
		project := filtered[itemIdx].project
		taskList := tasks.EffectiveTasks(project, m.config)
		labels := tasks.Placeholders(taskList)
		if len(labels) > 0 {
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
	} else {
		// Single click: move cursor.
		m.cursor = itemIdx
		m.scrollOffset = adjustScroll(m.cursor, m.scrollOffset, m.viewportSize())
		m.message = ""
	}

	return m, nil
}
