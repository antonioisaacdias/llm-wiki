package lint

import (
	"sort"

	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

type BrokenLink struct {
	FromID string `json:"from_id"`
	Target string `json:"target"`
}

type Report struct {
	Notes       int          `json:"notes"`
	Edges       int          `json:"edges"`
	BrokenLinks []BrokenLink `json:"broken_links"`
	OrphansIn   []string     `json:"orphans_in"`
	OrphansOut  []string     `json:"orphans_out"`
}

func Build(notes []note.Note) Report {
	exists := make(map[string]bool, len(notes))
	for _, n := range notes {
		exists[n.ID] = true
	}

	report := Report{
		Notes:       len(notes),
		BrokenLinks: []BrokenLink{},
		OrphansIn:   []string{},
		OrphansOut:  []string{},
	}
	hasIn := make(map[string]bool, len(notes))
	hasOut := make(map[string]bool, len(notes))

	for _, n := range notes {
		for _, target := range note.Links(n.Body) {
			if !exists[target] {
				report.BrokenLinks = append(report.BrokenLinks, BrokenLink{FromID: n.ID, Target: target})
				continue
			}
			report.Edges++
			hasOut[n.ID] = true
			hasIn[target] = true
		}
	}

	for _, n := range notes {
		if !hasIn[n.ID] {
			report.OrphansIn = append(report.OrphansIn, n.ID)
		}
		if !hasOut[n.ID] {
			report.OrphansOut = append(report.OrphansOut, n.ID)
		}
	}

	sort.Slice(report.BrokenLinks, func(i, j int) bool {
		a, b := report.BrokenLinks[i], report.BrokenLinks[j]
		if a.FromID != b.FromID {
			return a.FromID < b.FromID
		}
		return a.Target < b.Target
	})
	sort.Strings(report.OrphansIn)
	sort.Strings(report.OrphansOut)
	return report
}
