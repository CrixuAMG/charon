package tui

import (
	"fmt"
	"strings"

	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/execution"
	"github.com/crixuamg/charon/internal/tasks"
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
	if m.activeTag != "" {
		statusParts = append(statusParts, m.styles.FilterBadge.Render("Tag: "+m.activeTag))
	}
	if m.currentSort != sortByCustom {
		statusParts = append(statusParts, m.styles.FilterBadge.Render("Sort: "+m.currentSort.String()))
	}
	statusParts = append(statusParts, m.styles.FilterBadge.Render("Layout: "+m.currentLayout.String()))

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
		viewSize := m.viewportSize()
		offset := m.scrollOffset
		end := offset + viewSize
		if end > len(filtered) {
			end = len(filtered)
		}
		visible := filtered[offset:end]

		if offset > 0 {
			b.WriteString(m.styles.Subtext.Render(fmt.Sprintf("  ↑ %d more", offset)))
			b.WriteString("\n")
		}

		switch m.currentLayout {
		case layoutTable, layoutTableCompact:
			b.WriteString(m.renderProjectsTable(visible))
		default:
			for i, pw := range visible {
				isSelected := (offset + i) == m.cursor
				b.WriteString(m.renderProjectCard(pw.project, isSelected))
				if i < len(visible)-1 {
					b.WriteString("\n")
				}
			}
		}
		b.WriteString("\n")

		remaining := len(filtered) - end
		if remaining > 0 {
			b.WriteString(m.styles.Subtext.Render(fmt.Sprintf("  ↓ %d more", remaining)))
			b.WriteString("\n")
		}
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

	// Field order: Name(0), Path(1), Pinned(2), ExecType(3), Container(4), TasksFrom(5), Tasks(6), Tags(7)
	fields := []struct {
		index       int
		label       string
		description string
		value       string
		isInput     bool
	}{
		{0, "Name:", "Project identifier", m.formInputs[0].View(), true},
		{1, "Path:", "Local filesystem path", m.formInputs[1].View(), true},
		{2, "Pinned:", "Pin to top of list (space to toggle)", m.renderBooleanField(m.formPinned), false},
		{3, "Execution:", "Execution type (space/arrows to toggle)", m.renderExecutionField(), false},
		{4, "Container:", "Docker container name (only for docker execution)", m.formInputs[2].View(), true},
		{5, "TasksFrom:", "Select taskset (arrows to cycle)", m.renderTasksFromField(), false},
		{6, "Tasks:", "Comma-separated commands", m.formInputs[3].View(), true},
		{7, "Tags:", "Comma-separated tags (e.g. work, infra)", m.formInputs[4].View(), true},
	}

	for _, field := range fields {
		if field.index == 4 && m.formExecType != "docker" {
			continue
		}

		label := m.styles.FormLabel.Render(field.label)
		b.WriteString(label)

		// Highlight focused field
		if field.index == m.formFocus {
			b.WriteString(m.styles.SelectedProjectName.Render(field.value))
		} else {
			b.WriteString(field.value)
		}

		if field.index == m.formFocus {
			b.WriteString(" " + m.styles.Path.Render(field.description))
		}
		b.WriteString("\n")

		// Show taskset tasks when TasksFrom field is focused
		if field.index == 5 && m.formFocus == 5 && m.formTasksFrom != "" {
			if tasks, ok := m.config.TaskSets[m.formTasksFrom]; ok && len(tasks) > 0 {
				taskList := strings.Join(tasks, ", ")
				b.WriteString(m.styles.FormLabel.Render("  └─ Tasks: "))
				b.WriteString(m.styles.Subtext.Render(taskList))
				b.WriteString("\n")
			}
		}
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

func (m Model) renderBooleanField(value bool) string {
	if value {
		return "[✓] Yes"
	}
	return "[ ] No"
}

func (m Model) renderExecutionField() string {
	if m.formExecType == "docker" {
		return "◆ Docker"
	}
	return "○ Local"
}

func (m Model) renderTasksFromField() string {
	if m.formTasksFrom == "" {
		return "(none)"
	}
	return m.formTasksFrom
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

func (m Model) renderProjectCard(project config.Project, selected bool) string {
	var content strings.Builder

	nameStyle := m.styles.ProjectName
	if selected {
		nameStyle = m.styles.SelectedProjectName
	}

	exec := execution.Resolve(m.config, project)

	warningIcon := ""
	icon := "○ "
	pin := ""
	if !project.Exists {
		warningIcon = "⚠ "
	}
	if exec.Type == "docker" {
		icon = "◆ "
	}
	if project.Pinned {
		pin = "📌 "
	}

	name := nameStyle.Render(icon + warningIcon + pin + project.Name)
	content.WriteString(name)
	content.WriteString(" ")

	if gi, ok := m.gitInfos[project.Name]; ok && gi.IsRepo {
		gitIndicator := m.styles.GitClean.Render("git")
		if gi.HasChanges {
			gitIndicator = m.styles.GitDirty.Render("git*")
		}
		content.WriteString(gitIndicator)
		if gi.Branch != "" {
			content.WriteString(m.styles.Subtext.Render(":" + gi.Branch))
		}
		content.WriteString(" ")
	}

	var pathInfo string
	if exec.Type == "docker" {
		pathInfo = m.styles.DockerBadge.Render("docker") + " " + m.styles.Path.Render(project.Path+"/"+project.Name)
	} else {
		pathInfo = m.styles.LocalBadge.Render("local") + " " + m.styles.Path.Render(project.Path)
	}
	content.WriteString(pathInfo)

	if len(project.Tags) > 0 {
		content.WriteString("\n  ")
		for i, tag := range project.Tags {
			if i > 0 {
				content.WriteString(" ")
			}
			content.WriteString(m.styles.TagBadge.Render("#" + tag))
		}
	}

	// Show tasks based on layout mode
	isCompact := m.currentLayout == layoutCardCompact
	tasks := tasks.EffectiveTasks(project, m.config)

	if len(tasks) > 0 && !isCompact {
		content.WriteString("\n  ")
		tasksList := make([]string, len(tasks))
		for i, t := range tasks {
			if len(t) > 25 {
				t = t[:22] + "..."
			}
			tasksList[i] = t
		}
		taskLine := m.styles.Task.Render(strings.Join(tasksList, " • "))
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
				m.helpItem("T", "tag"),
				m.helpItem("l", "layout"),
				m.helpItem("a", "add"),
				m.helpItem("e", "edit"),
				m.helpItem("d", "delete"),
				m.helpItem("t", "tasksets"),
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
	case stateTasksetList:
		items = []string{
			m.helpItem("↑/↓", "navigate"),
			m.helpItem("a", "add"),
			m.helpItem("e", "edit"),
			m.helpItem("d", "delete"),
			m.helpItem("esc", "back"),
		}
	case stateTasksetAdd, stateTasksetEdit:
		items = []string{
			m.helpItem("tab", "next"),
			m.helpItem("shift+tab", "prev"),
			m.helpItem("ctrl+s", "save"),
			m.helpItem("esc", "cancel"),
		}
	case stateTasksetDelete:
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

// Taskset view functions

func (m Model) viewTasksetList() string {
	var b strings.Builder

	b.WriteString(m.styles.ProjectName.Render("📦 Taskset Management"))
	b.WriteString("\n\n")

	tasksetNames := m.getTasksetNames()

	if len(tasksetNames) == 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.Path.Render("No tasksets configured. Press 'a' to add one."))
		b.WriteString("\n\n")
	} else {
		for i, name := range tasksetNames {
			isSelected := i == m.tasksetCursor
			tasks := m.config.TaskSets[name]

			var line strings.Builder
			if isSelected {
				line.WriteString(m.styles.SelectedProjectName.Render("▶ " + name))
			} else {
				line.WriteString(m.styles.ProjectName.Render("  " + name))
			}

			line.WriteString("\n")
			taskList := strings.Join(tasks, ", ")
			if isSelected {
				line.WriteString(m.styles.SelectedProject.Render("  └─ " + taskList))
			} else {
				line.WriteString(m.styles.Project.Render("  └─ " + taskList))
			}

			b.WriteString(line.String())
			b.WriteString("\n")
		}
	}

	if m.message != "" {
		b.WriteString("\n")
		if m.isError {
			b.WriteString(m.styles.Error.Render("✗ " + m.message))
		} else {
			b.WriteString(m.styles.Success.Render("✓ " + m.message))
		}
	}

	return b.String()
}

func (m Model) viewTasksetForm(title string) string {
	var b strings.Builder

	b.WriteString(m.styles.ProjectName.Render(title))
	b.WriteString("\n\n")

	labels := []string{"Name:", "Tasks:"}
	descriptions := []string{
		"Taskset identifier",
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

func (m Model) viewTasksetDelete() string {
	var b strings.Builder

	tasksetNames := m.getTasksetNames()
	if m.tasksetCursor < len(tasksetNames) {
		name := tasksetNames[m.tasksetCursor]
		tasks := m.config.TaskSets[name]

		var msg strings.Builder
		msg.WriteString(m.styles.Error.Render("⚠ Delete Taskset"))
		msg.WriteString("\n\n")
		msg.WriteString("Are you sure you want to delete this taskset?\n\n")
		msg.WriteString(m.styles.ProjectName.Render(name))
		msg.WriteString("\n")
		msg.WriteString(m.styles.Path.Render("Tasks: " + strings.Join(tasks, ", ")))
		msg.WriteString("\n\n")
		msg.WriteString(m.styles.Path.Render("Press 'y' to confirm or 'n' to cancel"))

		b.WriteString(m.styles.DeleteBox.Render(msg.String()))
	}

	return b.String()
}

func (m Model) renderProjectsTable(filtered []projectWithIndex) string {
	var b strings.Builder

	isCompact := m.currentLayout == layoutTableCompact

	// Table header
	headerStyle := m.styles.ProjectName.Bold(true)
	if isCompact {
		b.WriteString(headerStyle.Render("NAME"))
		b.WriteString("  ")
		b.WriteString(headerStyle.Render("TYPE"))
		b.WriteString("  ")
		b.WriteString(headerStyle.Render("PATH"))
		b.WriteString("\n")
		// Add separator line if width is available
		if m.width > 4 {
			b.WriteString(m.styles.Subtext.Render(strings.Repeat("─", m.width-4)))
			b.WriteString("\n")
		}
	} else {
		b.WriteString(headerStyle.Render("NAME"))
		b.WriteString("  ")
		b.WriteString(headerStyle.Render("TYPE"))
		b.WriteString("  ")
		b.WriteString(headerStyle.Render("PATH"))
		b.WriteString("  ")
		b.WriteString(headerStyle.Render("TASKS"))
		b.WriteString("\n")
		// Add separator line if width is available
		if m.width > 4 {
			b.WriteString(m.styles.Subtext.Render(strings.Repeat("─", m.width-4)))
			b.WriteString("\n")
		}
	}

	// Table rows
	for i, pw := range filtered {
		isSelected := i == m.cursor
		b.WriteString(m.renderProjectTableRow(pw.project, isSelected, isCompact))
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderProjectTableRow(project config.Project, selected bool, compact bool) string {
	var row strings.Builder

	exec := execution.Resolve(m.config, project)

	// Icons and name
	warningIcon := ""
	icon := "○"
	pin := ""
	if !project.Exists {
		warningIcon = "⚠"
	}
	if exec.Type == "docker" {
		icon = "◆"
	}
	if project.Pinned {
		pin = "📌"
	}

	nameStyle := m.styles.ProjectName
	if selected {
		nameStyle = m.styles.SelectedProjectName
	}

	// Name column (max 20 chars)
	nameText := icon + " " + warningIcon + pin + project.Name
	if len(nameText) > 20 {
		nameText = nameText[:17] + "..."
	}
	row.WriteString(nameStyle.Render(fmt.Sprintf("%-20s", nameText)))
	row.WriteString("  ")

	// Type column
	var typeText string
	if exec.Type == "docker" {
		typeText = m.styles.DockerBadge.Render("docker")
	} else {
		typeText = m.styles.LocalBadge.Render("local ")
	}
	row.WriteString(typeText)
	row.WriteString("  ")

	// Path column (max 40 chars)
	pathText := project.Path
	if exec.Type == "docker" {
		pathText = project.Path + "/" + project.Name
	}
	if len(pathText) > 40 {
		pathText = "..." + pathText[len(pathText)-37:]
	}
	row.WriteString(m.styles.Path.Render(fmt.Sprintf("%-40s", pathText)))

	// Git column
	if gi, ok := m.gitInfos[project.Name]; ok && gi.IsRepo {
		row.WriteString("  ")
		if gi.HasChanges {
			row.WriteString(m.styles.GitDirty.Render("git*"))
		} else {
			row.WriteString(m.styles.GitClean.Render("git "))
		}
		if gi.Branch != "" {
			row.WriteString(m.styles.Subtext.Render(":" + gi.Branch))
		}
	}

	// Tasks column (only in non-compact mode)
	if !compact {
		row.WriteString("  ")
		taskList := tasks.EffectiveTasks(project, m.config)
		if len(taskList) > 0 {
			tasksText := strings.Join(taskList, ", ")
			if len(tasksText) > 50 {
				tasksText = tasksText[:47] + "..."
			}
			row.WriteString(m.styles.Task.Render(tasksText))
		}
	}

	// Apply selection style
	result := row.String()
	if selected {
		result = m.styles.SelectedProject.Render(result)
	}

	return result
}
