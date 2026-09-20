---
name: obsidian-vault
description: Build or refresh the VCV Obsidian brain - the notes that explain what VaultCertsViewer is, how it is built, and how it runs. Use when the user says "update the vault", "refresh the brain", "obsidian", "regenerate the VCV notes", or after a large change that makes the vault stale (new package, new route, new env var, version bump).
---

# VCV Obsidian Vault

Generate a set of linked Markdown notes that model VaultCertsViewer (vcv), so a
human can understand and operate it without reading the code.

## Where it lives

The vault is **outside this repo**:

```text
/Users/jh/Library/Mobile Documents/iCloud~md~obsidian/Documents/JhoBoxes/Projets/VCV
```

> [!important] iCloud Drive is TCC-protected
> iCloud Drive under `~/Library/Mobile Documents` is guarded by macOS TCC. A
> *shell* without Full Disk Access gets `Operation not permitted` on `ls`, not
> just on write. The file read/write tools may still reach it. Do not burn time
> debugging the shell.
>
> **Plan:** try the file tools first. If they work, write in place. If they are
> also blocked, stage into the repo at `.scratch/vcv-vault/` (`.scratch/` is
> gitignored), mirroring the vault layout, then give the user a one-line `cp` or
> a drag instruction.

## Do not invent content

Every note is derived from something in the repo. Read first, write second.

| Note group | Primary sources |
| --- | --- |
| Product | `README.md`, `README.fr.md`, `blogpost-en.md`, `app/internal/docs/ADMIN.md` |
| Architecture | `AGENTS.md`, `app/README.md`, `app/cmd/server/main.go` (`buildRouter`) |
| Data model | `app/internal/certs/certificate.go`, `app/internal/config/config.go` (`SettingsFile`) |
| Server | `app/internal/*` package files, `app/cmd/server/main.go` |
| API reference | `grep -n 'r\.\(Get\|Post\|Put\|Delete\|Handle\)' app/cmd/server/main.go app/internal/handlers/*.go` plus `app/internal/handlers/openapi.go` |
| Configuration | `app/internal/config/config.go` (the loader, not the docs), `settings.example.json` |
| Web | `app/web/frontend/src/`, `app/web/frontend/package.json` |
| Operations | `Makefile`, `app/Dockerfile`, `docker-compose.dev.yml`, `docker-compose.e2e.yml`, `vcv-deployment.yaml`, `.github/workflows/` |
| History | `git log`, `git tag`, `git rev-list --count HEAD` |

**Read the code, not the docs about the code.** The docs drift; this is not
hypothetical. Verify Go version and dependency count against `app/go.mod`, route
tables against `main.go`, and the version string against
`app/internal/version/version.go`. Every drift found goes into
`Reference/Open Questions and Drift.md` with the exact command to re-check it.

## Structure

```text
VCV Home.md                <- the entry point, a map of content
Product/                   Overview, Feature Inventory, Workflows, Design System, Accessibility and i18n
Architecture/              Overview, System Diagrams, Request Lifecycle, Data Model, Glossary
Server/                    Overview, Package Map, API Reference, Vault Layer, Certificate Model,
                           Metrics, Notifications, Middleware and Security, Admin API
Web/                       Overview, Route Map, Stores, Components, Utils
Operations/                Configuration, Deployment, Development Workflow, Testing, Troubleshooting
Reference/                 Project History, Open Questions and Drift, Reading the Codebase Cheaply
```

Create the folders only for notes that exist. A light folder tree plus a strong
`VCV Home.md` index is the balance: Obsidian's file explorer stays navigable,
and the graph view still works because every note links out.

## Conventions

- **Title case** for note names. Spaces, not hyphens.
- **`[[Wikilinks]]`** for every cross-reference. No relative Markdown links.
- **Every note ends with a `## Related` section** linking its neighbours.
- **Every index-style list is a wikilink list.**
- **Diagrams are text first.** ASCII in a fenced block for the version that
  renders anywhere, Mermaid in a ```mermaid block for the version Obsidian
  renders. `Architecture/System Diagrams.md` is the model: context, containers,
  a request sequence, the list flow, the admin flow, data flow, entities,
  deployment.
- **Callouts** (`> [!info]`, `> [!warning]`, `> [!tip]`) for the non-obvious
  fact on a page.
- **No em-dash (U+2014).** Use a spaced ASCII hyphen ` - `. This is a repo-wide
  rule and the vault follows it.
- **English.**
- Tables over prose for anything enumerable: env vars, routes, limits, tables,
  error codes.

## The facts worth carrying into every refresh

These are the ones a reader needs and that are easy to lose:

- **One Go binary embeds the Svelte SPA** (`app/web/dist` via `go:embed` in
  `app/web/web.go`). Rebuild the frontend (`make web-build`) before `go build`,
  or the binary serves a stale UI.
- **Public inventory APIs are unauthenticated on purpose** (private-network
  threat model). Only `/api/admin/*` needs the session cookie. Do not "fix" this
  without an explicit product decision.
- **Composite certificate IDs are `vault|mount:serial`.** Parse them carefully
  (`parseCertID` in the SPA, `extractVaultMountFromCertificateID` and
  `parseCompositeCertificateID` in Go).
- **Partial vault failure is normal.** `GET /api/certs` returns an envelope with
  `certificates` plus `errors[]`, so one dead vault warns instead of blanking the
  list.
- **All vaults (including disabled) get a client at startup** so the admin panel
  can toggle them without a restart. First enabled vault is primary.
- **Admin tokens are masked** in GET settings; PUT preserves the stored token
  when the field is blank or a mask sentinel (`***`, `__MASKED__`, ...).
- **Middleware order is load-bearing** and rate limiting is always on
  (health/ready/metrics and `/assets/` exempt).
- **Expiration thresholds (7/30 by default) come from `/api/config`** and the
  `config` store. Never hardcode them in UI logic.
- **`pnpm` only** under `app/web/frontend/`. Never npm or yarn.

## Refreshing an existing vault

1. Read `Reference/Open Questions and Drift.md` first and re-run its
   verification commands. Resolve or update each item.
2. Diff the sources: package list (`ls app/internal`), routes (`buildRouter`),
   config fields (`SettingsFile`), stores (`src/lib/stores`), components
   (`src/lib/components`), utils (`src/lib/utils`), make targets (`Makefile`).
3. Update only the notes whose sources moved. Append to
   `Reference/Project History.md`; do not rewrite history.
4. Report what changed and what you could not verify.

## Verify before handing over

```bash
cd <vault>

# every wikilink resolves to a real note.
# NOTE: use `while read`, not `for x in $(...)` - note names contain spaces,
# and a for-loop silently splits "Middleware and Security" into three words
# and reports three bogus dangling links.
grep -rho '\[\[[^]]*\]\]' . | sed 's/\[\[//;s/\]\]//;s/|.*//' | sort -u |
while IFS= read -r l; do
  [ -n "$(find . -name "$l.md" -print -quit)" ] || echo "DANGLING: $l"
done

# no em-dashes
grep -rn $'\u2014' . || echo "clean"

# mermaid fences balanced (opening count must equal closing count)
grep -rc '^```' . | awk -F: '{s+=$2} END {print "fence lines:", s, "(must be even)"}'

# markdownlint - MD013 (line length) off, wide tables are intentional.
# Write the config to a real file: markdownlint-cli2 rejects process
# substitution (`--config <(echo ...)`) with "Unable to use configuration file".
printf '{"MD013": false}\n' > /tmp/mdlint.json
npx --yes markdownlint-cli2 --config /tmp/mdlint.json "**/*.md"
```

### Four lint traps this vault has already hit

1. **Unescaped `|` inside a table row.** `` `GET|POST /api/*` `` splits the row
   into extra cells (MD056/MD060). Escape it: `` `GET\|POST` ``. Same for
   `?order=asc|desc`. This also bites **aliased wikilinks in tables**:
   `[[Server/Package Map|Package Map]]` injects a pipe. Prefer the bare form
   `[[Server/Package Map]]` in table cells.
2. **Bare URLs** (MD034). Wrap them: `<http://localhost:52000>`. URLs inside
   backticks are fine.
3. **Mixed emphasis** (MD049, style `consistent`). The *first* emphasis in a
   file sets the style for the whole file. Pick one: `_x_` or `*x*`. The vault
   uses `*x*` in prose and `_Avoid_:` in the glossary, so a lone `*x*` in a
   glossary-style file is the one that trips it.
4. **Bold pseudo-heading** (MD036). A line that is only `**In scope**` is
   flagged. Use a real heading (`### In scope`) instead.

### Obsidian rewrites links

Obsidian may rewrite wikilinks after they land: it can expand a link to an
absolute-vault path (`[[JhoBoxes/Projets/VCV/Product/Feature Inventory]]`) or
collapse it to a lowercased bare name (`[[Route map]]`). Both still resolve,
because Obsidian matches note names case-insensitively, but they look messy.

- Write the short form: `[[Folder/Note Name]]`, alias only when the display
  differs. Do not use an alias in a table cell (see trap 1).
- After a refresh, normalize absolute prefixes back out, then restore canonical
  case, then re-run the checks above.
- Verify link targets by **basename or folder/basename**, not by an exact-case
  match, or you will chase cosmetic false positives.

> [!tip] The shell may be TCC-blocked but not the file tools
> On macOS, `ls` / `find` on the iCloud vault can fail with "Operation not
> permitted" while reading and writing a known file path still works. Verify by
> iterating an explicit file list, not by `find`.

Fenced code blocks are invisible to these checks - scan fence-aware, or you
will "fix" ASCII diagrams.

Then tell the user the vault path and what you changed.
