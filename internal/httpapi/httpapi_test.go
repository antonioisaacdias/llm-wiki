package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

type fakeWriter struct{ called bool }

func (f *fakeWriter) Upsert(_ context.Context, _ note.Note) error {
	f.called = true
	return nil
}

func TestSearchEndpoint(t *testing.T) {
	srv := httptest.NewServer(New(Deps{Search: fakeStore{}, Write: &fakeWriter{}, Token: "secret"}))
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
	srv := httptest.NewServer(New(Deps{Search: fakeStore{}, Write: &fakeWriter{}, Token: "secret"}))
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/search")
	if resp.StatusCode != 400 {
		t.Fatalf("status %d, want 400", resp.StatusCode)
	}
}

func TestNoteEndpoint(t *testing.T) {
	srv := httptest.NewServer(New(Deps{Search: fakeStore{}, Write: &fakeWriter{}, Token: "secret"}))
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/note/abc")
	var n note.Note
	json.NewDecoder(resp.Body).Decode(&n)
	if n.ID != "abc" {
		t.Fatalf("got %+v", n)
	}
}

func TestPostNoteRequiresToken(t *testing.T) {
	srv := httptest.NewServer(New(Deps{Search: fakeStore{}, Write: &fakeWriter{}, Token: "secret"}))
	defer srv.Close()
	body := `{"id":"n","type":"fact","description":"d"}`
	resp, _ := http.Post(srv.URL+"/note", "application/json", strings.NewReader(body))
	if resp.StatusCode != 401 {
		t.Fatalf("no token: status %d, want 401", resp.StatusCode)
	}
	req, _ := http.NewRequest("POST", srv.URL+"/note", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret")
	resp2, _ := http.DefaultClient.Do(req)
	if resp2.StatusCode != 201 {
		t.Fatalf("with token: status %d, want 201", resp2.StatusCode)
	}
}

func TestGetStaysOpen(t *testing.T) {
	srv := httptest.NewServer(New(Deps{Search: fakeStore{}, Write: &fakeWriter{}, Token: "secret"}))
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/search?q=vram")
	if resp.StatusCode != 200 {
		t.Fatalf("read should be open: %d", resp.StatusCode)
	}
}
