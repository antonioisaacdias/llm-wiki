package writer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/antonioisaacdias/llm-wiki/internal/gitrepo"
	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

type Writer struct {
	dir     string
	push    bool
	reindex func(context.Context) error
	Now     func() time.Time
	mu      sync.Mutex
}

func New(dir string, push bool, reindex func(context.Context) error) *Writer {
	return &Writer{dir: dir, push: push, reindex: reindex, Now: time.Now}
}

func (w *Writer) Upsert(ctx context.Context, n note.Note) error {
	if n.Status == "" {
		n.Status = note.StatusActive
	}
	if err := note.Validate(n); err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	path := filepath.Join(w.dir, "facts", n.ID+".md")
	stamp := w.now().UTC().Format(time.RFC3339)
	n.Modified = stamp
	n.Created = w.createdFor(path, stamp)

	data, err := note.Serialize(n)
	if err != nil {
		return err
	}
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

func (w *Writer) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now()
}

func (w *Writer) createdFor(path, stamp string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return stamp
	}
	existing, err := note.Parse(raw)
	if err != nil || existing.Created == "" {
		return stamp
	}
	return existing.Created
}
