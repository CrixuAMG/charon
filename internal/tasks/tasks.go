package tasks

import (
	"regexp"
	"strings"
	"time"

	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/git"
)

// Context holds variables available for task template expansion.
type Context map[string]string

// placeholderRe matches ${label} patterns used for user-supplied parameters.
var placeholderRe = regexp.MustCompile(`\$\{([^}]+)\}`)

func expand(task string, ctx Context) string {
	for key, value := range ctx {
		task = strings.ReplaceAll(task, "$"+key, value)
	}
	return task
}

// Placeholders returns unique ordered placeholder labels found in tasks,
// e.g. ["branch", "env"] for tasks containing ${branch} and ${env}.
func Placeholders(taskList []string) []string {
	seen := make(map[string]struct{})
	var labels []string
	for _, t := range taskList {
		for _, m := range placeholderRe.FindAllStringSubmatch(t, -1) {
			label := m[1]
			if _, ok := seen[label]; !ok {
				seen[label] = struct{}{}
				labels = append(labels, label)
			}
		}
	}
	return labels
}

// ApplyInputs replaces ${label} placeholders in all tasks with the provided values.
func ApplyInputs(taskList []string, inputs map[string]string) []string {
	result := make([]string, len(taskList))
	for i, t := range taskList {
		result[i] = placeholderRe.ReplaceAllStringFunc(t, func(match string) string {
			label := match[2 : len(match)-1] // strip ${ and }
			if v, ok := inputs[label]; ok {
				return v
			}
			return match
		})
	}
	return result
}

// EffectiveTasks returns the final list of tasks for a project, with taskset
// tasks prepended and all built-in template variables expanded.
// ${label} placeholders are left in place for the caller to handle via
// Placeholders() and ApplyInputs().
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
		"BRANCH":  git.Branch(p.Path),
	}

	for i, task := range taskList {
		taskList[i] = expand(task, ctx)
	}

	return taskList
}
