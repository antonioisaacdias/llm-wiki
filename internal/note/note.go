package note

import (
	"bytes"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Note struct {
	ID           string   `yaml:"id" json:"id"`
	Type         string   `yaml:"type" json:"type"`
	Description  string   `yaml:"description" json:"description"`
	Tags         []string `yaml:"tags" json:"tags"`
	Status       string   `yaml:"status" json:"status"`
	SupersededBy string   `yaml:"superseded_by" json:"superseded_by"`
	Source       string   `yaml:"source" json:"source"`
	Created      string   `yaml:"created" json:"created"`
	Modified     string   `yaml:"modified" json:"modified"`
	Body         string   `yaml:"-" json:"body"`
}

type Stub struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Score       float64 `json:"score"`
}

var ErrNoFrontmatter = errors.New("note: missing frontmatter")

var sep = []byte("---")

func Parse(raw []byte) (Note, error) {
	if !bytes.HasPrefix(raw, sep) {
		return Note{}, ErrNoFrontmatter
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

func Serialize(n Note) ([]byte, error) {
	front, err := yaml.Marshal(n)
	if err != nil {
		return nil, fmt.Errorf("note: marshal: %w", err)
	}
	var b bytes.Buffer
	b.WriteString("---\n")
	b.Write(front)
	b.WriteString("---\n")
	b.WriteString(n.Body)
	return b.Bytes(), nil
}
