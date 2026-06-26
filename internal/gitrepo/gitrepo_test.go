package gitrepo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPull(t *testing.T) {
	ctx := context.Background()
	runGit := func(dir string, args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}

	bare := t.TempDir()
	runGit(bare, "init", "--bare", "--initial-branch=main")

	a := t.TempDir()
	runGit(a, "clone", bare, ".")
	runGit(a, "config", "user.email", "a@a")
	runGit(a, "config", "user.name", "a")
	if err := os.WriteFile(filepath.Join(a, "f.txt"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(a, "add", "-A")
	runGit(a, "commit", "-m", "c1")
	runGit(a, "push", "origin", "main")

	b := t.TempDir()
	runGit(b, "clone", bare, ".")

	if err := os.WriteFile(filepath.Join(a, "f.txt"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(a, "commit", "-am", "c2")
	runGit(a, "push")

	if err := Pull(ctx, b); err != nil {
		t.Fatalf("Pull: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(b, "f.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "v2" {
		t.Fatalf("after pull f.txt = %q, want v2", data)
	}
}

func TestCommit(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	for _, args := range [][]string{{"init"}, {"config", "user.email", "t@t"}, {"config", "user.name", "t"}} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Commit(ctx, dir, "first"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	out, _ := exec.Command("git", "-C", dir, "log", "--oneline").CombinedOutput()
	if len(out) == 0 {
		t.Fatal("no commit recorded")
	}
	if err := Commit(ctx, dir, "noop"); err != nil {
		t.Fatalf("Commit with no changes should be nil, got %v", err)
	}
}
