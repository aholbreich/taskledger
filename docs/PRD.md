# PRD: TaskLedger

**TaskLedger** (`tl`) is a Git-native task ledger for humans and AI coding
agents.

This document captures design intent — the parts of the spec that are not
derivable from code, features, or the README. For per-command behavior read
the [`features/`](../features) directory; for flags read
[`COMMANDS.md`](COMMANDS.md); for storage layout and exit codes read the
[README](../README.md); for the on-disk task schema read
[`internal/task/task.go`](../internal/task/task.go).

---

## 1. Problem

Coding agents lose context across sessions. They need a durable, repo-local
way to know:

- What work exists.
- What is blocked.
- What is ready.
- Who is already working on what.
- What was already attempted.
- What counts as done.
- How to safely hand work back to a human or another agent.

Existing options trade off poorly:

- Built-in agent task lists are session-local.
- Plain Markdown TODO files lack dependency and claim semantics.
- Issue trackers are external and not optimized for agent automation.
- Heavier agent-coordination systems introduce databases, daemons, sync
  complexity, or hidden state.

---

## 2. Target users

Primary: a developer or architect using Claude Code, Codex, Cursor, or
similar agents inside a Git repository.

Secondary: indie hackers, small product teams, technical founders,
engineering managers coordinating agent-assisted work, and developers who
want local-first task tracking without SaaS overhead.

---

## 3. Core product thesis

A good agent task system should be:

1. **Repo-local** — state lives in the Git repository.
2. **Human-readable** — a developer can edit tasks with a normal editor.
3. **Machine-readable** — every read command supports `--json`.
4. **Dependency-aware** — agents can ask "what is ready now?".
5. **Claim-safe** — claims have leases with stale-claim detection.
6. **Handoff-oriented** — notes preserve context between humans and agents.
7. **Small and predictable** — no daemon, no hidden database, no automatic
   remote push, no AGENTS.md magic.

---

## 4. Non-goals

TaskLedger will not initially support:

- Real-time collaboration.
- A web app or hosted backend.
- Complex role hierarchies.
- Multi-repository orchestration.
- Automatic Git pushing or merging.
- Long-running background workers.
- tmux/session management.
- A full Jira / Linear / GitHub Issues replacement.
- AI agent execution itself.

The tool tracks and coordinates work. It does not run the agent in v1.

---

## 5. Status model

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

## 6. Success metrics

MVP is successful if:

- A developer can initialize a repo and create tasks in under one minute.
- Claude Code, Codex, or similar agents can reliably query ready tasks via
  JSON.
- Two agents do not accidentally work on the same task when claims are
  respected.
- Stale claims can be detected and released.
- All task state is reviewable in Git.
- No hidden service or database is required.
- The project remains understandable after one hour of code reading.

---

## 7. Key differentiator

The strongest differentiator is not "file-based tasks". It is:

> **Agent-safe task coordination with readable Git-native state.**

That means:

- Claims are explicit.
- Stale work is detectable.
- Dependencies are computable.
- Handoffs are recorded.
- Humans can inspect everything.
