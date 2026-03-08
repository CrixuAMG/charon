package tui

import (
	"sort"
	"strings"
	"time"

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
	switch filter {
	case filterDocker:
		return execution.Resolve(cfg, p).Type == "docker"
	case filterLocal:
		return execution.Resolve(cfg, p).Type == "local"
	default:
		return true
	}
}

func fuzzyFilterProjects(projects []projectWithIndex, query string) []projectWithIndex {
	// Search across name and path so users can find projects by location too.
	targets := make([]string, len(projects))
	for i, p := range projects {
		targets[i] = p.project.Name + " " + p.project.Path
	}

	matches := fuzzy.Find(query, targets)
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
	// Cache access times upfront to avoid O(n log n) DB queries during sort.
	var accessTimes map[string]*time.Time
	if mode == sortByRecent && database != nil {
		accessTimes = make(map[string]*time.Time, len(projects))
		for _, p := range projects {
			t, _ := database.GetProjectAccessTime(p.project.Name)
			accessTimes[p.project.Name] = t
		}
	}

	sort.SliceStable(projects, func(i, j int) bool {
		pi, pj := projects[i].project, projects[j].project

		// Pinned projects always sort before unpinned.
		if pi.Pinned != pj.Pinned {
			return pi.Pinned
		}

		switch mode {
		case sortByName:
			return strings.ToLower(pi.Name) < strings.ToLower(pj.Name)

		case sortByRecent:
			ti := accessTimes[pi.Name]
			tj := accessTimes[pj.Name]
			if ti == nil && tj == nil {
				return false
			}
			if ti == nil {
				return false
			}
			if tj == nil {
				return true
			}
			return ti.After(*tj)

		default: // sortByCustom — preserve original config order
			return projects[i].index < projects[j].index
		}
	})
}

func getOriginalIndex(projects []projectWithIndex, cursor int) int {
	if cursor < 0 || cursor >= len(projects) {
		return 0
	}
	return projects[cursor].index
}
