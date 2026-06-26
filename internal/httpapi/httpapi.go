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

func New(s Searcher) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("GET /note/{id}", func(w http.ResponseWriter, r *http.Request) {
		n, err := s.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, n)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
