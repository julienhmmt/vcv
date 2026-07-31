# App skills

Task-specific checklists for agents working in this repository. Read the matching skill before starting a task and follow its verification steps.

## Format

Each skill is a Markdown file with YAML frontmatter:

```md
---
description: Short task description
---

# Title

## Goal

## Steps

## Verification
```

Keep skills action-oriented. Conventions that always apply belong in `AGENTS.md`, not here.

## Index

- **`vcv-high-coverage-tests.md`** — add or improve Go test coverage with `testify/mock`
- **`vcv-add-api-endpoint.md`** — add a JSON API endpoint end-to-end
- **`vcv-frontend-ui-change.md`** — add or change a Svelte UI component
- **`vcv-admin-settings-field.md`** — add or change an admin settings field safely
- **`vcv-security-check.md`** — security review before shipping
- **`vcv-debug-investigate.md`** — systematic debugging workflow

## Usage

When a task matches a skill, open the file and follow it. When conventions drift, update `AGENTS.md` (and the skill if the workflow changed).
