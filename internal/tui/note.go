package tui

import (
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/crixuamg/charon/internal/config"
)

type noteEditDoneMsg struct {
	note       string
	projectIdx int
	err        error
}

func (m Model) updateNote(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.Quit):
		m.state = stateList
	case key.Matches(msg, m.keys.EditNote):
		filtered := m.getFilteredProjects()
		if len(filtered) > 0 {
			return m.openNoteEditor(filtered)
		}
	}
	return m, nil
}

func (m Model) openNoteEditor(filtered []projectWithIndex) (tea.Model, tea.Cmd) {
	idx := getOriginalIndex(filtered, m.cursor)
	project := &m.config.Projects[idx]

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Write note to a temp file, open editor, read it back.
	tmp, err := os.CreateTemp("", "charon-note-*.md")
	if err != nil {
		m.message = "Failed to create temp file: " + err.Error()
		m.isError = true
		return m, nil
	}
	tmpPath := tmp.Name()

	if _, err := tmp.WriteString(project.Note); err != nil {
		tmp.Close()
		m.message = "Failed to write note: " + err.Error()
		m.isError = true
		return m, nil
	}
	tmp.Close()

	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run editor synchronously (tea.ExecProcess handles terminal restore).
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return noteEditDoneMsg{err: err}
		}
		data, readErr := os.ReadFile(tmpPath)
		os.Remove(tmpPath)
		return noteEditDoneMsg{
			note:       strings.TrimRight(string(data), "\n"),
			projectIdx: idx,
			err:        readErr,
		}
	})
}

func handleNoteEditDone(m Model, msg noteEditDoneMsg) Model {
	if msg.err != nil {
		m.message = "Error: " + msg.err.Error()
		m.isError = true
	} else {
		m.config.Projects[msg.projectIdx].Note = msg.note
		if err := config.Save(m.config); err != nil {
			m.message = "Error saving note: " + err.Error()
			m.isError = true
		} else {
			m.message = "Note saved"
			m.isError = false
		}
	}
	m.state = stateList
	return m
}
