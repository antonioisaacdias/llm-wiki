package note

import (
	"bytes"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Note struct {
	ID           string   `yaml:"id"`
	Type         string   `yaml:"type"`
	Description  string   `yaml:"description"`
	Tags         []string `yaml:"tags"`
	Status       string   `yaml:"status"`
	SupersededBy string   `yaml:"superseded_by"`
	Source       string   `yaml:"source"`
	Created      string   `yaml:"created"`
	Modified     string   `yaml:"modified"`
	Body         string   `yaml:"-"`
}

type Stub struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Score       float64 `json:"score"`
}

var sep = []byte("---")

func Parse(raw []byte) (Note, error) {
	if !bytes.HasPrefix(raw, sep) {
		return Note{}, errors.New("note: missing frontmatter")
	}
	rest := raw[len(sep):]
	end := bytes.Index(rest, append([]byte("\n"), sep...))
	if end < 0 {
		return Note{}, errors.New("note: unterminated frontmatter")
	}
	front := rest[:end]
	body := rest[end+len("\n")+len(sep):]
	body = bytes.TrimPrefix(body, []byte("\n"))

	var n Note
	if err := yaml.Unmarshal(front, &n); err != nil {
		return Note{}, fmt.Errorf("note: yaml: %w", err)
	}
	if n.ID == "" {
		return Note{}, errors.New("note: missing required field id")
	}
	if n.Status == "" {
		n.Status = "active"
	}
	n.Body = string(body)
	return n, nil
}
