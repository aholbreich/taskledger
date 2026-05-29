# tl by example

This is a walkthrough, not a flag reference. Task IDs are random
(`task-<3 chars>`), so yours will differ.

A few facts that make the rest read more easily:

- State lives in `.tl/` — `config.yaml`, one Markdown file per task under `tasks/`, and an append-only `events.jsonl` log.
- Mutating commands (`create`, `claim`, `note`, `close`, …) append a line to `events.jsonl`. Read commands (`list`, `show`, `ready`, `history`) take `--json`.
- Identity resolves in order: `--actor` flag, then `TL_ACTOR`, `ACTOR_NAME`, then agent auto-detection. Set `TL_ACTOR` once per session and forget about it.

```sh
export TL_ACTOR=alex # or claude-code:auth, agent-a:backend, …
```

**Contents:** [Starting a ledger](#starting-a-ledger) ·
[The work loop](#the-work-loop) · [When work stalls](#when-work-stalls) ·
[Two agents, one ledger](#two-agents-one-ledger) ·
[Reading the ledger](#reading-the-ledger) · [References](#references) ·
[Setup and housekeeping](#setup-and-housekeeping)

---

## Starting a ledger

One `init` per repository. It creates `.tl/` and nothing else.

```sh
$ tl init
Initialized task ledger at /home/you/project/.tl
Tip: run `tl completion --install` to enable TAB completion for task IDs.
```

Create work. A bare title is enough; flags add structure.

```sh
$ tl create "Add login form validation"
Created task task-hfv

$ tl create "Refactor auth error handling" -t chore -p high \
    --tag auth --tag backend \
    -d "Collapse the duplicated error branches in the auth package into one typed error."
Created task task-9jg
```

`-t` is the type (free-form: `chore`, `bug`, `feature`, …), `-p` is priority (`low|medium|high`, or just `l|m|h`), `--tag` repeats, and `-d` is the
description stored under `## Description` in the task file.

`refine` edits an existing task — title, type, priority, description. Tags are
set at creation; `refine` doesn't touch them (at least yet).

```sh
$ tl refine task-hfv -p high -t "Add login form validation (email + password)"
Refined task task-hfv
```

Declare a dependency when one task can't start until another finishes. `dep add` is silent on success.

```sh
$ tl dep add task-9jg --on task-hfv

$ tl show task-9jg
ID: task-9jg
Title: Refactor auth error handling
Status: open
Priority: high
Depends On:
  - task-hfv
Claim: none

## Description

Collapse the duplicated error branches in the auth package into one typed error.
```

---

## The work loop

`ready` shows unclaimed, unblocked work. Note that `task-9jg` is missing - it depends on `task-hfv`, which isn't done, so it isn't ready yet.

```sh
$ tl ready
ID        Status  Priority  Title
task-hfv  open    high      Add login form validation (email + password)
```

Claim it before editing files. The claim is a lease with an expiry.

```sh
$ tl claim task-hfv
Claimed task task-hfv (alex, expires 2026-05-29T18:16:18Z)
```

Record what matters as you go: decisions, what's left, etc. `note` is silent on success; the note lands in the task file and the event log.

```sh
$ tl note task-hfv -m "Added client-side checks; server-side validation still TODO."
```

Close it when it's done and verified.

```sh
$ tl close task-hfv
Closed task task-hfv
```

Closing `task-hfv` clears the dependency, so `task-9jg` is ready now.

---

## When work stalls

Not every task ends in `close`. There are four other exits, each for a specific situation.

**Block** — an external dependency you can't resolve yourself. The claim is released so someone else isn't locked out.

```sh
$ tl block task-9jg -m "Waiting on the auth provider migration to land."
Blocked task-9jg

$ tl unblock task-9jg
Unblocked task-9jg
```

**Pending** - you need a human decision before continuing. The claim is released and the task moves to `pending_human`, so it drops out of `ready`
until answered.

```sh
$ tl pending task-9jg --question "Wrap legacy errors or replace them outright?"
Marked task-9jg pending_human

$ tl show task-9jg | head -4
ID: task-9jg
Title: Refactor auth error handling
Status: pending_human
Priority: high
```

A human answers with `resolve`, which records the answer and reopens the task.

```sh
$ tl resolve task-9jg --answer "Replace outright; legacy callers are all in this PR."
Resolved task-9jg
```

**Cancel** - the work won't be done. The reason is required and kept in the record.

```sh
$ tl cancel task-iqq -m "Out of scope for this release."
Cancelled task-iqq
```

**Release** - you're stepping away but the work is still valid. Leave a note
first; the next actor reads it before picking up. (Covered next.)

---

## Two agents, one ledger

This is what the claim model is for. Set a distinct `TL_ACTOR` or use `--actor` per agent - the `name:purpose` convention reads well in the log.

Agent A claims the task:

```sh
$ tl claim task-9jg --actor agent-a:backend
Claimed task task-9jg (agent-a:backend, expires 2026-05-29T18:16:18Z)
```

Agent B tries the same task and is turned away. The exit code is `5`, so a
script can branch on it.

```sh
$ tl claim task-9jg --actor agent-b:backend
Error: task task-9jg is already claimed by agent-a:backend
$ echo $?
5
```

A claim is a lease, not a lock - it expires. On long work, re-claim to extend it. This is the heartbeat: same actor, fresh expiry.

```sh
$ tl claim task-9jg --actor agent-a:backend
Claimed task task-9jg (agent-a:backend, expires 2026-05-29T18:16:31Z)
```

To hand off cleanly, leave a note that says where to pick up, then release. After `release`, the task is claimable by anyone again.

```sh
$ tl note task-9jg \
    -m "Typed AuthError in place; wiring callers next. Pick up at internal/auth/errors.go." --actor agent-a:backend
$ tl release task-9jg --actor agent-a:backend
Released claim on task-9jg
```

If an actor disappears without releasing, the lease eventually expires and the
work surfaces as stale — see [Reading the ledger](#reading-the-ledger).

---

## Reading the ledger

`list` shows open work by default; closed and cancelled tasks are hidden unless
you ask.

```sh
$ tl list
ID        Status  Priority  Claimed By  Title
task-9jg  open    high      -           Refactor auth error handling
```

Useful filters: `--all` includes done and cancelled, `--status <status>` shows one status, `--tag <tag>` filters by tag, and `--mine` shows only what the
current actor holds.

For scripts and agents, `--json` is available on every read command. The shape is stable:

```sh
$ tl ready --json
[
  {
    "id": "task-hfv",
    "title": "Add login form validation (email + password)",
    "status": "open",
    "priority": "high",
    "created_at": "2026-05-29T17:15:41Z",
    "updated_at": "2026-05-29T17:16:06Z",
    "created_by": "human",
    "assignee": null,
    "depends_on": [],
    "claim": { "actor": null, "claimed_at": null, "expires_at": null, "heartbeat_at": null },
    "tags": []
  }
]
```

`history` replays every event for a task — the full audit trail, in order. This is the same data agents read before picking up handed-off work.

```sh
$ tl history task-9jg
2026-05-29T17:15:41Z  created           task-9jg  human
2026-05-29T17:16:06Z  dependency_added  task-9jg  -
2026-05-29T17:16:18Z  claimed           task-9jg  agent-a:backend
2026-05-29T17:16:31Z  claim_renewed     task-9jg  agent-a:backend
2026-05-29T17:16:31Z  note_added        task-9jg  agent-a:backend
2026-05-29T17:16:31Z  released          task-9jg  agent-a:backend
```

`stale` lists claims whose lease has expired — work an actor took but never
finished or released. It prints nothing when every claim is fresh.

```sh
$ tl stale
```

`agents` prints a ready-to-paste workflow guide for a coding agent. Drop its
output into `AGENTS.md` or an agent's context so it knows the claim/note/close
discipline without being told each time.

```sh
$ tl agents | head -3
## tl workflow

This repository uses `tl` for local task coordination between humans and agents.
```

---

## References
References are a great opportunity to maintain context and keep it close to the task.
A reference can point to anything: a file path, a URL, a ticket ID, an ADR, a Gherkin feature — anything that helps a human or an LLM find the
context behind the work. References are plain strings; `tl` stores them verbatim and does not validate them at input time.

Attach them at creation with `--ref` (repeatable, like `--tag`):

```sh
$ tl create "Add login" --ref src/auth/login.go --ref https://github.com/aholbreich/tl/pull/42 --ref JIRA-1234
Created task task-gn0
```

`show` lists them between Depends On and Claim:

```sh
$ tl show task-gn0
ID: task-gn0
Title: Add login
Status: open
Priority: medium
Depends On: none
References:
  - src/auth/login.go
  - https://github.com/aholbreich/tl/pull/42
  - JIRA-1234
Claim: none
```

Add and remove later with `refine`. Both are idempotent: adding a reference that already exists, or removing one that isn't there, changes nothing and
records no event.

```sh
$ tl refine task-gn0 --add-ref features/login.feature --remove-ref JIRA-1234
Refined task task-gn0
```

Each add or remove writes a `reference_added` / `reference_removed` event carrying the reference as its value, so the history shows what was attached and
when. In `--json`, references are always an array — `[]` when there are none.

---

## Setup and housekeeping

**Shell completion.** One step, auto-detecting your shell from `$SHELL`. Pressing TAB on a task-ID argument completes against the real IDs in the
current ledger.

```sh
$ tl completion --install            # or: tl completion --install bash
```

**Actor identity** resolves in this order, first match wins:

1. `--actor` flag
2. `TL_ACTOR`
3. `ACTOR_NAME`
4. agent auto-detection

**Exit codes** — for scripting:

| Code | Meaning            |
| ---- | ------------------ |
| `0`  | success            |
| `1`  | generic error      |
| `2`  | invalid arguments  |
| `3`  | task not found     |
| `4`  | task not ready     |
| `5`  | already claimed    |
| `7`  | lock failed        |

**Locking.** Mutating commands take an advisory `flock` on `.tl/.lock` for the duration of one read-modify-write, then release it. If another process holds it, the command waits up to five seconds and then exits `7`. Read commands take no lock — task files are written atomically and `events.jsonl` is append-only.

**Don't hand-edit `.tl/events.jsonl`.** It's the append-only source of truth; let the commands write it.
