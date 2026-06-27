package mcp

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

func connectHTTP(t *testing.T, s fakeSearcher) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	srv := httptest.NewServer(Handler(s, &fakeWriter{}, s))
	t.Cleanup(srv.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: srv.URL}, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func TestHTTPListTools(t *testing.T) {
	session := connectHTTP(t, fakeSearcher{})

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}

	got := map[string]bool{}
	for _, tool := range res.Tools {
		got[tool.Name] = true
	}
	for _, want := range []string{"search_wiki", "get_note", "upsert_note"} {
		if !got[want] {
			t.Errorf("missing tool %q; got %v", want, got)
		}
	}
}

func TestHTTPSearchWiki(t *testing.T) {
	s := fakeSearcher{stubs: []note.Stub{
		{ID: "n1", Description: "first", Type: "doc", Score: 1.5},
		{ID: "n2", Description: "second", Type: "doc", Score: 0.9},
	}}
	session := connectHTTP(t, s)

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
