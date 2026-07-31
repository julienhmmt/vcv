# App skills

Task-specific checklists for agents working in this repository.

Claude Code auto-discovers these: each skill is `<name>/SKILL.md` with `name` + `description` frontmatter, and the model loads the matching one on its own. Other harnesses (Codex, Devin, Cursor) should read the matching file directly.

## Index

| Skill | Use when |
| --- | --- |
| `vcv-add-api-endpoint` | adding or changing a JSON API route, end to end |
| `vcv-frontend-ui-change` | Svelte component, store, or util work |
| `vcv-i18n-string` | any user-visible copy change (five languages) |
| `vcv-admin-settings-field` | admin settings gains or changes a field |
| `vcv-high-coverage-tests` | adding Go coverage or fixing a Go test |
| `vcv-security-check` | before shipping anything touching handlers/middleware/config |
| `vcv-debug-investigate` | investigating a bug, regression, or flaky test |

## Writing a skill

- `description` is the trigger — say **what** and **when to use it**, in the words a request would use. It is the only part the model sees before deciding to load the file.
- Keep it action-oriented: steps, then a runnable **Verification** block.
- Encode the traps that cost time (silent PUT drops, blank translations, nil-panicking mocks), not the things a compiler already catches.
- Conventions that always apply belong in `AGENTS.md`, not here. Skills are per-task.
- Cite real symbols and paths, and re-check them when the code moves — a stale skill is worse than none.
