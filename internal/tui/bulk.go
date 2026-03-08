package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/crixuamg/charon/internal/config"
)

func (m Model) updateBulkConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("y", "enter"))):
		switch m.pendingBulkOp {
		case bulkDelete:
			// Delete in reverse order to preserve indices.
			indices := make([]int, 0, len(m.selected))
			for idx := range m.selected {
				indices = append(indices, idx)
			}
			// Sort descending.
			for i := 0; i < len(indices); i++ {
				for j := i + 1; j < len(indices); j++ {
					if indices[j] > indices[i] {
						indices[i], indices[j] = indices[j], indices[i]
					}
				}
			}
			for _, idx := range indices {
				m.config.Projects = append(m.config.Projects[:idx], m.config.Projects[idx+1:]...)
			}
			if m.cursor >= len(m.config.Projects) && m.cursor > 0 {
				m.cursor = len(m.config.Projects) - 1
			}
			if err := config.Save(m.config); err != nil {
				m.message = "Error saving: " + err.Error()
				m.isError = true
			} else {
				m.message = "Deleted selected projects"
				m.isError = false
			}

		case bulkArchive:
			// Toggle archived state: if any is not archived, archive all; otherwise unarchive all.
			anyNotArchived := false
			for idx := range m.selected {
				if !m.config.Projects[idx].Archived {
					anyNotArchived = true
					break
				}
			}
			for idx := range m.selected {
				m.config.Projects[idx].Archived = anyNotArchived
			}
			if err := config.Save(m.config); err != nil {
				m.message = "Error saving: " + err.Error()
				m.isError = true
			} else {
				if anyNotArchived {
					m.message = "Archived selected projects"
				} else {
					m.message = "Unarchived selected projects"
				}
				m.isError = false
			}
		}
		m.selected = nil
		m.state = stateList

	case key.Matches(msg, key.NewBinding(key.WithKeys("n", "esc", "q"))):
		m.state = stateList
	}

	return m, nil
}
