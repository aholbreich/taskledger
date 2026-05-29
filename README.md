# tl tool - Task ledger for your repository

> A Git-native task ledger for humans and AI coding agents.

Task ledger (`tl`) stores tasks as Markdown files with YAML frontmatter inside your repository, gives agents a dependency-aware ready queue, supports safe claim leases with automatic actor resolution, and records every change in an append-only event journal. The Handover work between agents (humans) can be task/story centric by utilizing notes feature.

No daemon. No hidden database. No automatic push. No AGENTS.md magic.

**Contents:** [Installation Options](#installation-options) · [Quickstart](#quickstart) · [Commands](#commands) · [Implementation status](#implementation-status) · [Development](#development) · [Further reading](#further-reading)

---

## Installation Options


### Homebrew (macOS / Linux)

```sh
brew install aholbreich/tap/tl           # latest stable release
brew install --HEAD aholbreich/tap/tl    # or: build from current main
```

If you install multiple tools from the same tap, you can tap once:

```sh
brew tap aholbreich/tap
brew install tl
```

Prebuilt binaries are available for **macOS (Intel + Apple Silicon)** and **Linux (amd64 + arm64)**.

### RPM (Fedora / Red Hat)

Add the Holbreich RPM repository:

```sh
# Documentation: https://aholbreich.github.io/rpm-repo/#installation-fedora-centos-redhat
echo '[Holbreich]
name=Holbreich Repository
baseurl=https://aholbreich.github.io/rpm-repo/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/holbreich.repo
```

Install `tl`:

```sh
sudo dnf install tl
tl --version
```

If you run into issues with the RPM repository, see the
[rpm-repo project](https://github.com/aholbreich/rpm-repo).

### Install script (macOS / Linux)

```sh
curl -fsSL https://raw.githubusercontent.com/aholbreich/tl/main/install.sh | sh
```

Install a specific version or target directory:

```sh
curl -fsSL https://raw.githubusercontent.com/aholbreich/tl/main/install.sh | sh -s -- --version 0.4.4
curl -fsSL https://raw.githubusercontent.com/aholbreich/tl/main/install.sh | sh -s -- --bin-dir "$HOME/.local/bin"
```


### From source

```sh
git clone https://github.com/aholbreich/tl
cd tl
make install                # installs `tl` to $HOME/bin
```

Cross-platform release archives:

```sh
make dists                  # tl-linux-amd64.tar.gz, tl-darwin-arm64.tar.gz, …
```

---

## Quickstart

```sh
tl init                                                          # one-time per repo
tl create "Add login form validation"
tl create "Refactor auth errors" -t chore -p low --tag auth
tl list
tl show <id>                                                     # full id or bare short code
```

Agent workflow:

```sh
tl ready --json                                                  # what's available?
tl claim <id>                                                    # take a lease (actor auto-detected)
tl show <id>                                                     # read the details
tl note <id> -m "Initial implementation done."                   # record a handoff note
tl close <id>                                                    # mark as done
```

Actor identity resolves in order: `--actor` flag > `TL_ACTOR` env >
`ACTOR_NAME` env > `BEADS_ACTOR` env > agent auto-detection.

### Shell completion

`tl` ships completions for bash, zsh, fish, and PowerShell. Pressing TAB on
a `TASK_ID` argument suggests the actual task IDs from the current ledger.

```sh
# bash (current shell)
source <(tl completion bash)
# zsh
tl completion zsh > "${fpath[1]}/_tl"
# fish
tl completion fish > ~/.config/fish/completions/tl.fish
```

`tl completion --help` prints persistent-install instructions for each shell.

---

## Commands

- Flag reference: [`docs/COMMANDS.md`](docs/COMMANDS.md)
- Behavioral spec: [`features/`](features) (one `.feature` file per command)
- At the terminal: `tl <cmd> --help`

---

## Implementation status

Implemented commands carry the `@implemented` tag in their feature file.
`make bdd` runs only the implemented suite; untagged features are the
binding contract for unimplemented commands.

---

## Exit codes

`0` success · `1` generic · `2` invalid args · `3` task not found ·
`4` task not ready · `5` already claimed · `7` lock failed

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

---

## Further reading

- [`docs/COMMANDS.md`](docs/COMMANDS.md) - per-command flag reference
- [`docs/tech-docs.md`](docs/tech-docs.md) - some implementation detail
- [`docs/PRD.md`](docs/PRD.md) — design intent, non-goals, status enum


