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

	"github.com/antonioisaacdias/llm-wiki/internal/httpapi"
	"github.com/antonioisaacdias/llm-wiki/internal/index"
	wikimcp "github.com/antonioisaacdias/llm-wiki/internal/mcp"
	"github.com/antonioisaacdias/llm-wiki/internal/vault"
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

	notes, err := vault.Load(vaultDir)
	if err != nil {
		return err
	}
	store, err := index.Open(dbPath)
	if err != nil {
		return err
	}
	defer store.Close()
	if err := store.Build(context.Background(), notes); err != nil {
		return err
	}
	slog.Info("indexed vault", "notes", len(notes), "dir", vaultDir)

	root := http.NewServeMux()
	mcpHandler := wikimcp.Handler(store)
	root.Handle("/mcp", mcpHandler)
	root.Handle("/mcp/", mcpHandler)
	root.Handle("/", http.TimeoutHandler(httpapi.New(store), 15*time.Second, "request timeout"))

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
		if err := wikimcp.New(store).Run(ctx, &sdkmcp.StdioTransport{}); err != nil {
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
