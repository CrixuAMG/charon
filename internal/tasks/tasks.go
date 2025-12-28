package tasks

import (
	"strings"
	"time"

	"github.com/crixuamg/charon/internal/config"
)

func ExpandTask(task string, ctx TaskContext) string {
	for key, value := range ctx {
		task = strings.ReplaceAll(task, "$"+key, value)
	}
	return task
}

func EffectiveTasks(p config.Project, cfg *config.Config) []string {
	var tasks []string

	if p.TasksFrom != "" {
		tasks = append(tasks, cfg.TaskSets[p.TasksFrom]...)
	}

	tasks = append(tasks, p.Tasks...)

	ctx := TaskContext{
		"PROJECT": p.Name,
		"PATH":    p.Path,
		"DATE":    time.Now().Format("2006-01-02"),
		// TODO: add Git branch
	}

	for i, task := range tasks {
		tasks[i] = ExpandTask(task, ctx)
	}

	return tasks
}
