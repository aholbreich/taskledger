---
id: task-dsp
title: Implement tl refine command to update tasks title and description
status: done
priority: medium
created_at: 2026-05-22T21:11:13Z
updated_at: 2026-05-24T16:24:22Z
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

Implement a `tl refine` command so task fields can be edited from the CLI instead
of hand-editing `.md` files. 

**Contract:** `features/refine.feature` - already written, intentionally left **untagged** until the command is built. Tag it `@implemented` as the final step.

**Editable fields:** `--title`/`-t`, `--description`/`-d`, `--type`. . Status is **not** editable
here - it stays owned by the dedicated lifecycle commands (`claim`, `close`,
`block`, `pending`, `resolve`, `unblock`, `cancel`).

**Steps:**
1. Add step definitions in `bdd/bdd_test.go` for the new phrasings the feature
   introduces - e.g. `"<id>" has title "..."`, `has the description "..."`,
   `does not have the tag "..."`, `with tags "..." and "..."`, and the
   "no fields were given to update" assertion.
2. Implement `cmd/refine.go`: load the task, apply only the provided flags,
   validate priority (reuse `create`'s validation; exit 2 on invalid), reject an
   unknown id (exit 3), reject when no editable flag is given (exit 2), append an
   `updated` event to `events.jsonl`, bump `updated_at`, and support `--json`.
3. Tag `features/refine.feature` `@implemented` and run `make bdd`.


- `refine` is allowed on a task claimed by another active 

## Notes

- 2026-05-24T16:24:22Z [pi:refine-command] note: Implemented tl refine command with title/description/type/priority updates, JSON output, refined events, BDD steps, and docs. Tagged refine.feature @implemented. Verification: make bdd, go test ./..., and make test passed.
