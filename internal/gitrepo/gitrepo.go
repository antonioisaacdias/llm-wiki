package gitrepo

import (
	"context"
	"fmt"
	"os/exec"
)

func git(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v: %w: %s", args, err, out)
	}
	return nil
}

func Commit(ctx context.Context, dir, msg string) error {
	if err := git(ctx, dir, "add", "-A"); err != nil {
		return err
	}
	check := exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	check.Dir = dir
	if check.Run() == nil {
		return nil
	}
	return git(ctx, dir, "commit", "-m", msg)
}

func Push(ctx context.Context, dir string) error {
	return git(ctx, dir, "push")
}

func Pull(ctx context.Context, dir string) error {
	return git(ctx, dir, "pull", "--ff-only")
}
