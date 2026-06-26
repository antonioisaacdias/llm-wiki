package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

type fakeSearcher struct {
	stubs []note.Stub
	note  note.Note
	err   error
}

func (f fakeSearcher) Search(_ context.Context, _ string, _ int) ([]note.Stub, error) {
	return f.stubs, f.err
}

func (f fakeSearcher) Get(_ context.Context, _ string) (note.Note, error) {
	return f.note, f.err
}

func connect(t *testing.T, s fakeSearcher) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	server := New(s)
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
	for _, want := range []string{"search_wiki", "get_note"} {
		if !got[want] {
			t.Errorf("missing tool %q; got %v", want, got)
		}
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
