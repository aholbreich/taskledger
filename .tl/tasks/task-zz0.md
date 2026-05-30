---
id: task-zz0
title: 'Revise agents snippet content: remove make bdd leak, add --actor pattern, soften rigid language, add missing commands'
status: open
priority: medium
type: chore
created_at: 2026-05-30T18:24:09Z
updated_at: 2026-05-30T18:24:09Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - agents
references:
  - cmd/agents_snippet.md
  - cmd/agents.go
  - features/agents.feature
  - docs/usage.md
---

## Description

The tl agents snippet (cmd/agents_snippet.md) has several content issues:

1. 'Do not begin implementation from chat instructions alone if there is no matching tl task' — too rigid, fights real workflow. Reword to guide rather than forbid.
2. No mention of --actor flag. Agents that can't set env vars need --actor on each call. Add examples of both patterns (TL_ACTOR env vs --actor flag).
3. 'check the current @implemented set with make bdd before relying on them' — leaks implementation detail into agent workflow. Remove or replace with a simple 'tl <cmd> --help to verify availability.'
4. Missing commands: tl history <id> (mentioned in steps but should be listed explicitly), tl dep add/remove, tl stale, tl unblock, tl resolve.
5. Step 1 says 'run tl show <task-id> and tl history <task-id>' but 'history' is missing from the command listing in Step 6's preamble.
6. No context-economy mode — snippet is ~70 lines. Consider adding --compact flag (separate task scope, but the base snippet should be tighter).

See the critical review in session history for full context.
