package writer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/antonioisaacdias/llm-wiki/internal/gitrepo"
	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

type Writer struct {
	dir     string
	push    bool
	reindex func(context.Context) error
	mu      sync.Mutex
}

func New(dir string, push bool, reindex func(context.Context) error) *Writer {
	return &Writer{dir: dir, push: push, reindex: reindex}
}

func (w *Writer) Upsert(ctx context.Context, n note.Note) error {
	if n.ID == "" || n.Type == "" {
		return fmt.Errorf("writer: id and type are required")
	}
	if n.Status == "" {
		n.Status = "active"
	}
	data, err := note.Serialize(n)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	path := filepath.Join(w.dir, "facts", n.ID+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("writer: mkdir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writer: write %s: %w", n.ID, err)
	}
	src := n.Source
	if src == "" {
		src = "unknown"
	}
	if err := gitrepo.Commit(ctx, w.dir, fmt.Sprintf("wiki: upsert %s (by %s)", n.ID, src)); err != nil {
		return err
	}
	if w.push {
		if err := gitrepo.Push(ctx, w.dir); err != nil {
			return err
		}
	}
	return w.reindex(ctx)
}
