package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkipsMalformed(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "facts"), 0o755); err != nil {
		t.Fatal(err)
	}
	good := "---\nid: good\ntype: fact\ndescription: ok\n---\nbody"
	bad := "---\nid: bad\ndescription: 'unterminated\n---\nbody"
	if err := os.WriteFile(filepath.Join(dir, "facts/good.md"), []byte(good), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "facts/bad.md"), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	notes, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(notes) != 1 || notes[0].ID != "good" {
		t.Fatalf("got %+v, want only good", notes)
	}
}

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
