# Command reference

User-visible flags for each implemented `tl` command. For canonical per-command behavior, read the matching file in [`../features/`](../features); the feature file wins if this page drifts. Every command also accepts
`--help` at the terminal.

---

## `tl init`

Initialize a `.taskledger/` ledger in the current directory.

```
(no flags)
```

## `tl create [title] [options]`

Create a new task. The title is a required positional argument or passed via
`--title`.

```
    --title              Task title (required, or positional argument)
-d, --description        Task description (stored under ## Description)
-p, --priority           Task priority (l|low, m|medium, h|high) [default: medium]
-t, --type               Task type
    --tag                Tag to apply (repeatable)
    --actor              Creator actor [default: human]
    --json               Emit JSON output
```

## `tl list`

List active tasks in the ledger, sorted by priority then identifier. Closed
statuses (`done`, `cancelled`) are hidden by default. Passing `--status`
with a closed status reveals matching tasks without needing `--all`. Human
output includes `ID`, `Status`, `Priority`, `Claimed By`, and `Title`.

```
-a, --all                Include closed tasks (done and cancelled)
    --status             Only show tasks with this status (e.g. pending_human, blocked)
    --tag                Only show tasks carrying this tag
    --claimed-by         Only show tasks claimed by this actor
    --mine               Shortcut for --claimed-by <resolved actor>
    --json               Emit JSON output
```

## `tl show TASK_ID`

Show a task in detail. Human output includes the identifier, title, status,
priority, dependencies, claim state, and Markdown body content such as notes.
Accepts both full IDs (`task-k5g`) and bare short codes (`k5g`).

```
    --json               Emit JSON output
```

## `tl ready`

List tasks that are ready to be claimed. A task is ready when it is `open`
(or `in_progress` with an expired claim), all dependencies are `done`, and
no active claim exists.

Use `--tag` to narrow the queue to a role-ish dimension (review, docs,
arch, etc.). See [`../.decisions/0001-multi-agent-coordination-via-tags.md`](../.decisions/0001-multi-agent-coordination-via-tags.md)
for the rationale.

```
    --tag                Only show tasks carrying this tag
    --json               Emit JSON output
```

## `tl claim TASK_ID`

Claim a task with a time-limited lease. Sets status to `in_progress` and
records claim expiry. Rejects if another actor holds an active claim (exit
5) or dependencies are unmet (exit 4). Uses the same actor resolution chain
as `note`.

Re-running `tl claim` as the same actor extends the lease — this is the
heartbeat pattern for long-running work. Use `--force` to take over an
active claim held by a different actor.

```
    --actor              Claiming actor (optional; auto-resolved if unset)
    --ttl                Lease duration, e.g. 60m or 2h (default from config)
    --force              Take over an active claim held by a different actor
    --json               Emit JSON output
```

## `tl dep add TASK_ID --on TASK_ID`

Add a dependency link between tasks. Both tasks must exist (exit 3 if not).
Idempotent — adding the same dependency twice is a no-op.

```
    --on                 Target task to depend on (required)
```

## `tl dep remove TASK_ID --on TASK_ID`

Remove a dependency link. Both tasks must exist (exit 3 if not).
Idempotent — removing a non-existent dependency is a no-op.

```
    --on                 Target task to drop as a dependency (required)
```

## `tl note TASK_ID`

Append a timestamped note to a task's body under a `## Notes` section. Notes
are the human-facing audit trail — use them for progress updates, handoff
context, and decision records.

```
-m, --message            Note message (required)
    --actor              Actor writing the note (optional; auto-resolved)
```

## `tl close TASK_ID`

Mark a task as `done`. Unclaimed open tasks may be closed by any actor.
Claimed tasks may be closed by the claiming actor, or by another actor with
`--force`. Rejects blocked and already-done tasks.

```
    --actor              Actor closing the task (optional; auto-resolved)
    --force              Close even when another actor holds an active claim
    --json               Emit JSON output
```

## `tl cancel TASK_ID -m "<reason>"`

Mark a task as `cancelled`. Use when work will not be completed —
superseded, duplicated, no-longer-needed — so the audit trail records
intentional abandonment rather than falsely claiming completion. A reason
is required and stored as a note. Cancelling a claimed task releases the
claim; another actor's active claim requires `--force`. Rejects already
`done` and already `cancelled` tasks.

```
-m, --message            Cancellation reason (required, stored as a note)
    --actor              Actor cancelling the task (optional; auto-resolved)
    --force              Cancel even when another actor holds an active claim
    --json               Emit JSON output
```

## `tl release TASK_ID`

Voluntarily release a claim on a task, returning it to `open`. Only the
claiming actor may release unless `--force` is used.

```
    --actor              Actor releasing the claim (optional; auto-resolved)
    --force              Release even when another actor holds the claim
    --json               Emit JSON output
```

## `tl block TASK_ID -m "<blocker>"`

Mark a task `blocked` and record the blocking condition as a note. Use for
external blockers (waiting on upstream, infra down, third-party fix) —
distinct from `pending_human`, which is "I need an answer". Blocking a
claimed task releases the claim so others can see it is not actively being
worked.

```
-m, --message            Blocker description (required, stored as a note)
    --actor              Actor reporting the blocker (optional; auto-resolved)
    --json               Emit JSON output
```

## `tl unblock TASK_ID`

Clear the `blocked` status and return the task to `open` so it becomes
eligible for the ready queue again. Rejects tasks that are not blocked.

```
    --actor              Actor clearing the blocker (optional; auto-resolved)
    --json               Emit JSON output
```

## `tl history [TASK_ID]`

Print events recorded in `.taskledger/events.jsonl`. With a `TASK_ID`,
filters to that task's audit trail. With `--since <duration>`, filters to
events within the given window (e.g. `24h`, `7d`) across all tasks —
useful for "what did the team do today?" review. At least one of
`TASK_ID` or `--since` is required.

```
    --since              Only show events within this duration (e.g. 24h, 7d)
    --json               Emit JSON output (array of raw event objects)
```

## `tl stale`

List tasks whose claims have expired. Use before `tl release --force` to
clean up abandoned claims.

```
    --json               Emit JSON output
```

## `tl agents`

Print a recommended `AGENTS.md` snippet for TaskLedger-aware agents. Writes
only to stdout and never edits `AGENTS.md` for you. Commands in the snippet
are formatted as Markdown code spans, for example `tl ready --json`.

```
(no flags)
```
---

## Setup errors

Commands that need ledger state are non-interactive for agent safety. If
`.taskledger/` is missing, they exit with code `1` and print:

```text
TaskLedger is not initialized in this repository.
Run `tl init` from the repository root to create .taskledger/.
```
