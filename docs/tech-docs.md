
## BDD Style Features List

See [`features/`](features/) - Gherkin behavioral spec, one file per command
  
##  Status model

| Status          | Meaning                                                  |
| --------------- | -------------------------------------------------------- |
| `open`          | Work exists but is not currently claimed.                |
| `in_progress`   | Work is claimed by an actor.                             |
| `blocked`       | Work cannot continue (dependency or technical blocker).  |
| `pending_human` | Work requires human clarification or decision.           |
| `done`          | Work is completed.                                       |
| `cancelled`     | Work is intentionally abandoned.                         |

A task is **ready** (eligible for `tl ready` / `tl claim`) when:

- status is `open` (or `in_progress` with an expired claim), **and**
- all dependencies are `done`, **and**
- it has no active claim.

All six statuses are reachable through dedicated commands — see
[`COMMANDS.md`](COMMANDS.md) for the per-command transitions. `tl pending`
and `tl block` release the active claim when they flip status; the
collaborating actor re-claims after `tl resolve` or `tl unblock` if they
want to resume.

The task file is the current state; `.taskledger/events.jsonl` is the
append-only audit trail. Every mutating command appends one JSON line;
`tl history` reads them back.

Long-running work renews its lease by re-running `tl claim` as the same
actor — there is no separate heartbeat command.

---

## Storage

```
.taskledger/
  config.yaml      # defaults
  tasks/
    task-<3>.md    # one file per task (Markdown + YAML frontmatter)
  events.jsonl     # append-only audit trail
```

A created task looks like:

```markdown
---
id: task-x3n
title: Add login validation
status: open
priority: medium
type: ""
created_at: 2026-05-17T00:45:40Z
updated_at: 2026-05-17T00:45:40Z
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

Validate email format and require a password.
```

As a task moves through its lifecycle, frontmatter and body gain fields.

A `pending_human` task records the question structurally so `tl resolve`
can consume it; the claim is released while waiting:

```yaml
status: pending_human
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
pending:
  question: Which OAuth provider should we ship first?
  requester: claude-code:frontend
  requested_at: 2026-05-17T01:15:22Z
```

A `blocked` task carries no extra frontmatter — the status is the signal,
and the blocker reason lives in the body as a normal note appended by
`tl block`:

```markdown
status: blocked
```

```markdown
## Notes

- 2026-05-17T02:11:08Z [claude-code:main] blocked: Waiting on upstream library release (tracking GH issue 412).
```
