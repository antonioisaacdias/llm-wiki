package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/antonioisaacdias/llm-wiki/internal/httpapi"
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

func New(s httpapi.Searcher) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "llm-wiki", Version: "0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_wiki",
		Description: "Search the wiki and return matching note stubs ranked by relevance.",
	}, searchHandler(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_note",
		Description: "Fetch a single wiki note by its id.",
	}, getHandler(s))

	return server
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
