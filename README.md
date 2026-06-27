# llm-wiki

A shared, factual knowledge base for LLM agents. Notes are plain Markdown in git;
a small Go daemon indexes them and serves search and retrieval over **HTTP REST**
and the **Model Context Protocol (MCP)** — so heterogeneous agents (and humans) read
and write the same knowledge.

## Why

Agents that don't share memory keep relearning the same facts. `llm-wiki` is a single
service multiple agents consult and maintain: one writes a fact, the others can find it.
Markdown + git keeps the content portable, diffable, and reviewable — no proprietary store.

## Architecture

```
            ┌──────────────┐      ┌──────────────┐
   HTTP ───►│              │      │              │
            │  wiki-service│─────►│ SQLite FTS5  │  in-memory index, rebuilt from disk
   MCP  ───►│    (Go)      │      │              │
            └──────┬───────┘      └──────────────┘
                   │ read / write + commit
            ┌──────▼───────┐
            │ vault (.md)  │  markdown notes, git-versioned (source of truth)
            └──────────────┘
```

- **One domain layer** (`internal/index`) owns the FTS5 index; two thin transports
  (`internal/httpapi`, `internal/mcp`) expose it — one logic, two protocols.
- **Writes go through the service**, which validates the note, writes the `.md`,
  commits (and optionally pushes) to git, then reindexes. A single writer serializes
  writes; the index is `RWMutex`-guarded.
- **Reads are open; writes are token-gated** (bearer token).

## Note schema

Each note is one atomic Markdown file with YAML frontmatter:

```yaml
---
id: vram-overflow-cliff          # unique slug = filename
type: fact                       # fact | reference | decision | procedure
description: one-line recall hint (weighted highest in search)
tags: [topic, area]
status: active                   # active | superseded
superseded_by: null              # id of the note that replaced this one
source: claude-code              # who wrote it
created: 2026-06-25T00:00:00Z
modified: 2026-06-26T00:00:00Z
---

Self-contained body in Markdown. Link related notes with [[other-note]].
```

Search ranks `description` and `tags` above the body (BM25), and returns only `active` notes.

## API

| Method & path | Auth | Purpose |
|---|---|---|
| `GET /search?q=` | open | ranked note stubs (`id`, `description`, `type`, `score`) |
| `GET /note/{id}` | open | full note as JSON |
| `GET /lint` | open | graph report: broken links, orphans, counts |
| `GET /healthz` | open | liveness |
| `POST /note` | bearer | create/update a note (JSON body) |
| `POST /reindex` | bearer | `git pull` + rebuild the index (webhook target) |
| `/mcp` | bearer | MCP over streamable HTTP |

**MCP tools:** `search_wiki`, `get_note`, `upsert_note`, `lint_wiki`. Also available
over stdio (`WIKI_MCP=stdio`).

Note schema, enums, slug and linking rules live in [`CONVENTIONS.md`](CONVENTIONS.md).

## Configuration (env)

| Var | Default | Meaning |
|---|---|---|
| `WIKI_VAULT` | *(required)* | path to the markdown vault (a git checkout) |
| `WIKI_DB` | `:memory:` | SQLite path for the FTS5 index |
| `WIKI_HTTP_ADDR` | `:8080` | HTTP listen address |
| `WIKI_WRITE_TOKEN` | *(empty)* | bearer token required for writes / `/mcp` |
| `WIKI_GIT_PUSH` | *(off)* | `1` to push to the vault remote after each commit |
| `WIKI_MCP` | *(off)* | `stdio` to also serve MCP on stdio |

## Build & run

```sh
go build ./...
go test ./...

WIKI_VAULT=./vault go run ./cmd/wiki-service
# or
docker build -t llm-wiki . && docker run -p 8080:8080 -v "$PWD/vault:/vault" llm-wiki
```

## Tech

Go (stdlib `net/http`, `database/sql`, `log/slog`), `modernc.org/sqlite` (pure-Go, FTS5),
`github.com/modelcontextprotocol/go-sdk`, `gopkg.in/yaml.v3`.
