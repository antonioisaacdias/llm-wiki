package index

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/antonioisaacdias/llm-wiki/internal/note"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

func Open(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("index: open: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

const schema = `
DROP TABLE IF EXISTS notes;
CREATE VIRTUAL TABLE notes USING fts5(
	id UNINDEXED, type UNINDEXED, status UNINDEXED,
	superseded_by UNINDEXED, source UNINDEXED,
	created UNINDEXED, modified UNINDEXED, tags_raw UNINDEXED,
	description, tags, body
);`

func (s *Store) Build(ctx context.Context, notes []note.Note) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("index: schema: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("index: begin: %w", err)
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO notes
		(id,type,status,superseded_by,source,created,modified,tags_raw,description,tags,body)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("index: prepare: %w", err)
	}
	defer stmt.Close()
	for _, n := range notes {
		tags := strings.Join(n.Tags, " ")
		if _, err := stmt.ExecContext(ctx, n.ID, n.Type, n.Status, n.SupersededBy,
			n.Source, n.Created, n.Modified, tags, n.Description, tags, n.Body); err != nil {
			return fmt.Errorf("index: insert %s: %w", n.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("index: commit: %w", err)
	}
	return nil
}

func (s *Store) Search(ctx context.Context, query string, limit int) ([]note.Stub, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	const q = `SELECT id, description, type,
		bm25(notes, 0,0,0,0,0,0,0,0, 10.0, 5.0, 1.0) AS score
		FROM notes WHERE notes MATCH ? AND status = 'active'
		ORDER BY score LIMIT ?`
	rows, err := s.db.QueryContext(ctx, q, query, limit)
	if err != nil {
		return nil, fmt.Errorf("index: search: %w", err)
	}
	defer rows.Close()
	var out []note.Stub
	for rows.Next() {
		var st note.Stub
		if err := rows.Scan(&st.ID, &st.Description, &st.Type, &st.Score); err != nil {
			return nil, fmt.Errorf("index: scan: %w", err)
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *Store) Get(ctx context.Context, id string) (note.Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	const q = `SELECT id,type,description,tags_raw,status,superseded_by,source,created,modified,body
		FROM notes WHERE id = ? LIMIT 1`
	var n note.Note
	var tags string
	err := s.db.QueryRowContext(ctx, q, id).Scan(&n.ID, &n.Type, &n.Description,
		&tags, &n.Status, &n.SupersededBy, &n.Source, &n.Created, &n.Modified, &n.Body)
	if errors.Is(err, sql.ErrNoRows) {
		return note.Note{}, fmt.Errorf("index: note %q not found", id)
	}
	if err != nil {
		return note.Note{}, fmt.Errorf("index: get: %w", err)
	}
	if tags != "" {
		n.Tags = strings.Fields(tags)
	}
	return n, nil
}
