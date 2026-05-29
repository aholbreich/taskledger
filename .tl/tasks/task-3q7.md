---
id: task-3q7
title: add autocompletion for task IDs
status: done
priority: medium
type: task
created_at: 2026-05-17T20:50:50Z
updated_at: 2026-05-29T12:17:28Z
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

## Notes

- 2026-05-29T11:29:56Z [claude-code] note: Scope clarification: implementing completion for refine in addition to the 12 commands listed in the description (refine also takes a task ID, added after this task was written). Total: 13 single-arg commands + dep add/remove positional + --on flag completion. Approach: one shared completeTaskIDs helper returns canonical task-IDs with title as description; ledger missing => silent zero suggestions; bare short-code completion (typing 'abc<TAB>') intentionally NOT supported (users type 't<TAB>' or 'task-...'). Self-dependency filter for dep --on deferred unless trivial.
- 2026-05-29T12:06:53Z [claude-code] note: Implementation complete. Files: cmd/completion.go (shared completeTaskIDs helper); cmd/completion_test.go (unit test for missing-ledger case); features/completion.feature (8 BDD scenarios, @implemented); bdd/completion_test.go (step defs); bdd/bdd_test.go (splitArgs preserves explicit empty quoted args, e.g. for 'tl __complete show ""'; registered initializeCompletionSteps); ValidArgsFunction wired into show/claim/close/note/history/block/cancel/unblock/pending/resolve/release/refine/dep add/dep remove; --on flag completion registered on dep add/dep remove. Suggestions format: 'task-id\tTitle'; directive: ShellCompDirectiveNoFileComp; ledger-missing => silent zero suggestions. README updated with shell completion install snippets. Tests: 175/175 pass; gofmt/vet clean.
- 2026-05-29T12:17:28Z [claude-code] note: Added bare short-code completion per user choice: typing 'ab<TAB>' now suggests 'abc123' (without the task- prefix), matching the existing input flexibility documented for show. Logic: if toComplete starts with task- (or is a prefix of task-), emit canonical task-xxx; otherwise emit bare xxx. New feature scenario 'Completion supports bare short codes'. Tests: 148/148 BDD pass + 175/175 total; gofmt/vet clean. Empirical verification across paths: empty/'t'/'task-5' => canonical; '5'/'ab' => bare.
