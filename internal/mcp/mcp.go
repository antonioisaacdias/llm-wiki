package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/antonioisaacdias/llm-wiki/internal/httpapi"
	"github.com/antonioisaacdias/llm-wiki/internal/lint"
	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

const searchLimit = 8

type searchInput struct {
	Query string `json:"query" jsonschema:"the full-text query to search the wiki"`
}

type getInput struct {
	ID string `json:"id" jsonschema:"the identifier of the note to fetch"`
}

type searchOutput struct {
	Results []note.Stub `json:"results"`
}

type upsertInput struct {
	ID          string   `json:"id" jsonschema:"unique slug id of the note"`
	Type        string   `json:"type" jsonschema:"fact | reference | decision | procedure"`
	Description string   `json:"description" jsonschema:"one-line recall hint"`
	Tags        []string `json:"tags" jsonschema:"topic tags"`
	Body        string   `json:"body" jsonschema:"the note body in markdown"`
	Source      string   `json:"source" jsonschema:"who is writing: claude-code | hermes | human"`
}

type upsertOutput struct {
	OK bool   `json:"ok"`
	ID string `json:"id"`
}

type lintInput struct{}

func New(s httpapi.Searcher, wr httpapi.Writer, l httpapi.Lister) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "llm-wiki", Version: "0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_wiki",
		Description: "Search the wiki and return matching note stubs ranked by relevance.",
	}, searchHandler(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_note",
		Description: "Fetch a single wiki note by its id.",
	}, getHandler(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "upsert_note",
		Description: "Create or update a wiki note. Requires id and type.",
	}, upsertHandler(wr))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "lint_wiki",
		Description: "Report broken links and orphan notes across the wiki graph.",
	}, lintHandler(l))

	return server
}

func Handler(s httpapi.Searcher, wr httpapi.Writer, l httpapi.Lister) http.Handler {
	server := New(s, wr, l)
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, nil)
}

func searchHandler(s httpapi.Searcher) mcp.ToolHandlerFor[searchInput, searchOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, searchOutput, error) {
		stubs, err := s.Search(ctx, in.Query, searchLimit)
		if err != nil {
			return nil, searchOutput{}, fmt.Errorf("search_wiki %q: %w", in.Query, err)
		}
		return nil, searchOutput{Results: stubs}, nil
	}
}

func getHandler(s httpapi.Searcher) mcp.ToolHandlerFor[getInput, note.Note] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getInput) (*mcp.CallToolResult, note.Note, error) {
		n, err := s.Get(ctx, in.ID)
		if err != nil {
			return nil, note.Note{}, fmt.Errorf("get_note %q: %w", in.ID, err)
		}
		return nil, n, nil
	}
}

func lintHandler(l httpapi.Lister) mcp.ToolHandlerFor[lintInput, lint.Report] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ lintInput) (*mcp.CallToolResult, lint.Report, error) {
		notes, err := l.All(ctx)
		if err != nil {
			return nil, lint.Report{}, fmt.Errorf("lint_wiki: %w", err)
		}
		return nil, lint.Build(notes), nil
	}
}

func upsertHandler(wr httpapi.Writer) mcp.ToolHandlerFor[upsertInput, upsertOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in upsertInput) (*mcp.CallToolResult, upsertOutput, error) {
		n := note.Note{ID: in.ID, Type: in.Type, Description: in.Description, Tags: in.Tags, Body: in.Body, Source: in.Source}
		if err := wr.Upsert(ctx, n); err != nil {
			return nil, upsertOutput{}, fmt.Errorf("upsert_note %q: %w", in.ID, err)
		}
		return nil, upsertOutput{OK: true, ID: in.ID}, nil
	}
}
