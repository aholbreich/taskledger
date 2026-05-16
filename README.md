# TaskLedger

> A Git-native task ledger for humans and AI coding agents.

TaskLedger (`tl`) stores tasks as Markdown files with YAML frontmatter inside
your repository. The MVP is in progress: the implemented slice covers
initialization, task creation, task listing, task detail display, JSON output
for implemented read commands, creation events, and friendly setup errors.

The broader product goal is dependency-aware ready queues, safe claim leases,
stale claim detection, verification commands, and handoff notes.

No daemon. No hidden database. No automatic push. No AGENTS.md magic.

---

## Install

From source (Go 1.25+):

```sh
git clone https://github.com/aholbreich/taskledger
cd taskledger
make install                # installs `tl` to $HOME/bin
```

Cross-platform release archives:

```sh
make dists                  # produces tl-linux-amd64.tar.gz, tl-darwin-arm64.tar.gz, …
```

---

## Quickstart

```sh
tl init                                                          # one-time per repo
tl create "Add login form validation"
tl create "Refactor auth errors" -t chore -p low --tag auth
tl list
tl show task-abc
```

Planned agent workflow:

```sh
tl ready --json                                                  # what's available?
tl claim task-abc --actor claude-code:main                       # take a lease
tl show task-abc                                                 # read the details
tl note task-abc --actor claude-code:main --message "Initial implementation done."
tl verify task-abc                                               # run the task's checks
tl close task-abc --actor claude-code:main
```

Only `init`, `create`, `list`, and `show` are implemented today. The rest of
the planned workflow is specified under [`features/`](features/).

---

## Commands

Full behavioral spec lives under [`features/`](features). Flags currently
exposed by the implemented commands:

### `tl init`

Initialize a `.taskledger/` ledger in the current directory.

```
(no flags)
```

### `tl create [title] [options]`

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

### `tl list`

List every task in the ledger, sorted by priority then identifier.

```
    --json               Emit JSON output
```

### `tl show TASK_ID`

Show a task in detail. Human output includes the identifier, title, status,
priority, dependencies, claim state, and Markdown body content such as notes.

```
    --json               Emit JSON output
```

### Not yet implemented

`ready`, `dep add`, `dep remove`, `claim`, `release`, `stale`, `note`,
`verify`, `close`, `pending`, `resolve`, `prime` — specified in
[`features/`](features), implementation in progress. See
[`docs/PRD.md`](docs/PRD.md) section 7 for the MVP command list.

### Setup errors

Commands that need ledger state are non-interactive for agent safety. If
`.taskledger/` is missing, they exit with code `1` and print:

```text
TaskLedger is not initialized in this repository.
Run `tl init` from the repository root to create .taskledger/.
```

---

## Implementation Status

The BDD suite runs features tagged `@implemented`.

| Area | Status |
|---|---|
| `tl init` | Implemented |
| `tl create` | Implemented |
| `tl list` | Implemented |
| `tl show` | Implemented |
| Friendly missing-ledger hint | Implemented |
| `tl ready` | Specified, pending |
| `tl dep add` / `tl dep remove` | Specified, pending |
| `tl claim` / `tl release` / `tl stale` | Specified, pending |
| `tl note` | Specified, pending |
| `tl verify` / `tl close` | Specified, pending |
| `tl pending` / `tl resolve` | Specified, pending |
| `tl prime` | Specified, pending |

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
verify:
  commands: []
  evidence_required: []
tags: []
---

## Description

Validate email format and require a password.
```

---

## Exit codes

`0` success · `1` generic · `2` invalid args · `3` task not found ·
`4` task not ready · `5` already claimed · `6` verify failed · `7` lock failed

---

## Development

```sh
make build                  # version-stamped local binary
make test                   # all Go tests
make bdd                    # godog suite only
make dists                  # cross-platform release archives
make clean
```

CI runs `gofmt`, `go vet`, `make build`, `make test` on every PR and push to
`main` (see [`.github/workflows/ci.yaml`](.github/workflows/ci.yaml)).
Tag-triggered releases build all platforms and publish a GitHub Release.

The BDD suite lives in [`bdd/`](bdd/) and runs the features tagged
`@implemented`; the rest are pending-implementation specs.

---

## Further reading

- [`docs/PRD.md`](docs/PRD.md) — design intent, non-goals, status enum, exit codes
- [`features/`](features/) — Gherkin behavioral spec, one file per command
- [`AGENTS.md`](AGENTS.md) — leading doc for any agent working in this repo
- [`docs/gherkin-guidelines.md`](docs/gherkin-guidelines.md) — Gherkin style rules
