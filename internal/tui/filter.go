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

func filterProjects(projects []config.Project, cfg *config.Config, filter filterMode, activeTag string, showArchived bool, searchQuery string) []projectWithIndex {
	var filtered []projectWithIndex

	for i, p := range projects {
		if p.Archived && !showArchived {
			continue
		}
		if !matchesFilter(p, cfg, filter) {
			continue
		}
		if activeTag != "" && !hasTag(p, activeTag) {
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

func hasTag(p config.Project, tag string) bool {
	for _, t := range p.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AllTags returns a sorted, deduplicated list of all tags across all projects.
func allTags(projects []config.Project) []string {
	seen := make(map[string]struct{})
	var tags []string
	for _, p := range projects {
		for _, t := range p.Tags {
			if _, ok := seen[t]; !ok {
				seen[t] = struct{}{}
				tags = append(tags, t)
			}
		}
	}
	return tags
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
	// Cache DB data upfront to avoid O(n log n) queries during sort.
	var accessTimes map[string]*time.Time
	var accessCounts map[string]int

	if database != nil {
		switch mode {
		case sortByRecent:
			accessTimes = make(map[string]*time.Time, len(projects))
			for _, p := range projects {
				t, _ := database.GetProjectAccessTime(p.project.Name)
				accessTimes[p.project.Name] = t
			}
		case sortByFrequent:
			accessCounts = make(map[string]int, len(projects))
			for _, p := range projects {
				c, _ := database.GetProjectAccessCount(p.project.Name)
				accessCounts[p.project.Name] = c
			}
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

		case sortByFrequent:
			return accessCounts[pi.Name] > accessCounts[pj.Name]

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

// groupedRow is a single rendered line in the grouped layout — either a tag
// header or a project entry.
type groupedRow struct {
	isHeader bool
	tag      string // group name (header rows and project rows)
	count    int    // number of projects in the group (header only)
	pw       projectWithIndex
}

// buildGroupedRows converts a flat filtered list into grouped rows with headers.
// Tags are sorted alphabetically; "Untagged" is always last.
func buildGroupedRows(filtered []projectWithIndex) []groupedRow {
	var tagOrder []string
	groups := make(map[string][]projectWithIndex)

	for _, pw := range filtered {
		tag := "Untagged"
		if len(pw.project.Tags) > 0 {
			tag = pw.project.Tags[0]
		}
		if _, exists := groups[tag]; !exists {
			tagOrder = append(tagOrder, tag)
		}
		groups[tag] = append(groups[tag], pw)
	}

	sort.Slice(tagOrder, func(i, j int) bool {
		if tagOrder[i] == "Untagged" {
			return false
		}
		if tagOrder[j] == "Untagged" {
			return true
		}
		return strings.ToLower(tagOrder[i]) < strings.ToLower(tagOrder[j])
	})

	var rows []groupedRow
	for gi, tag := range tagOrder {
		// Blank separator between groups (not before the first).
		if gi > 0 {
			rows = append(rows, groupedRow{isHeader: false, tag: ""}) // blank spacer
		}
		rows = append(rows, groupedRow{isHeader: true, tag: tag, count: len(groups[tag])})
		for _, pw := range groups[tag] {
			rows = append(rows, groupedRow{pw: pw})
		}
	}
	return rows
}

// findGroupedCursorRow returns the row index (in the grouped rows slice) that
// corresponds to the project with the given original config index.
func findGroupedCursorRow(originalIdx int, rows []groupedRow) int {
	for i, row := range rows {
		if !row.isHeader && row.tag == "" && row.pw.index == 0 {
			// spacer row — skip
			continue
		}
		if !row.isHeader && row.pw.index == originalIdx {
			return i
		}
	}
	return 0
}
