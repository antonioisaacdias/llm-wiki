package index

import (
	"context"
	"testing"

	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	notes := []note.Note{
		{ID: "vram-cliff", Type: "fact", Status: "active", Description: "ollama vram overflow slow", Tags: []string{"ollama"}, Body: "penhasco de vram"},
		{ID: "old-fact", Type: "fact", Status: "superseded", Description: "llama.cpp service", Body: "descontinuado"},
	}
	if err := s.Build(context.Background(), notes); err != nil {
		t.Fatalf("Build: %v", err)
	}
	return s
}

func TestSearchFindsActive(t *testing.T) {
	s := newTestStore(t)
	stubs, err := s.Search(context.Background(), "vram", 8)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(stubs) != 1 || stubs[0].ID != "vram-cliff" {
		t.Fatalf("got %+v, want [vram-cliff]", stubs)
	}
}

func TestSearchExcludesSuperseded(t *testing.T) {
	s := newTestStore(t)
	stubs, err := s.Search(context.Background(), "llama", 8)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(stubs) != 0 {
		t.Fatalf("got %+v, want none (superseded excluded)", stubs)
	}
}

func TestGet(t *testing.T) {
	s := newTestStore(t)
	n, err := s.Get(context.Background(), "vram-cliff")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if n.Body != "penhasco de vram" {
		t.Errorf("Body = %q", n.Body)
	}
}

func TestGetMissing(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Get(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}
