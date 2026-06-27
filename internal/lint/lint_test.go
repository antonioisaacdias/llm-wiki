package lint

import (
	"testing"

	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

func TestBuildBrokenLinks(t *testing.T) {
	notes := []note.Note{
		{ID: "a", Body: "links [[b]] and [[ghost]]"},
		{ID: "b", Body: "links back [[a]]"},
	}
	r := Build(notes)
	if r.Notes != 2 {
		t.Errorf("notes = %d, want 2", r.Notes)
	}
	if r.Edges != 2 {
		t.Errorf("edges = %d, want 2", r.Edges)
	}
	if len(r.BrokenLinks) != 1 {
		t.Fatalf("broken = %v, want 1", r.BrokenLinks)
	}
	if r.BrokenLinks[0].FromID != "a" || r.BrokenLinks[0].Target != "ghost" {
		t.Errorf("broken = %+v", r.BrokenLinks[0])
	}
}

func TestBuildOrphans(t *testing.T) {
	notes := []note.Note{
		{ID: "hub", Body: "points to [[leaf]]"},
		{ID: "leaf", Body: "no links"},
		{ID: "island", Body: "alone"},
	}
	r := Build(notes)

	wantIn := map[string]bool{"hub": true, "island": true}
	if len(r.OrphansIn) != 2 {
		t.Fatalf("orphans_in = %v, want hub,island", r.OrphansIn)
	}
	for _, id := range r.OrphansIn {
		if !wantIn[id] {
			t.Errorf("unexpected orphan_in %q", id)
		}
	}

	wantOut := map[string]bool{"leaf": true, "island": true}
	if len(r.OrphansOut) != 2 {
		t.Fatalf("orphans_out = %v, want leaf,island", r.OrphansOut)
	}
	for _, id := range r.OrphansOut {
		if !wantOut[id] {
			t.Errorf("unexpected orphan_out %q", id)
		}
	}
}

func TestBuildAliasLinkResolves(t *testing.T) {
	notes := []note.Note{
		{ID: "a", Body: "see [[b|the other note]]"},
		{ID: "b", Body: "[[a]]"},
	}
	r := Build(notes)
	if len(r.BrokenLinks) != 0 {
		t.Errorf("alias link should resolve, got broken %v", r.BrokenLinks)
	}
	if r.Edges != 2 {
		t.Errorf("edges = %d, want 2", r.Edges)
	}
}

func TestBuildEmptyJSONSlicesNotNil(t *testing.T) {
	r := Build([]note.Note{{ID: "a", Body: "[[a]]"}})
	if r.BrokenLinks == nil || r.OrphansIn == nil || r.OrphansOut == nil {
		t.Fatal("empty result slices must be non-nil for stable json")
	}
}
