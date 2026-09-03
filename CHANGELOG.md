# Changelog

All notable changes to VaultCertsViewer (vcv) are documented here.
This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.9.1] - 2026-09-03

Maintenance release. No user-facing behavior change; refreshes dependencies and
internal code quality. Safe in-place upgrade from 1.9.

### Changed

- Bumped version to 1.9.1 across deployment files, READMEs, and `version.go`.

### Dependencies

- `github.com/go-chi/chi/v5` 5.3.1 → 5.3.2
- `github.com/stretchr/testify` (latest)
- `github.com/prometheus/procfs` 0.21.1 → 0.22.0
- `github.com/prometheus/common` 0.70.1 → 0.71.0
- `github.com/prometheus/client_model` 0.6.2 → 0.6.3
- `github.com/mattn/go-isatty` 0.0.22 → 0.0.24
- `github.com/yuin/goldmark` 1.8.5 → 1.8.6
- `golang.org/x/crypto` 0.54.0 → 0.56.0
- `golang.org/x/net` 0.57.0 → 0.58.0
- `google.golang.org/protobuf` 1.36.11 → 1.36.12
- `go` directive 1.25.4 → 1.26.0 (aligns with Dockerfile `golang:1.26.4-alpine`)
- `bits-ui` 2.18.1 → 2.19.0
- `svelte` 5.56.8 → 5.57.0
- `svelte-sonner` 1.1.1 → 1.2.1
- `vite` 8.1.5 → 8.2.2
- `@lucide/svelte`, `@types/node`, `tailwind-variants`, `jsdom`,
  `@testing-library/jest-dom`, `svelte-check`, `@sveltejs/vite-plugin-svelte`
  bumped to latest.

### Internal

- Resolved SonarQube Go findings (S3776 cognitive complexity, S1192 string
  duplication, S1186 empty functions, S107 function length) in
  `internal/vault/multi.go` and `internal/vault/real.go`.
- Resolved SonarQube web findings in Svelte components and utils.
- Added local SonarQube analysis service (Go + web projects) for development
  use only; not shipped in the binary.
- Fixed `make go-update` target: `go get -u all` → `go get -u ./...` to avoid
  failures upgrading transitive test deps (e.g. `prometheus/procfs` removing
  `internal/util`).

## [1.9]

- Expiry timeline component showing upcoming certificate expirations.
- Sortable certificate table column headers.
- Web reliability fixes.
- Admin login, certificate sources selector, and signing authority (CA) view.
