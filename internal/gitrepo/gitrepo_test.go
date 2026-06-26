package gitrepo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

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
