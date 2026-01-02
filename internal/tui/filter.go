package tui

import (
	"sort"
	"strings"

	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/db"
	"github.com/crixuamg/charon/internal/execution"
	"github.com/sahilm/fuzzy"
)

type projectWithIndex struct {
	project config.Project
	index   int
}

func filterProjects(projects []config.Project, cfg *config.Config, filter filterMode, searchQuery string) []projectWithIndex {
	var filtered []projectWithIndex

	for i, p := range projects {
		if !matchesFilter(p, cfg, filter) {
			continue
		}
		filtered = append(filtered, projectWithIndex{project: p, index: i})
	}

	if searchQuery == "" {
		return filtered
	}

	return fuzzyFilterProjects(filtered, searchQuery)
}

func matchesFilter(p config.Project, cfg *config.Config, filter filterMode) bool {
	exec := execution.Resolve(cfg, p)

	switch filter {
	case filterDocker:
		return exec.Type == "docker"
	case filterLocal:
		return exec.Type == "local"
	default:
		return true
	}
}

func fuzzyFilterProjects(projects []projectWithIndex, query string) []projectWithIndex {
	if query == "" {
		return projects
	}

	names := make([]string, len(projects))
	for i, p := range projects {
		names[i] = p.project.Name
	}

	matches := fuzzy.Find(query, names)
	if len(matches) == 0 {
		return nil
	}

	result := make([]projectWithIndex, len(matches))
	for i, match := range matches {
		result[i] = projects[match.Index]
	}

	return result
}

func sortProjects(projects []projectWithIndex, mode sortMode, database *db.DB) {
	switch mode {
	case sortByName:
		sort.Slice(projects, func(i, j int) bool {
			return strings.ToLower(projects[i].project.Name) < strings.ToLower(projects[j].project.Name)
		})
	case sortByRecent:
		if database != nil {
			sort.Slice(projects, func(i, j int) bool {
				timeI, _ := database.GetProjectAccessTime(projects[i].project.Name)
				timeJ, _ := database.GetProjectAccessTime(projects[j].project.Name)

				if timeI == nil && timeJ == nil {
					return false
				}
				if timeI == nil {
					return false
				}
				if timeJ == nil {
					return true
				}
				return timeI.After(*timeJ)
			})
		}
	case sortByCustom:
		// Keep original order from config
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].project.Pinned && !projects[j].project.Pinned
	})
}

func getOriginalIndex(projects []projectWithIndex, cursor int) int {
	if cursor < 0 || cursor >= len(projects) {
		return 0
	}
	return projects[cursor].index
}
