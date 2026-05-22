---
id: task-3q7
title: add autocompletion for task IDs
status: open
priority: medium
type: task
created_at: 2026-05-17T20:50:50Z
updated_at: 2026-05-22T21:04:26Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags: []
---
## Description

Add dynamic shell completion of task IDs as positional arguments. Today
`tl completion <shell>` only completes flag names and subcommands; typing a task
ID (e.g. `tl show <TAB>`) suggests nothing.

**Goal:** pressing TAB where a task ID is expected suggests the actual task IDs
from the current ledger.

**Scope:** every command that takes a task-id argument — `show`, `claim`,
`close`, `note`, `history`, `block`, `cancel`, `unblock`, `pending`, `resolve`,
`release`, and `dep` (add/remove).

**Implementation hint:** register a cobra `ValidArgsFunction` on each such
command that reads the ledger (reuse `store.List`) and returns matching IDs;
consider annotating each suggestion with the task title for context. Factor the
lookup into one shared helper so all commands stay consistent.
