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

# TaskLedger Workflow

This repository uses TaskLedger (`tl`) for local task coordination between humans and agents.

When starting work:

1. - Run `tl ready --json` to find tasks that are open, unblocked, and unclaimed.
   - Or `tl show <task-id>` when asked to do a particular task.
2. Claim one task before editing files:
   `tl claim <task-id> --actor <your-agent-name>`
3. Inspect the task details:
   `tl show <task-id>`
4. Do the work. Record important context, decisions, blockers, or handoff notes:
   `tl note <task-id> --actor <your-agent-name> -m "..."`
5. When the task is complete, close it:
   `tl close <task-id> --actor <your-agent-name>`


Rules:

- Do **not** work on a task claimed by another active actor unless explicitly told.
- Prefer tasks from `tl ready`; blocked, pending, done, cancelled, or actively claimed tasks are not ready.
- Leave notes for partial progress, failed approaches, decisions, and handoffs.
- Do **not** edit `.taskledger/events.jsonl` manually.
- Set `TL_ACTOR` when possible so commands can resolve your identity consistently.
- Ask before editing `AGENTS.md` or other project instruction files.
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