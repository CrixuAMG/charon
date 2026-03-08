package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crixuamg/charon/internal/config"
)

func newTestModel(projectNames ...string) Model {
	ps := make([]config.Project, len(projectNames))
	for i, n := range projectNames {
		ps[i] = config.Project{Name: n, Path: "/tmp/" + n, Exists: true}
	}
	cfg := &config.Config{Projects: ps}
	m := NewModel(cfg)
	m.width = 120
	m.height = 40
	return m
}

func pressKey(m Model, key string) Model {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	if key == "up" {
		msg = tea.KeyMsg{Type: tea.KeyUp}
	} else if key == "down" {
		msg = tea.KeyMsg{Type: tea.KeyDown}
	} else if key == "g" {
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
	} else if key == "G" {
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}
	}
	next, _ := m.Update(msg)
	return next.(Model)
}

// ── navigation ────────────────────────────────────────────────────────────────

func TestNavigateDown(t *testing.T) {
	m := newTestModel("a", "b", "c")
	m = pressKey(m, "down")
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}
}

func TestNavigateDoesNotGoAboveZero(t *testing.T) {
	m := newTestModel("a", "b", "c")
	m = pressKey(m, "up")
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
}

func TestNavigateDoesNotGoBelowLast(t *testing.T) {
	m := newTestModel("a", "b")
	m = pressKey(m, "down")
	m = pressKey(m, "down")
	m = pressKey(m, "down") // extra press
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}
}

func TestJumpToFirst(t *testing.T) {
	m := newTestModel("a", "b", "c")
	m = pressKey(m, "down")
	m = pressKey(m, "down")
	// press "g" for first
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
	next, _ := m.Update(msg)
	m = next.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
}

func TestJumpToLast(t *testing.T) {
	m := newTestModel("a", "b", "c")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}
	next, _ := m.Update(msg)
	m = next.(Model)
	if m.cursor != 2 {
		t.Errorf("expected cursor 2, got %d", m.cursor)
	}
}

// ── search ────────────────────────────────────────────────────────────────────

func TestSearchModeActivation(t *testing.T) {
	m := newTestModel("a", "b")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	next, _ := m.Update(msg)
	m = next.(Model)
	if !m.searchMode {
		t.Error("expected searchMode to be true after pressing /")
	}
}

func TestSearchQuery(t *testing.T) {
	m := newTestModel("alpha", "beta")
	// enter search mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	next, _ := m.Update(msg)
	m = next.(Model)
	// type "alp"
	for _, ch := range "alp" {
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		next, _ = m.Update(msg)
		m = next.(Model)
	}
	if m.searchQuery != "alp" {
		t.Errorf("expected searchQuery 'alp', got %q", m.searchQuery)
	}
	filtered := m.getFilteredProjects()
	if len(filtered) != 1 || filtered[0].project.Name != "alpha" {
		t.Errorf("expected only alpha in results, got %v", filtered)
	}
}

func TestSearchEscClears(t *testing.T) {
	m := newTestModel("alpha", "beta")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	next, _ := m.Update(msg)
	m = next.(Model)
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	next, _ = m.Update(msg)
	m = next.(Model)
	if m.searchMode {
		t.Error("expected searchMode false after Esc")
	}
	if m.searchQuery != "" {
		t.Errorf("expected empty query, got %q", m.searchQuery)
	}
}

// ── sort cycling ──────────────────────────────────────────────────────────────

func TestSortCycle(t *testing.T) {
	m := newTestModel("a")
	initial := m.currentSort
	for i := 0; i < 4; i++ {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	if m.currentSort != initial {
		t.Errorf("expected sort to cycle back to %d, got %d", initial, m.currentSort)
	}
}

// ── window size ───────────────────────────────────────────────────────────────

func TestWindowSizeUpdate(t *testing.T) {
	m := newTestModel("a")
	next, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	m = next.(Model)
	if m.width != 200 || m.height != 50 {
		t.Errorf("expected 200×50, got %d×%d", m.width, m.height)
	}
}

// ── viewportSize ──────────────────────────────────────────────────────────────

func TestViewportSizeFallback(t *testing.T) {
	m := newTestModel("a")
	m.height = 0
	if got := m.viewportSize(); got != 20 {
		t.Errorf("expected fallback 20, got %d", got)
	}
}

func TestAdjustScroll(t *testing.T) {
	// cursor above viewport → scroll up
	if got := adjustScroll(0, 5, 3); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
	// cursor below viewport → scroll down
	if got := adjustScroll(8, 0, 3); got != 6 {
		t.Errorf("expected 6, got %d", got)
	}
	// cursor in viewport → no change
	if got := adjustScroll(2, 0, 5); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

// ── search history ────────────────────────────────────────────────────────────

func TestAppendUniqueHistory(t *testing.T) {
	h := appendUniqueHistory(nil, "foo", 5)
	h = appendUniqueHistory(h, "bar", 5)
	h = appendUniqueHistory(h, "bar", 5) // duplicate
	if len(h) != 2 {
		t.Errorf("expected 2 entries, got %d", len(h))
	}
}

func TestHistoryMaxSize(t *testing.T) {
	var h []string
	for i := 0; i < 25; i++ {
		h = appendUniqueHistory(h, string(rune('a'+i)), 20)
	}
	if len(h) != 20 {
		t.Errorf("expected max 20, got %d", len(h))
	}
}
