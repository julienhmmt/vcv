# VaultCertsViewer Constitution

## Core Principles

### I. Single Binary Architecture

The application MUST be distributed as a single Go binary that embeds the static HTML/CSS/JS frontend. There is no separate frontend build step or Node.js runtime required in the final production artifact.

### II. Frontend Minimalism

The frontend MUST use plain HTML, CSS, and Vanilla JavaScript with HTMX. Do not introduce heavy frontend frameworks, bundlers (like Webpack/Vite), or Node.js dependencies for the UI.

### III. Read-Only Visibility

VaultCertsViewer is designed for observability. It MUST default to a read-only view of the Vault/OpenBao certificates. Any state-changing actions (like the admin panel) MUST be strictly protected, optional, and separate from the core viewing experience.

### IV. Testing Rigor

All Go code MUST target >90% test coverage. Tests MUST be table-driven where applicable, using `testify/assert` and `testify/mock`. Black-box testing with the `_test` package suffix is preferred.

### V. Multi-Tenant Compatibility

The core logic MUST support querying and aggregating data from multiple Vault/OpenBao instances and multiple PKI mounts concurrently. Single-vault assumptions MUST NOT be hardcoded into the data models or metrics.

## Development Standards

- **Go**: Keep functions single-purpose (<20 lines). Avoid deeply nested blocks (use early returns). Avoid using primitive types excessively. Use RO-RO (Receive Object, Return Object) where applicable.
- **Frontend**: Use single-hyphen class names (e.g., `vcv-button-primary`). Use modern CSS color-function notation. Ensure consistency between `app.js` and `styles.css`.
- **Naming**: PascalCase for structs/interfaces (public), camelCase for variables/functions, kebab-case for files/directories.

## UI / UX Standards

- The application MUST support internationalization (i18n).
- Real-time updates and interactivity SHOULD be handled via HTMX and minimal Vanilla JS.
- Certificate counts, connection statuses, and metrics MUST be exposed transparently to the user and via Prometheus.

## Governance

All Pull Requests MUST be reviewed for compliance with the Core Principles and Development Standards.

Changes to the constitution require a version bump according to semantic versioning.

Complexity must be justified. Use `app/README.md` for runtime development guidance.
