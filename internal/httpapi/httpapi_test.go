package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

type fakeStore struct{}

func (fakeStore) Search(_ context.Context, q string, _ int) ([]note.Stub, error) {
	if q == "vram" {
		return []note.Stub{{ID: "vram-cliff", Description: "d", Type: "fact", Score: -1}}, nil
	}
	return nil, nil
}
func (fakeStore) Get(_ context.Context, id string) (note.Note, error) {
	return note.Note{ID: id, Body: "hello"}, nil
}

func TestSearchEndpoint(t *testing.T) {
	srv := httptest.NewServer(New(fakeStore{}))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/search?q=vram")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var stubs []note.Stub
	json.NewDecoder(resp.Body).Decode(&stubs)
	if len(stubs) != 1 || stubs[0].ID != "vram-cliff" {
		t.Fatalf("got %+v", stubs)
	}
}

func TestSearchMissingQuery(t *testing.T) {
	srv := httptest.NewServer(New(fakeStore{}))
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/search")
	if resp.StatusCode != 400 {
		t.Fatalf("status %d, want 400", resp.StatusCode)
	}
}

func TestNoteEndpoint(t *testing.T) {
	srv := httptest.NewServer(New(fakeStore{}))
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/note/abc")
	var n note.Note
	json.NewDecoder(resp.Body).Decode(&n)
	if n.ID != "abc" {
		t.Fatalf("got %+v", n)
	}
}
