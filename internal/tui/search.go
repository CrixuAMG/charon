package tui

import tea "github.com/charmbracelet/bubbletea"

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
