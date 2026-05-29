---
id: task-reg
title: Add references field to task frontmatter
status: done
priority: medium
type: feature
created_at: 2026-05-29T11:29:22Z
updated_at: 2026-05-29T18:06:33Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - schema
  - cli
---

## Description

Add an optional `references` field to task frontmatter so a task can point at related files, URLs, ADRs, Gherkin features, ticket IDs — anything that gives an LLM or human enough context to find the artefacts behind the work.

## Semantics (decided)

References are **generic strings** — no schema, no rejection at input.
Anything goes: repo-relative file paths, full URLs (GitHub PR/JIRA/Confluence/RFC), ticket IDs (`JIRA-1234`), commit hashes, free text.

The user's intent shapes the value; tl just stores the list.

## Schema

Add `References []string` to `internal/task/task.go` with yaml:"references,omitempty" and json:"references,omitempty". Field is optional, list ordering preserved on write.

## CLI

- `tl create --ref <value>` (repeatable, like `--tag`)
- `tl <title> --ref <value>` — bare shorthand inherits the flag, mirroring how `--tag` already works at the root.
- `tl refine --add-ref <value>` (repeatable; idempotent — adding an existing ref is a no-op)
- `tl refine --remove-ref <value>` (repeatable; idempotent — removing a missing ref is a no-op)

Pattern mirrors `tl dep add/remove` for symmetry.

## Display

`tl show` adds a `References:` field between `Depends On` and `Claim` in human output, formatted like the Depends On block:

```
ID: task-abc
Status: open
Depends On: none
References:
  - src/auth/login.go
  - features/login.feature
  - https://github.com/aholbreich/tl/pull/42
  - JIRA-1234
Claim: none
```

When the list is empty, the field is shown as `References: none` (matching the Depends On convention).

JSON output includes `references` as a top-level array; empty list emits `[]`.

The task body markdown is unchanged — frontmatter is the source of truth.

## Validation (in tl doctor — depends on task-4sh)

Each reference is classified by a simple heuristic and validated only when it can be checked:

- URL-shaped — starts with `<scheme>:` matching `[a-z][a-z0-9+.-]*:` (`http://`, `https://`, `mailto:`, `ftp://`, …) → **skip**, no network calls.
- Path-shaped — contains `/` and no scheme prefix → treated as **repo-relative file path** (relative to the parent of `.tl/`); doctor reports a `references` issue with severity `warning` if the file does not exist.
- Bare identifier or free text — anything else (`JIRA-1234`, `ABC-789`, `"see Anna's comment"`) → **skip**.

`tl doctor --fix` removes dead file-path references (the unfixable URL/identifier cases are reported as `fixable: false`).

## Events

- `reference_added` on `tl create --ref` and `tl refine --add-ref` (one event per reference added)
- `reference_removed` on `tl refine --remove-ref`

Standard event shape (actor, time, task_id, event).

## Process

- Write `features/references.feature` first; agree the Gherkin before implementation.
- Once references is implemented, extend `features/doctor.feature` with the validation scenarios above.
- Tag both `@implemented` only when their respective code passes.

## What to implement

### 1. Schema change
Add `References []string` field to the Task struct in internal/task/task.go with yaml:"references" and json:"references" tags. Optional (omitempty).

### 2. CLI: tl create
- `--ref <path>` flag (repeatable, like `--tag`)
- Sets initial references on the new task

### 3. CLI: tl refine
- `--add-ref <path>` flag (repeatable) — appends a reference
- `--remove-ref <path>` flag (repeatable) — removes a reference
- Follows the same pattern as tl dep add/remove for idempotency

### 4. Display: tl show
- Renders a `References:` section listing all references
- JSON output includes the `references` field

### 5. Validation: tl doctor
- Detect referenced file paths that do not exist (dead link detection)
- Optionally clear dead references with `--fix`

### 6. Events
- `tl refine --add-ref` records event `reference_added`
- `tl refine --remove-ref` records event `reference_removed`

## Process
- Write Gherkin feature file first (features/references.feature)
- Follow BDD approach: Gherkin → implement step defs → implement command
- Tag @implemented when done

## Notes

- 2026-05-29T11:29:26Z [main-pc] note: Created from traceability analysis. Implements Option C: references field as plain string list. Follows BDD approach — Gherkin first, then implementation. Connected to task-4sh (tl doctor design) as references validation will be part of tl doctor.
- 2026-05-29T14:04:59Z [main-pc] note: The references field from this task and the external field from the sync PRD are related but distinct concepts. See docs/import-sync-PRD.md section 8 and open question Q19 for the distinction.
- 2026-05-29T14:05:17Z [claude-code] note: Design refined and Gherkin drafted. features/references.feature has 16 scenarios covering create (single/multi/bare/empty), refine (add/remove/idempotent/mixed), display (frontmatter field between Depends On and Claim; 'none' when empty; JSON array), and events (one per add/remove; idempotent does not emit). Cross-cutting reference validation in tl doctor is sketched in this task's description but the scenarios live in features/references.feature only after both task-reg and task-4sh land — whichever ships second extends the partner's feature.
- 2026-05-29T18:06:33Z [claude-code:references] note: Implemented the references field end to end. Schema: References []string with yaml:"references,omitempty" json:"references"; store.Read normalizes nil->[] so JSON always emits an array (keeps existing task/marshal tests green and avoids churning the 17 existing task files). Added Event.Value (json:"value,omitempty") to carry the reference on reference_added/reference_removed. CLI: create --ref (+ bare shorthand via root), refine --add-ref/--remove-ref (idempotent, mirror dep add/remove). show renders References between Depends On and Claim ('none' when empty). Events: reference_added per ref on create (with actor) and refine add; reference_removed on refine remove; idempotent ops write nothing and emit no event. refine ref events carry no actor, mirroring dependency_added (refine has no actor plumbing). features/references.feature tagged @implemented; all 17 scenarios green; full suite 205 pass. Docs: usage.md References section un-stubbed, README command listing updated. DEFERRED: tl doctor reference validation (dead file-path detection + --fix) belongs to task-4sh — see note there.
