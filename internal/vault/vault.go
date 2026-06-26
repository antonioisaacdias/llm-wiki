package vault

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

func Load(root string) ([]note.Note, error) {
	var notes []note.Note
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		n, err := note.Parse(raw)
		if err != nil {
			return nil
		}
		notes = append(notes, n)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("vault: walk %s: %w", root, err)
	}
	return notes, nil
}
