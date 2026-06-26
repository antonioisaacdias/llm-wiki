package note

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	raw := []byte(`---
id: vram-overflow-cliff
type: fact
description: Ollama com poucas layers fica 4x mais lento.
tags: [ollama, vram]
status: active
source: claude-code
created: 2026-05-10T00:00:00Z
---
Corpo da nota. Linka [[outra-nota]].
`)
	n, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if n.ID != "vram-overflow-cliff" {
		t.Errorf("ID = %q, want vram-overflow-cliff", n.ID)
	}
	if n.Type != "fact" {
		t.Errorf("Type = %q, want fact", n.Type)
	}
	if n.Status != "active" {
		t.Errorf("Status = %q, want active", n.Status)
	}
	if len(n.Tags) != 2 || n.Tags[0] != "ollama" {
		t.Errorf("Tags = %v, want [ollama vram]", n.Tags)
	}
	if n.Body != "Corpo da nota. Linka [[outra-nota]].\n" {
		t.Errorf("Body = %q", n.Body)
	}
}

func TestParseMissingID(t *testing.T) {
	_, err := Parse([]byte("---\ntype: fact\n---\nx"))
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestParseNoFrontmatter(t *testing.T) {
	_, err := Parse([]byte("just a body, no frontmatter"))
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}

func TestNoteJSONKeysSnakeCase(t *testing.T) {
	n := Note{ID: "x", Description: "d", Body: "b", SupersededBy: "y"}
	out, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, key := range []string{`"id"`, `"description"`, `"body"`, `"superseded_by"`} {
		if !strings.Contains(s, key) {
			t.Errorf("missing %s in %s", key, s)
		}
	}
}

func TestParseNoFrontmatterSentinel(t *testing.T) {
	_, err := Parse([]byte("no frontmatter here"))
	if !errors.Is(err, ErrNoFrontmatter) {
		t.Fatalf("got %v, want ErrNoFrontmatter", err)
	}
}
