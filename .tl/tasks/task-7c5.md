---
id: task-7c5
title: 'Polish task-ID completion: filter closed, status prefix, preserve order'
status: done
priority: medium
type: task
created_at: 2026-05-29T12:39:02Z
updated_at: 2026-05-29T12:41:46Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - dx
  - completion
---

## Description

Live feedback on task-3q7's autocompletion: (1) closed tasks (`done`/`cancelled`) should not TAB-complete for mutating commands; (2) suggestions should show status as `[<status>] Title` in the description column (portable across shells, unlike ANSI color); (3) bash/zsh should preserve the priority-then-ID order returned by store.List via `cobra.ShellCompDirectiveKeepOrder`.

Per-command filter: `tl show` and `tl history` still suggest closed tasks (legitimate archival inspection); everything else (claim/close/note/block/cancel/unblock/pending/resolve/release/refine/dep) hides done+cancelled.

Approach: split helper into completeTaskIDs (active only) and completeAllTaskIDs (everything), with a shared inner function. Re-wire show.go and history.go to completeAllTaskIDs.

## Notes

- 2026-05-29T12:41:46Z [claude-code] note: Implemented: cmd/completion.go split into completeTaskIDs (active only, hides done/cancelled) + completeAllTaskIDs (everything, for show/history). Shared completeTaskIDsFiltered. Directive = NoFileComp | KeepOrder (preserves priority-then-ID order from store.List). Description format: 'id\t[<status>] <Title>'. show.go + history.go re-wired to completeAllTaskIDs. Feature scenarios updated + 3 new (mutating hides closed, show includes closed, history includes closed). Unit test updated for new directive value. 151/151 BDD scenarios pass; gofmt/vet clean. make install run so installed tl in PATH reflects new behavior.
