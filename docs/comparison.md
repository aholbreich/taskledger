# How tl compares

`tl` sits in a small but real category: Git-native task trackers built for
humans **and** AI coding agents. Its nearest neighbours are
[Beads](https://github.com/steveyegge/beads) and
[Backlog.md](https://github.com/MrLesk/Backlog.md). The honest shape of the
trade-offs (these tools move fast — verify before quoting):

|                                   | **tl**                                | **Beads**                    |   **Backlog.md**          | **GitHub Issues**   |
| --------------------------------- | ------------------------------------- | ---------------------------  | ----------------------- | ------------------- |
| State lives in                    | Markdown + append-only `events.jsonl` | Embedded Dolt DB (`.beads/`) | Markdown + YAML         | A remote service    |
| Read / edit a task in your editor | ✅                                    | ❌ (binary DB)               | ✅                      | Via the web UI      |
| Inspect history with `git diff`   | ✅                                    | ～ (DB diffs)                 | ✅                      | ❌                  |
| Dependency-aware `ready`          | ✅                                    | ✅                           | ～                      | ～                  |
| Coordination primitive            | Leases + stale-work detection         | `ready` / gates / routing    | Status + Kanban board   | Assignees           |
| Sync model                        | Plain `git` — you stay in control     | `bd dolt push` / `pull`      | `git`                   | Always online       |
| Extra surface area                | None — one static binary              | Embedded database            | React web UI + MCP server | SaaS + API        |
| Works offline / at a commit       | ✅                                    | ✅                           | ✅                      | ❌                  |

**Why `tl` and not Beads?** Beads keeps its ledger in an embedded Dolt
database under `.beads/` and syncs it with `bd dolt push/pull`. That buys a
genuine dependency graph, "memory decay", and cross-machine sync — at the price
of a binary store you commit to git but cannot read or `diff` as text. `tl`
makes the opposite bet: every task is a Markdown file you can read, `grep`,
edit, and `git diff` with zero tooling, and the only moving parts are those
files plus an append-only event log. Want a database that remembers things for
your agent? Use Beads. Want a ledger you can read at any commit and reason about
yourself? Use `tl`.

**Why `tl` and not Backlog.md?** Backlog.md shares the plain-Markdown,
Git-native philosophy but centers on a Kanban board and ships a web UI and an
MCP server. `tl` is deliberately leaner: no board, no server, no daemon. Its
core primitive is *coordination* — explicit claims with leases, detectable
stale work, and a recorded handoff trail — not visualization.

The shared backbone, in `tl`'s words:

> **Agent-safe task coordination with readable, Git-native state.**
> Claims are explicit · stale work is detectable · dependencies are computable
> · handoffs are recorded · humans can inspect everything.

**What `tl` deliberately leaves out** (see the
[non-goals](PRD.md)): no web UI or Kanban board, no embedded database or
automatic sync, no hosted backend, and it does not run the agent itself. `tl`
tracks and coordinates the work; Git and your agent do the rest.
