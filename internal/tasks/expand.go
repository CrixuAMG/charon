package tasks

import "strings"

func Expand(task string, ctx TaskContext) string {
	for key, value := range ctx {
		task = strings.ReplaceAll(task, "$"+key, value)
	}
	return task
}
