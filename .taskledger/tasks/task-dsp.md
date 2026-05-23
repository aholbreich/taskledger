---
id: task-dsp
title: Implement tl update command for editing task fields
status: open
priority: medium
created_at: 2026-05-22T21:11:13Z
updated_at: 2026-05-22T21:11:13Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - cli
---

## Description

Implement a `tl update` command so task fields can be edited from the CLI instead
of hand-editing `.md` files. (This is the gap behind the earlier manual
description/title corrections — there is currently no edit command.)

**Contract:** `features/update.feature` — already written, intentionally left
**untagged** until the command is built. Tag it `@implemented` as the final step.

**Editable fields:** `--title`/`-t`, `--description`/`-d`, `--priority`, `--type`,
`--add-tag` (repeatable), `--remove-tag` (repeatable). Status is **not** editable
here — it stays owned by the dedicated lifecycle commands (`claim`, `close`,
`block`, `pending`, `resolve`, `unblock`, `cancel`), per PRD §5.

**Steps (per CLAUDE.md "Adding a command"):**
1. Add step definitions in `bdd/bdd_test.go` for the new phrasings the feature
   introduces — e.g. `"<id>" has title "..."`, `has the description "..."`,
   `does not have the tag "..."`, `with tags "..." and "..."`, and the
   "no fields were given to update" assertion.
2. Implement `cmd/update.go`: load the task, apply only the provided flags,
   validate priority (reuse `create`'s validation; exit 2 on invalid), reject an
   unknown id (exit 3), reject when no editable flag is given (exit 2), append an
   `updated` event to `events.jsonl`, bump `updated_at`, and support `--json`.
3. Tag `features/update.feature` `@implemented` and run `make bdd`.

**Design questions to confirm:**
- Tag surface: `--add-tag`/`--remove-tag` (as specced) vs a single `--tag` that
  replaces the whole set.
- Should `update` require a claim, or be allowed on a task claimed by another
  active actor? Currently the spec leaves it unconstrained.
