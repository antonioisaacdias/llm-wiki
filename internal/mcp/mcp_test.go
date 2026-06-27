package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/antonioisaacdias/llm-wiki/internal/httpapi"
	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

type fakeSearcher struct {
	stubs []note.Stub
	note  note.Note
	all   []note.Note
	err   error
}

func (f fakeSearcher) Search(_ context.Context, _ string, _ int) ([]note.Stub, error) {
	return f.stubs, f.err
}

func (f fakeSearcher) Get(_ context.Context, _ string) (note.Note, error) {
	return f.note, f.err
}

func (f fakeSearcher) All(_ context.Context) ([]note.Note, error) {
	return f.all, f.err
}

type fakeWriter struct {
	called bool
	got    note.Note
	err    error
}

func (f *fakeWriter) Upsert(_ context.Context, n note.Note) error {
	f.called = true
	f.got = n
	return f.err
}

func connect(t *testing.T, s fakeSearcher) *mcp.ClientSession {
	return connectWith(t, s, &fakeWriter{})
}

func connectWith(t *testing.T, s fakeSearcher, wr httpapi.Writer) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	server := New(s, wr, s)
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func TestListTools(t *testing.T) {
	session := connect(t, fakeSearcher{})

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}

	got := map[string]bool{}
	for _, tool := range res.Tools {
		got[tool.Name] = true
	}
	for _, want := range []string{"search_wiki", "get_note", "upsert_note", "lint_wiki"} {
		if !got[want] {
			t.Errorf("missing tool %q; got %v", want, got)
		}
	}
	if len(res.Tools) != 4 {
		t.Errorf("want 4 tools, got %d", len(res.Tools))
	}
}

func TestLintWiki(t *testing.T) {
	s := fakeSearcher{all: []note.Note{
		{ID: "a", Body: "see [[b]] and [[ghost]]"},
		{ID: "b", Body: "back to [[a]]"},
	}}
	session := connect(t, s)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "lint_wiki",
		Arguments: lintInput{},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}
	out, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content not an object: %T", res.StructuredContent)
	}
	if out["notes"].(float64) != 2 {
		t.Errorf("notes = %v, want 2", out["notes"])
	}
	broken, ok := out["broken_links"].([]any)
	if !ok || len(broken) != 1 {
		t.Fatalf("broken_links = %v, want 1", out["broken_links"])
	}
}

func TestUpsertNote(t *testing.T) {
	wr := &fakeWriter{}
	session := connectWith(t, fakeSearcher{}, wr)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "upsert_note",
		Arguments: upsertInput{ID: "n1", Type: "fact"},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}

	out, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content not an object: %T", res.StructuredContent)
	}
	if out["ok"] != true {
		t.Errorf("want ok true, got %v", out["ok"])
	}
	if out["id"] != "n1" {
		t.Errorf("want id n1, got %v", out["id"])
	}
	if !wr.called {
		t.Error("writer was not invoked")
	}
	if wr.got.ID != "n1" || wr.got.Type != "fact" {
		t.Errorf("writer got unexpected note: %+v", wr.got)
	}
}

func TestSearchWiki(t *testing.T) {
	s := fakeSearcher{stubs: []note.Stub{
		{ID: "n1", Description: "first", Type: "doc", Score: 1.5},
		{ID: "n2", Description: "second", Type: "doc", Score: 0.9},
	}}
	session := connect(t, s)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "search_wiki",
		Arguments: searchInput{Query: "anything"},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}

	out, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content not an object: %T", res.StructuredContent)
	}
	results, ok := out["results"].([]any)
	if !ok {
		t.Fatalf("results not a list: %T", out["results"])
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}
}

func TestSearchWikiError(t *testing.T) {
	session := connect(t, fakeSearcher{err: errors.New("boom")})

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "search_wiki",
		Arguments: searchInput{Query: "x"},
	})
	if err != nil {
		t.Fatalf("call tool transport: %v", err)
	}
	if !res.IsError {
		t.Fatal("want tool error, got success")
	}
}
