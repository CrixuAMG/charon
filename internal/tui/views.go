package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

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
	if m.showArchived {
		statusParts = append(statusParts, m.styles.FilterBadge.Render("Archived: visible"))
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

	if len(m.selected) > 0 {
		statusParts = append(statusParts, m.styles.Error.Render(fmt.Sprintf("Selected: %d  ctrl+d delete  ctrl+r archive", len(m.selected))))
	}

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
		case layoutDetail:
			for i, pw := range visible {
				isSelected := (offset + i) == m.cursor
				isMultiSelected := m.selected[pw.index]
				b.WriteString(m.renderProjectDetail(pw.project, isSelected, isMultiSelected))
				if i < len(visible)-1 {
					b.WriteString("\n")
				}
			}
		default:
			for i, pw := range visible {
				isSelected := (offset + i) == m.cursor
				isMultiSelected := m.selected[pw.index]
				b.WriteString(m.renderProjectCard(pw.project, isSelected, isMultiSelected))
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

func (m Model) renderProjectCard(project config.Project, selected bool, multiSelected bool) string {
	var content strings.Builder

	nameStyle := m.styles.ProjectName
	if selected {
		nameStyle = m.styles.SelectedProjectName
	}

	exec := execution.Resolve(m.config, project)

	checkbox := ""
	if len(m.selected) > 0 {
		if multiSelected {
			checkbox = "[x] "
		} else {
			checkbox = "[ ] "
		}
	}

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

	archivePrefix := ""
	if project.Archived {
		archivePrefix = "🗄 "
	}

	name := nameStyle.Render(checkbox + icon + warningIcon + pin + archivePrefix + project.Name)
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
				m.helpItem("space", "select"),
				m.helpItem("/", "search"),
				m.helpItem("s", "sort"),
				m.helpItem("f", "filter"),
				m.helpItem("T", "tag"),
				m.helpItem("l", "layout"),
				m.helpItem("ctrl+o", "reveal"),
				m.helpItem("a", "add"),
				m.helpItem("e", "edit"),
				m.helpItem("d", "delete"),
				m.helpItem("A", "archive"),
				m.helpItem("ctrl+a", "show archived"),
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
	case stateBulkConfirm:
		items = []string{
			m.helpItem("y", "confirm"),
			m.helpItem("n", "cancel"),
		}
	case stateNote:
		items = []string{
			m.helpItem("ctrl+n", "edit"),
			m.helpItem("esc", "back"),
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
		isMultiSelected := m.selected[pw.index]
		b.WriteString(m.renderProjectTableRow(pw.project, isSelected, isCompact, isMultiSelected))
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderProjectTableRow(project config.Project, selected bool, compact bool, multiSelected bool) string {
	var row strings.Builder

	exec := execution.Resolve(m.config, project)

	// Icons and name
	checkbox := ""
	if len(m.selected) > 0 {
		if multiSelected {
			checkbox = "[x] "
		} else {
			checkbox = "[ ] "
		}
	}
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
	nameText := checkbox + icon + " " + warningIcon + pin + project.Name
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

func (m Model) viewInput() string {
	var b strings.Builder

	b.WriteString(m.styles.ProjectName.Render("Parameters required"))
	if m.pendingProject != nil {
		b.WriteString(m.styles.Subtext.Render(" — " + m.pendingProject.Name))
	}
	b.WriteString("\n\n")

	for i, label := range m.inputLabels {
		lbl := m.styles.FormLabel.Render(label + ":")
		b.WriteString(lbl)
		if i == m.inputFocus {
			b.WriteString(m.styles.SelectedProjectName.Render(m.formInputs[i].View()))
		} else {
			b.WriteString(m.formInputs[i].View())
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Subtext.Render("tab next  ctrl+s open  esc cancel"))

	return m.styles.Form.Render(b.String())
}

func (m Model) viewNote() string {
	var b strings.Builder

	filtered := m.getFilteredProjects()
	if m.cursor >= len(filtered) {
		return ""
	}
	project := filtered[m.cursor].project

	b.WriteString(m.styles.ProjectName.Render("Note — " + project.Name))
	b.WriteString("\n\n")

	if project.Note == "" {
		b.WriteString(m.styles.Subtext.Render("(no note)"))
	} else {
		b.WriteString(project.Note)
	}

	b.WriteString("\n\n")
	b.WriteString(m.styles.Subtext.Render("ctrl+n edit  esc back"))

	return m.styles.Form.Render(b.String())
}

func (m Model) viewBulkConfirm() string {
	var b strings.Builder
	n := len(m.selected)
	action := "delete"
	if m.pendingBulkOp == bulkArchive {
		action = "archive/unarchive"
	}
	b.WriteString(m.styles.ProjectName.Render(fmt.Sprintf("%s %d project(s)?", strings.Title(action), n)))
	b.WriteString("\n\n")
	b.WriteString(m.styles.Subtext.Render("y / enter confirm  n / esc cancel"))
	return m.styles.Form.Render(b.String())
}

func (m Model) renderProjectDetail(project config.Project, selected bool, multiSelected bool) string {
	var content strings.Builder

	exec := execution.Resolve(m.config, project)

	nameStyle := m.styles.ProjectName
	borderStyle := m.styles.Project
	if selected {
		nameStyle = m.styles.SelectedProjectName
		borderStyle = m.styles.SelectedProject
	}

	// ── Line 1: icon + name + badges ────────────────────────────────────────
	icon := "○ "
	if exec.Type == "docker" {
		icon = "◆ "
	}
	prefix := ""
	if len(m.selected) > 0 {
		if multiSelected {
			prefix += "[x] "
		} else {
			prefix += "[ ] "
		}
	}
	if !project.Exists {
		prefix += "⚠ "
	}
	if project.Pinned {
		prefix += "📌 "
	}
	if project.Archived {
		prefix += "🗄 "
	}

	line1 := nameStyle.Render(icon+prefix+project.Name) + "  "
	if exec.Type == "docker" {
		line1 += m.styles.DockerBadge.Render("docker")
	} else {
		line1 += m.styles.LocalBadge.Render("local")
	}
	content.WriteString(line1)
	content.WriteString("\n")

	// ── Line 2: path ─────────────────────────────────────────────────────────
	content.WriteString("  " + m.styles.Path.Render(project.Path))
	content.WriteString("\n")

	// ── Line 3: git info ─────────────────────────────────────────────────────
	if gi, ok := m.gitInfos[project.Name]; ok && gi.IsRepo {
		gitPart := ""
		if gi.HasChanges {
			gitPart = m.styles.GitDirty.Render("● uncommitted changes")
		} else {
			gitPart = m.styles.GitClean.Render("✓ clean")
		}
		branchPart := ""
		if gi.Branch != "" {
			branchPart = "  " + m.styles.Subtext.Render("branch: "+gi.Branch)
		}
		content.WriteString("  " + gitPart + branchPart)
		content.WriteString("\n")
	}

	// ── Line 4: tasks ────────────────────────────────────────────────────────
	taskList := tasks.EffectiveTasks(project, m.config)
	if len(taskList) > 0 {
		taskParts := make([]string, len(taskList))
		for i, t := range taskList {
			if len(t) > 30 {
				t = t[:27] + "..."
			}
			taskParts[i] = t
		}
		content.WriteString("  " + m.styles.Task.Render("tasks: "+strings.Join(taskParts, " • ")))
		content.WriteString("\n")
	}

	// ── Line 5: hooks ────────────────────────────────────────────────────────
	if project.Hooks != nil {
		var hookParts []string
		if project.Hooks.Before != "" {
			hookParts = append(hookParts, "before: "+project.Hooks.Before)
		}
		if project.Hooks.After != "" {
			hookParts = append(hookParts, "after: "+project.Hooks.After)
		}
		if len(hookParts) > 0 {
			content.WriteString("  " + m.styles.Subtext.Render("hooks: "+strings.Join(hookParts, "  ")))
			content.WriteString("\n")
		}
	}

	// ── Line 5: env vars ─────────────────────────────────────────────────────
	if len(project.Env) > 0 {
		keys := make([]string, 0, len(project.Env))
		for k := range project.Env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		envParts := make([]string, len(keys))
		for i, k := range keys {
			envParts[i] = k + "=" + project.Env[k]
		}
		content.WriteString("  " + m.styles.Subtext.Render("env: "+strings.Join(envParts, "  ")))
		content.WriteString("\n")
	}

	// ── Line 6: note (first line) ────────────────────────────────────────────
	if project.Note != "" {
		firstLine := project.Note
		if idx := strings.Index(firstLine, "\n"); idx >= 0 {
			firstLine = firstLine[:idx]
		}
		if len(firstLine) > 60 {
			firstLine = firstLine[:57] + "..."
		}
		content.WriteString("  " + m.styles.Subtext.Render("note: "+firstLine))
		content.WriteString("\n")
	}

	// ── Line 7: tags ─────────────────────────────────────────────────────────
	if len(project.Tags) > 0 {
		var tagParts []string
		for _, tag := range project.Tags {
			tagParts = append(tagParts, m.styles.TagBadge.Render("#"+tag))
		}
		content.WriteString("  " + strings.Join(tagParts, " "))
		content.WriteString("\n")
	}

	// ── Line 6: access stats ──────────────────────────────────────────────────
	if m.db != nil {
		count, _ := m.db.GetProjectAccessCount(project.Name)
		accessedAt, _ := m.db.GetProjectAccessTime(project.Name)
		if count > 0 || accessedAt != nil {
			statParts := []string{}
			if count > 0 {
				statParts = append(statParts, fmt.Sprintf("opened %d×", count))
			}
			if accessedAt != nil {
				statParts = append(statParts, "last: "+humanizeTime(*accessedAt))
			}
			content.WriteString("  " + m.styles.Subtext.Render(strings.Join(statParts, "  ")))
			content.WriteString("\n")
		}
	}

	return borderStyle.Render(content.String())
}

func humanizeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}
