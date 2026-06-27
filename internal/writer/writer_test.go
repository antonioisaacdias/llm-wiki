package writer

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

func gitInit(t *testing.T, dir string) {
	t.Helper()
	for _, a := range [][]string{{"init"}, {"config", "user.email", "t@t"}, {"config", "user.name", "t"}} {
		c := exec.Command("git", a...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestUpsertWritesCommitsReindexes(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	var reindexed int
	w := New(dir, false, func(context.Context) error { reindexed++; return nil })
	n := note.Note{ID: "new-note", Type: "fact", Description: "hello", Body: "b"}
	if err := w.Upsert(context.Background(), n); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "facts", "new-note.md")); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if reindexed != 1 {
		t.Fatalf("reindex called %d times, want 1", reindexed)
	}
	out, _ := exec.Command("git", "-C", dir, "log", "--oneline").CombinedOutput()
	if len(out) == 0 {
		t.Fatal("no commit")
	}
}

func TestUpsertRejectsMissingFields(t *testing.T) {
	w := New(t.TempDir(), false, func(context.Context) error { return nil })
	if err := w.Upsert(context.Background(), note.Note{ID: "x"}); err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestUpsertRejectsInvalid(t *testing.T) {
	cases := map[string]note.Note{
		"bad type":      {ID: "x", Type: "wiki"},
		"bad slug":      {ID: "Bad_ID", Type: "fact"},
		"orphan target": {ID: "x", Type: "fact", Status: "active", SupersededBy: "y"},
		"superseded":    {ID: "x", Type: "fact", Status: "superseded"},
	}
	for name, n := range cases {
		t.Run(name, func(t *testing.T) {
			w := New(t.TempDir(), false, func(context.Context) error { return nil })
			if err := w.Upsert(context.Background(), n); err == nil {
				t.Fatalf("expected validation error for %+v", n)
			}
		})
	}
}

func readNote(t *testing.T, dir, id string) note.Note {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "facts", id+".md"))
	if err != nil {
		t.Fatal(err)
	}
	n, err := note.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestUpsertStampsAndPreservesCreated(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	w := New(dir, false, func(context.Context) error { return nil })

	first := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	w.Now = func() time.Time { return first }
	if err := w.Upsert(context.Background(), note.Note{ID: "n", Type: "fact", Body: "v1"}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	got := readNote(t, dir, "n")
	want := first.Format(time.RFC3339)
	if got.Created != want || got.Modified != want {
		t.Fatalf("first stamp: created=%q modified=%q want %q", got.Created, got.Modified, want)
	}

	second := first.Add(48 * time.Hour)
	w.Now = func() time.Time { return second }
	if err := w.Upsert(context.Background(), note.Note{ID: "n", Type: "fact", Body: "v2"}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	got = readNote(t, dir, "n")
	if got.Created != want {
		t.Errorf("created changed on re-upsert: got %q want %q", got.Created, want)
	}
	if got.Modified != second.Format(time.RFC3339) {
		t.Errorf("modified not updated: got %q want %q", got.Modified, second.Format(time.RFC3339))
	}
	if got.Body != "v2" {
		t.Errorf("body not updated: %q", got.Body)
	}
}
