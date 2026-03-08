package tasks

import (
	"strings"
	"time"

	"github.com/crixuamg/charon/internal/config"
)

// Context holds variables available for task template expansion.
type Context map[string]string

func expand(task string, ctx Context) string {
	for key, value := range ctx {
		task = strings.ReplaceAll(task, "$"+key, value)
	}
	return task
}

// EffectiveTasks returns the final list of tasks for a project, with taskset
// tasks prepended and all template variables expanded.
func EffectiveTasks(p config.Project, cfg *config.Config) []string {
	var taskList []string

	if p.TasksFrom != "" {
		taskList = append(taskList, cfg.TaskSets[p.TasksFrom]...)
	}

	taskList = append(taskList, p.Tasks...)

	ctx := Context{
		"PROJECT": p.Name,
		"PATH":    p.Path,
		"DATE":    time.Now().Format("2006-01-02"),
	}

	for i, task := range taskList {
		taskList[i] = expand(task, ctx)
	}

	return taskList
}
