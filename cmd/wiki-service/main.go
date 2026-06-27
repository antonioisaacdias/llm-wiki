package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/antonioisaacdias/llm-wiki/internal/gitrepo"
	"github.com/antonioisaacdias/llm-wiki/internal/httpapi"
	"github.com/antonioisaacdias/llm-wiki/internal/index"
	wikimcp "github.com/antonioisaacdias/llm-wiki/internal/mcp"
	"github.com/antonioisaacdias/llm-wiki/internal/vault"
	"github.com/antonioisaacdias/llm-wiki/internal/writer"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	vaultDir := os.Getenv("WIKI_VAULT")
	if vaultDir == "" {
		return errors.New("WIKI_VAULT is required")
	}
	dbPath := envOr("WIKI_DB", ":memory:")
	addr := envOr("WIKI_HTTP_ADDR", ":8080")
	token := os.Getenv("WIKI_WRITE_TOKEN")
	push := os.Getenv("WIKI_GIT_PUSH") == "1"

	store, err := index.Open(dbPath)
	if err != nil {
		return err
	}
	defer store.Close()

	reindex := func(ctx context.Context) error {
		notes, err := vault.Load(vaultDir)
		if err != nil {
			return err
		}
		return store.Build(ctx, notes)
	}
	if err := reindex(context.Background()); err != nil {
		return err
	}
	slog.Info("indexed vault", "dir", vaultDir)

	wr := writer.New(vaultDir, push, reindex)

	root := http.NewServeMux()
	mcpHandler := httpapi.RequireToken(token, wikimcp.Handler(store, wr, store))
	root.Handle("/mcp", mcpHandler)
	root.Handle("/mcp/", mcpHandler)
	root.Handle("POST /reindex", httpapi.RequireToken(token, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if push {
			if err := gitrepo.Pull(r.Context(), vaultDir); err != nil {
				slog.Warn("reindex: git pull failed, reindexing from disk", "err", err)
			}
		}
		if err := reindex(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})))
	root.Handle("/", http.TimeoutHandler(httpapi.New(httpapi.Deps{Search: store, Write: wr, List: store, Token: token}), 15*time.Second, "request timeout"))

	srv := &http.Server{
		Addr:         addr,
		Handler:      root,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("http listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http", "err", err)
		}
	}()

	if os.Getenv("WIKI_MCP") == "stdio" {
		slog.Info("mcp serving on stdio")
		if err := wikimcp.New(store, wr, store).Run(ctx, &sdkmcp.StdioTransport{}); err != nil {
			slog.Error("mcp", "err", err)
		}
	} else {
		<-ctx.Done()
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutCtx)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
