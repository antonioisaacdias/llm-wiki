package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/antonioisaacdias/llm-wiki/internal/note"
)

type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]note.Stub, error)
	Get(ctx context.Context, id string) (note.Note, error)
}

type Writer interface {
	Upsert(ctx context.Context, n note.Note) error
}

type Deps struct {
	Search Searcher
	Write  Writer
	Token  string
}

func New(d Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("GET /search", searchHandler(d.Search))
	mux.Handle("GET /note/{id}", getHandler(d.Search))
	mux.Handle("POST /note", RequireToken(d.Token, postNote(d.Write)))
	return mux
}

func searchHandler(s Searcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			http.Error(w, "missing query param q", http.StatusBadRequest)
			return
		}
		stubs, err := s.Search(r.Context(), q, 8)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, stubs)
	})
}

func getHandler(s Searcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, err := s.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, n)
	})
}

func RequireToken(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" || r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func postNote(wr Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var n note.Note
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := wr.Upsert(r.Context(), n); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
