# Project entry points

- `README.md` — installation, quickstart, and CLI flag reference.
- `docs/PRD.md` — design intent, non-goals, status enum, exit codes.
- `features/` — Gherkin behavioral spec, one `.feature` file per command.

The `@implemented` tag on each feature marks which commands are actually
built; the godog suite under `bdd/` scopes to those. Untagged features still
serve as the binding contract for unimplemented commands.

## Workflow

- `make bdd` runs the godog suite; `make test` runs everything.
- Adding a command: extend its `.feature`, add step defs in `bdd/bdd_test.go`,
  implement the cobra command under `cmd/<name>.go` (plus any needed package
  under `internal/`), then tag the feature `@implemented`.
- Mutating commands append a JSON line to `.taskledger/events.jsonl`. Read
  commands must support `--json`.

## Gherkin / BDD tests

When writing or changing `.feature` files, follow `docs/gherkin-guidelines.md`.

Rules:
- Write behavior, not implementation.
- Use one behavior per scenario.
- Keep scenarios independent.
- Use concrete examples.
- Do not write vague `Then` steps.
- Do not include step definitions in feature files unless explicitly requested.

## tl tool Workflow

This repository uses TaskLedger (`tl`) for local task coordination between humans and agents.

Set `TL_ACTOR` once at the start of your session so you don't need `--actor` on each command:

```sh
#when TL_ACTOR, AGENT_NAME are not set, try to set it.
export TL_ACTOR=<agent-name>:<purpose>
```

When starting work:

1. Pick a task:
   - `tl ready --json` for unclaimed work, or `tl ready --tag <role> --json` to filter by role-ish tags.
   - `tl show <task-id>` when handed a specific task.
   - `tl history <task-id>` if the task was previously worked on; read prior notes before starting.
2. Claim it before editing files:
   `tl claim <task-id>`
3. Inspect the task details:
   `tl show <task-id>`
4. Do the work. Re-run `tl claim <task-id>` periodically on long work — it extends the lease (heartbeat pattern).
   Record important context, decisions, blockers, or handoff notes:
   `tl note <task-id> -m "..."`
5. Pick the correct exit:
   - `tl close <task-id>` — work is done and verified.
   - `tl cancel <task-id> -m "<reason>"` — work won't be done.
   - `tl block <task-id> -m "<blocker>"` — external blocker; claim is released.
   - `tl pending <task-id> --question "..."` — you need a human decision; claim is released.
   - `tl release <task-id>` — you're stepping away cleanly; leave a comprehensive note first.

   (`cancel`, `block`, `pending` are spec'd in `features/`; check the current `@implemented` set with `make bdd` before relying on them.)

Rules:

- Do **not** work on a task claimed by another active actor unless explicitly told.
- If your work uncovers a separable piece of work, create a follow-up task with `tl create` rather than silently expanding scope.
- Prefer tasks from `tl ready`; blocked, pending, done, cancelled, or actively claimed tasks are not ready.
- Leave notes for partial progress, failed approaches, decisions, and handoffs.
- Do **not** edit `.taskledger/events.jsonl` manually.
- Create new task/story/bug when new work is defined. Always check whether its already covered by existing tasks.
- If `.taskledger/` is missing, ask the human whether to run `tl init`.


## Implementation notes

- **Major libs:** `spf13/cobra` (CLI), `gopkg.in/yaml.v3` (frontmatter),
  `cucumber/godog` (BDD acceptance tests).
- **ID generation:** `task-<3 lowercase alphanumeric>`, generated with
  `crypto/rand` and a collision-retry loop. Namespace ≈ 47k; well above the
  realistic ceiling for the project sizes TaskLedger targets.
- **Atomic writes:** task files write to `<id>.md.tmp` and `rename` over the
  target.
- **Locking:** an advisory `flock(2)` on `.taskledger/.lock` (via
  `github.com/gofrs/flock`) guards mutating commands. Acquired once at
  command start, held across the read-modify-write, released on exit (or
  via deferred unlock). Lock contention surfaces as exit code 7 after a
  5-second wait. Read commands need no lock — task files use `.tmp` +
  atomic `rename`, and `events.jsonl` uses `O_APPEND`.
- **Repository detection:** commands walk upward from CWD to find
  `.taskledger/`.

---