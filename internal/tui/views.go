package tui

import (
	"fmt"
	"strings"

	"github.com/crixuamg/charon/internal/config"
)

func (m Model) viewList() string {
	var b strings.Builder

	filtered := m.getFilteredProjects()

	statusParts := []string{}
	if m.searchQuery != "" {
		statusParts = append(statusParts, m.styles.SearchInput.Render("🔍 "+m.searchQuery))
	}
	if m.currentFilter != filterNone {
		statusParts = append(statusParts, m.styles.FilterBadge.Render("Filter: "+m.currentFilter.String()))
	}
	if m.currentSort != sortByCustom {
		statusParts = append(statusParts, m.styles.FilterBadge.Render("Sort: "+m.currentSort.String()))
	}

	count := m.styles.Count.Render(fmt.Sprintf("%d", len(filtered)))
	if len(m.config.Projects) != len(filtered) {
		count += m.styles.Subtext.Render(fmt.Sprintf("/%d", len(m.config.Projects)))
	}
	statusParts = append(statusParts, m.styles.StatusBar.Render("Projects: "+count))

	if len(statusParts) > 0 {
		b.WriteString(strings.Join(statusParts, " "))
		b.WriteString("\n\n")
	}

	if len(filtered) == 0 {
		emptyMsg := "No projects configured. Press 'a' to add one."
		if m.searchQuery != "" {
			emptyMsg = "No projects match your search."
		}
		b.WriteString("\n")
		b.WriteString(m.styles.Path.Render(emptyMsg))
		b.WriteString("\n\n")
	} else {
		for i, pw := range filtered {
			isSelected := i == m.cursor
			b.WriteString(m.renderProject(pw.project, isSelected))
			if i < len(filtered)-1 {
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	if m.message != "" {
		b.WriteString("\n")
		if m.isError {
			b.WriteString(m.styles.Error.Render("✗ " + m.message))
		} else {
			b.WriteString(m.styles.Success.Render("✓ " + m.message))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) viewForm(title string) string {
	var b strings.Builder

	b.WriteString(m.styles.ProjectName.Render(title))
	b.WriteString("\n\n")

	labels := []string{"Name:", "Path:", "Docker:", "Tasks:"}
	descriptions := []string{
		"Project identifier",
		"Local filesystem path",
		"Docker container path (empty for local)",
		"Comma-separated commands",
	}

	for i, input := range m.formInputs {
		label := m.styles.FormLabel.Render(labels[i])
		b.WriteString(label + input.View())
		if i == m.formFocus {
			b.WriteString(" " + m.styles.Path.Render(descriptions[i]))
		}
		b.WriteString("\n")
	}

	formContent := m.styles.Form.Render(b.String())

	var footer strings.Builder
	footer.WriteString(formContent)

	if m.message != "" {
		footer.WriteString("\n\n")
		if m.isError {
			footer.WriteString(m.styles.Error.Render("✗ " + m.message))
		} else {
			footer.WriteString(m.styles.Success.Render("✓ " + m.message))
		}
	}

	return footer.String()
}

func (m Model) viewDelete() string {
	var b strings.Builder

	filtered := m.getFilteredProjects()
	if m.cursor < len(filtered) {
		project := filtered[m.cursor].project

		var msg strings.Builder
		msg.WriteString(m.styles.Error.Render("⚠ Delete Project"))
		msg.WriteString("\n\n")
		msg.WriteString("Are you sure you want to delete ")
		msg.WriteString(m.styles.ProjectName.Render("'" + project.Name + "'"))
		msg.WriteString("?\n\n")
		msg.WriteString(m.styles.Subtext.Render("This action cannot be undone."))

		b.WriteString(m.styles.DeleteBox.Render(msg.String()))
	}

	return b.String()
}

func (m Model) renderProject(project config.Project, selected bool) string {
	var content strings.Builder

	nameStyle := m.styles.ProjectName
	if selected {
		nameStyle = m.styles.SelectedProjectName
	}

	icon := "○ "
	if project.DockerPath != "" || m.config.DockerPath != "" {
		icon = "◆ "
	}

	name := nameStyle.Render(icon + project.Name)
	content.WriteString(name)
	content.WriteString(" ")

	var pathInfo string
	if project.DockerPath != "" {
		pathInfo = m.styles.DockerBadge.Render("docker") + " " + m.styles.Path.Render(project.DockerPath+"/"+project.Name)
	} else if m.config.DockerPath != "" {
		pathInfo = m.styles.DockerBadge.Render("docker") + " " + m.styles.Path.Render(m.config.DockerPath+"/"+project.Name)
	} else {
		pathInfo = m.styles.LocalBadge.Render("local") + " " + m.styles.Path.Render(project.Path)
	}
	content.WriteString(pathInfo)

	if len(project.Tasks) > 0 {
		content.WriteString("\n  ")
		tasks := make([]string, len(project.Tasks))
		for i, t := range project.Tasks {
			if len(t) > 25 {
				t = t[:22] + "..."
			}
			tasks[i] = t
		}
		taskLine := m.styles.Task.Render(strings.Join(tasks, " • "))
		content.WriteString(taskLine)
	}

	style := m.styles.Project
	if selected {
		style = m.styles.SelectedProject
	}

	return style.Render(content.String())
}

func (m Model) renderHelpBar() string {
	var items []string

	switch m.state {
	case stateList:
		if m.searchMode {
			items = []string{
				m.helpItem("enter", "confirm"),
				m.helpItem("esc", "cancel"),
			}
		} else {
			items = []string{
				m.helpItem("↑/↓", "navigate"),
				m.helpItem("enter", "open"),
				m.helpItem("/", "search"),
				m.helpItem("s", "sort"),
				m.helpItem("f", "filter"),
				m.helpItem("a", "add"),
				m.helpItem("e", "edit"),
				m.helpItem("d", "delete"),
				m.helpItem("q", "quit"),
			}
		}
	case stateAdd, stateEdit:
		items = []string{
			m.helpItem("tab", "next"),
			m.helpItem("shift+tab", "prev"),
			m.helpItem("ctrl+s", "save"),
			m.helpItem("esc", "cancel"),
		}
	case stateDelete:
		items = []string{
			m.helpItem("y", "confirm"),
			m.helpItem("n", "cancel"),
		}
	}

	return m.styles.Help.Render(strings.Join(items, " "+m.styles.HelpSeparator.String()+" "))
}

func (m Model) helpItem(key, desc string) string {
	return m.styles.HelpKey.Render(key) + " " + m.styles.HelpDesc.Render(desc)
}

