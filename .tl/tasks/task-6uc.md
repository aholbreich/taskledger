---
id: task-6uc
title: Add tl remove command for deleting mistaken task files
status: done
priority: high
type: feature
created_at: 2026-05-31T18:43:45Z
updated_at: 2026-05-31T18:47:14Z
created_by: pi:remove
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - cli
references:
  - features/remove.feature
  - cmd/remove.go
  - bdd/remove_test.go
---

## Description

Add a BDD-specified tl remove command that deletes a task file from the active ledger with a required reason, appends a removed audit event, preserves existing event history, and uses --force for risky removals such as active claims, dependents, or non-cancelled tasks.

## Notes

- 2026-05-31T18:47:14Z [pi:remove] note: Implemented BDD-first tl remove: feature spec, cmd/remove.go, root wiring, history support for removed task files, completion coverage, README/docs usage updates. remove requires a reason, records removed event value, deletes task file, requires --force for non-cancelled tasks, active claims by other actors, and dependents. Validation: go test ./cmd ./bdd, make bdd, make test passed.
