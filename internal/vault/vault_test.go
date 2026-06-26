package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("facts/a.md", "---\nid: a\ntype: fact\ndescription: first\n---\nbody a")
	write("facts/b.md", "---\nid: b\ntype: reference\ndescription: second\n---\nbody b")
	write("README.md", "not a note, should be ignored if no frontmatter")

	notes, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("got %d notes, want 2", len(notes))
	}
}
