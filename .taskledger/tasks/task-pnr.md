---
id: task-pnr
title: Test autocompletion
status: open
priority: medium
created_at: 2026-05-17T21:06:38Z
updated_at: 2026-05-22T21:04:26Z
created_by: human
assignee: null
depends_on:
  - task-3q7
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags: []
---

## Description

Verify and document the task-ID shell completion delivered in task-3q7.

**Deliverable:**
- Add automated test coverage for the completion logic — assert the shared
  `ValidArgsFunction` returns the expected task IDs for a given ledger state
  (and an empty/partial-match ledger).
- Manually verify `tl <cmd> <TAB>` completes task IDs in bash, zsh, and fish.
- Add a README section documenting how to install/enable shell completion
  (`tl completion <shell>`) and the task-ID completion behaviour.

Depends on task-3q7 — the completion feature must exist before it can be tested
or documented.
