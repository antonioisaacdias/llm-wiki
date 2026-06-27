# Wiki Conventions

Single source of truth for every writing agent (Claude on odin, Hermes on saga, and any future author). Notes are markdown files with YAML frontmatter, stored as `facts/<id>.md` in the vault repo. Reads are open; writes go through the write-path, which enforces the hard rules below.

## Note schema

```yaml
---
id: vram-overflow-cliff      # required, slug
type: fact                   # required, enum
description: one-line recall hint   # required in practice
tags: [ollama, debian]       # optional
status: active               # optional, defaults to active
superseded_by:               # required iff status=superseded
source: claude-code          # who wrote it
created: 2026-06-27T12:00:00Z   # auto-stamped, do not set by hand
modified: 2026-06-27T12:00:00Z  # auto-stamped on every write
---
Body in markdown. Links other notes with [[other-id]].
```

| Field | Required | Notes |
|-------|----------|-------|
| `id` | yes | slug, immutable identity, equals filename |
| `type` | yes | enum below |
| `description` | yes (practice) | one-line recall hint shown in search results |
| `tags` | no | topic tags, see vocabulary |
| `status` | no | `active` (default) or `superseded` |
| `superseded_by` | conditional | set iff `status=superseded` |
| `source` | no | `claude-code` \| `hermes` \| `human` |
| `created` | auto | stamped on first write only |
| `modified` | auto | stamped on every write |

`created` and `modified` are owned by the service. Anything you send is overwritten: `modified` becomes now, `created` is preserved from the existing file (or set to now on first write).

## `type` enum

- **fact** — objective state of the world that is true now (a config value, an IP, a hardware spec, a gotcha).
- **reference** — reusable how-to-recall or pointer (where a key lives, how to reuse a token, a catalog).
- **decision** — a choice that was made and why, so it is not relitigated.
- **procedure** — ordered steps to accomplish a task.

## `status` and supersession

- **active** — current. `superseded_by` must be empty.
- **superseded** — replaced; kept for history and excluded from search. `superseded_by` must name the note that replaces it.

When a fact changes, write the new note and flip the old one to `status: superseded`, `superseded_by: <new-id>`.

## `id` slug

Lowercase, must match `^[a-z0-9]+(-[a-z0-9]+)*$`: words of `[a-z0-9]` joined by single hyphens, no leading/trailing/double hyphens, no underscores, spaces or uppercase.

Good: `vram-overflow-cliff`, `homelab-truenas-storage-pool`, `npm-dns01-cloudflare`.
Bad: `Vram_Cliff`, `-leading`, `double--dash`, `has space`.

## Linking

- Reference another note in the body with `[[id]]`; an alias is allowed as `[[id|display text]]` (only the part before `|` is the target).
- For reciprocity add a relation line near the end:
  `**Relacionado:** [[a]], [[b]]`
- Forward-links are allowed and intentional: `[[note-not-written-yet]]` marks a note still to write. A broken link is therefore a **lint** signal, never a write error.
- When renaming a note (changing its `id`), scan inbound links and update every `[[old-id]]`.

## Tags

- Use node tags when a fact is node-specific: `odin`, `debian`, `truenas`, `saga`, `thinkpad`, `diaslabs`.
- Generic gotchas stay node-agnostic (no node tag).
- Reuse an existing tag instead of coining a variant (`ollama`, not `ollama-server`). Check search before inventing one.

## When to write

Write only facts that are **objective, durable, and shareable**:

- yes: infra state, decisions, procedures, reusable references, gotchas.
- no: agent behavior or identity (that is each agent's private memory).
- no: secrets or credential values.
- no: ephemeral session state.

## Lint

The graph is checked, never blocked, by lint:

- `GET /lint` (open, no token) — JSON report.
- MCP tool `lint_wiki` — same report.

Report fields: `notes`, `edges` (resolved links), `broken_links` (`{from_id, target}`), `orphans_in` (no inbound link), `orphans_out` (no outbound link).

Target: **0 `broken_links`** and ideally **0 `orphans_out`** (every note links out to at least one neighbor). Inbound orphans are tolerated for hub/index notes.
