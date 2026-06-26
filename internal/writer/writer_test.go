package writer

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

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
