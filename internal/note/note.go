package note

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

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

var (
	Types    = []string{"fact", "reference", "decision", "procedure"}
	Statuses = []string{"active", "superseded"}
)

const (
	StatusActive     = "active"
	StatusSuperseded = "superseded"
)

var (
	slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	linkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
)

func Validate(n Note) error {
	if !slugRe.MatchString(n.ID) {
		return fmt.Errorf("note: invalid id %q: must be a slug matching %s", n.ID, slugRe.String())
	}
	if !contains(Types, n.Type) {
		return fmt.Errorf("note: invalid type %q: must be one of %v", n.Type, Types)
	}
	status := n.Status
	if status == "" {
		status = StatusActive
	}
	if !contains(Statuses, status) {
		return fmt.Errorf("note: invalid status %q: must be one of %v", status, Statuses)
	}
	if status == StatusSuperseded && n.SupersededBy == "" {
		return fmt.Errorf("note: status superseded requires superseded_by")
	}
	if status == StatusActive && n.SupersededBy != "" {
		return fmt.Errorf("note: status active must not set superseded_by")
	}
	return nil
}

func Links(body string) []string {
	matches := linkRe.FindAllStringSubmatch(body, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		target := strings.TrimSpace(m[1])
		if i := strings.IndexByte(target, '|'); i >= 0 {
			target = strings.TrimSpace(target[:i])
		}
		if target != "" {
			out = append(out, target)
		}
	}
	return out
}

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

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
