# Contributing to `tl`

Thanks for your interest in `tl` — the Git-native task ledger for humans and AI coding agents.

This project is solo-maintained atm. Issues and pull requests are reviewed on a best-effort basis. 
Small, focused contributions land fastest; large or speculative changes are best discussed in an issue first.

By contributing you agree your work is released under the project's [MIT License](LICENSE).

---

## Ways to contribute

- **Report a bug** - open a GitHub issue.
- **Suggest a feature or design change** - open a GitHub issue or discussion;
  for anything non-trivial, please outline the user-facing behavior before writing code.
- **Improve the docs** — `README.md`, `docs/`, or the per-command `features/*.feature` specs.
- **Send a pull request** — see workflow below.

## Filing a good issue

Please include:

- `tl --version`
- OS and architecture (e.g. `Linux x86_64`, `macOS arm64`)
- Minimal steps to reproduce
- What you expected vs. what happened (paste exact output where useful)

For crashes or lock-contention reports, the contents of `.tl/events.jsonl`
around the failing command are usually decisive.

---

## Development setup

Prerequisites:

- **Go 1.25** (matches CI; see [`.github/workflows/ci.yaml`](.github/workflows/ci.yaml))
- `make`

Common targets (full list in the [`Makefile`](Makefile)):

```sh
make build        # version-stamped local binary
make test         # all Go tests (unit + BDD)
make bdd          # godog BDD suite only
make install      # install to $HOME/bin
```

CI runs `gofmt -l`, `go vet ./...`, `make build`, and `make test` on every PR. PRs whose CI is red won't be reviewed until it's green.

---

## The workflow rule (please read before opening a PR)

`tl` is **behavior-first**. Every command has a Gherkin feature file in [`features/`](features) that is the binding contract for what the command
does. The `@implemented` tag marks which features are wired up to the godog suite under [`bdd/`](bdd).

**Adding or changing a command:**

1. Edit (or create) the matching `features/<name>.feature` first. Describe
   the user-visible behavior, not the implementation. Follow
   [`docs/gherkin-guidelines.md`](docs/gherkin-guidelines.md).
2. Add or update step definitions in [`bdd/bdd_test.go`](bdd/bdd_test.go).
3. Implement the cobra command under `cmd/<name>.go`, plus any supporting
   package under `internal/`.
4. Once the scenarios pass, tag the feature `@implemented`.
5. Run `make test` and make sure it's green.

Bug fixes follow the same loop: add or extend a scenario that fails on `main`, then make it pass.

For design intent and explicit non-goals, see [`docs/PRD.md`](docs/PRD.md).

## Using the `tl` ledger to coordinate

This repo dogfoods its own tool. In-flight work lives in `.tl/`. If you'd
like to pick up an existing task:

```sh
tl ready --json        # what's available
tl claim <task-id>     # take a lease before editing
tl note <task-id> -m "…"   # leave handoff context
tl close <task-id>     # when done
```

The full agent/human workflow (claim leases, notes, blockers, handoffs) is documented in [`AGENTS.md`](AGENTS.md). 
Humans are welcome to follow the same flow; setting `TL_ACTOR=<your-name>:contrib` once per session avoids typing `--actor`.

If you uncover separable work, prefer creating a follow-up task with `tl create` over silently expanding tasks scope.

---

## Pull request guidelines

- Keep PRs **small and single-purpose**. One logical change per PR.
- Update or add tests for any behavioral change.
- Update `features/*.feature` and relevant `docs/` pages alongside code.
- Make sure `make test` passes locally.
- Rebase rather than merge when bringing in `main` updates.

## Questions

Open a GitHub Discussion or issue. For anything sensitive, contact the maintainer directly via the address in the commit log.
