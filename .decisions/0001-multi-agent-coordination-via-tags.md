# 0001. Multi-agent coordination via tags

**Status:** Accepted (2026-05-17)

## Context

A product owner using TaskLedger may want to plan multi-agent work in
advance — for example, "implement", "code review", and "documentation"
passes for a single feature, each picked up by a differently-configured
agent. The straightforward way to express this is a dedicated `role` field
on tasks (or on actors) and gate agent visibility on role match. A related
move is to introduce a `type: story` with parent/child task relationships
to capture the planning structure.

Both moves push TaskLedger toward Jira-style hierarchy and role
enforcement, which the PRD explicitly disclaims under "Complex role
hierarchies" and "not a Jira / Linear / GitHub Issues replacement" (PRD
§4 Non-goals).

## Decision

Encode role-ish dimensions as **tags on tasks**, not as a dedicated `role`
schema field or a separate task type.

- A PO tags tasks at create time: `--tag review`, `--tag docs`,
  `--tag arch`.
- Agents discover their work via tag-filtered read commands:
  - `tl ready --tag review`
  - `tl list --tag review`
- "Story / umbrella" tasks are a **convention**, not a first-class type:
  create a normal task whose `depends_on` lists its children. It becomes
  ready only after all children are `done`; closing it represents feature
  completion.
- Recent cross-task activity is surfaced through `tl history --since
  <duration>`, not through a hierarchy view.

## Consequences

**Accepted trade-offs:**

- No story rollup or status aggregation. A PO asking "is the login story
  done?" reads the umbrella task's status, which is gated on its
  dependencies.
- No role-based identity enforcement. Whatever an agent claims is what it
  picks up; the filter is convention, not gatekeeping.
- Tags must be agreed on at the project level. There is no taxonomy
  validation.

**Avoided costs:**

- No new schema fields (`role`, `parent`, `subtasks`).
- No new task types (`story`, `subtask`, `epic`).
- No identity semantics for "what role is this agent currently in" (per
  claim? per session? per actor name?).
- No move toward Jira-style hierarchy.

## Alternatives considered

1. **`type: story` with `parent` / `children` fields.** Rejected —
   duplicates what `depends_on` already does, and the hierarchy semantics
   it would introduce (rollup, child-status aggregation) are explicit
   non-goals.

2. **`role` field on tasks plus agent self-declared mode.** Rejected —
   roles are brittle (real work crosses lines such as "review needs a
   small doc tweak"), and "agent mode" introduces ambiguous identity (per
   claim? per session? per actor name?).

3. **Status quo, no filtering.** Rejected — without any way to filter,
   agents in different roles compete for the same ready queue, defeating
   the point of planning passes.

## Related work in this session

- Added scenarios: `tl ready --tag X` in `features/ready.feature` and
  `tl list --tag X` in `features/list.feature`.
- Added scenarios: `tl history --since <duration>` for global recent
  activity in `features/history.feature`.
- Not added (deferred): `tl create --parent <id>` syntactic sugar for the
  umbrella pattern. Defer until empirical friction with `tl create + tl
  dep add` warrants it.

## References

- PRD `docs/PRD.md` §3 (Core thesis: small and predictable)
- PRD `docs/PRD.md` §4 (Non-goals: complex role hierarchies; not a Jira
  / Linear replacement)
