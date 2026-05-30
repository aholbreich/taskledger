---
id: task-hz9
title: Add make changelog + make release automation (replace backfill)
status: open
priority: medium
type: feature
created_at: 2026-05-29T16:52:45Z
updated_at: 2026-05-30T19:32:08Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - docs
  - automation
  - make
  - ci
---

references:
  - Makefile
  - .github/workflows/ci.yaml

## Description

**Scope change** — after analysis, backfilling a CHANGELOG.md from 12 pre-1.0
tags (0.1.0→0.7.0) across 6 weeks of rapid iteration provides minimal value.
The release notes would be generic, and the audience is early adopters who
found tl via the demo/README.

Instead, automate forward: add `make changelog` and `make release` so every
future release gets proper categorized notes with zero manual effort.

### Background

- tl has 0 merged PRs — commits go directly to main
- `gh release create --generate-notes` is PR-based and produces only a
  compare link when no PRs exist
- `make bump` is broken (interactive `read` in Makefile, tags only, no
  GitHub Release, no notes)
- Past releases (0.4.0–0.7.0) have empty release bodies like
  `"tl 0.7.0 - multi-platform binaries."`

### Deliverables

#### 1. `make changelog` — generate release notes from git log

A new Makefile target that:
- Reads `git log --oneline` between the last tag and HEAD
- Groups commits by conventional commit prefix (`feat:`, `fix:`, `docs:`,
  `refactor:`, `chore:`/`chores:`) with emoji headers
- Outputs formatted markdown to stdout
- Handles the edge case of no previous tag gracefully

Example output:
```
## What's Changed

### 🚀 Features
  • feat: add references field to task frontmatter
  • feat: add animated terminal demo SVG to README

### 📖 Documentation
  • docs: add status badges to README header

### 🔧 Maintenance
  • chore: ledger state for promotion tasks
```

#### 2. `make release VERSION=x.y.z` — one-command release

A new Makefile target that:
- Requires `VERSION` argument (fails with usage error if missing)
- Runs `make dists` first (build cross-platform archives)
- Runs `make changelog` to capture notes
- Tags HEAD with `$(VERSION)`
- Pushes the tag
- Creates a GitHub Release with proper notes and attaches built archives
- Cleans up temp files

```makefile
release: dists
	@prev=$$(git describe --tags --abbrev=0); \
	notes=$$(mktemp); \
	$(MAKE) changelog > $$notes; \
	git tag $(VERSION); \
	git push origin $(VERSION); \
	gh release create $(VERSION) --notes-file $$notes --title "tl $(VERSION)" tl-*.tar.gz tl-*.zip; \
	rm $$notes
```

#### 3. Remove `make bump` (broken, replaced)

Remove the old interactive `bump` target from the Makefile.

#### 4. Minimal CHANGELOG.md (optional)

If a file in-repo is wanted, create a minimal one that simply points to
GitHub Releases — no backfill:

```markdown
# Changelog

All notable changes are documented in the
[GitHub Releases](https://github.com/aholbreich/tl/releases) page.

Releases follow [Semantic Versioning](https://semver.org/).
```

### Not in scope

- `.github/release.yml` — useless since tl has no PRs
- Backfilling past releases — skipped by design
- NPM/Gems/PyPI publishing — not relevant for a Go binary

## Notes

- 2026-05-30T19:32:08Z [pi:planning] note: Task scope fully revised after analysis. Key finding: tl has 0 merged PRs (direct-to-main workflow), so GitHub's PR-based --generate-notes produces only a compare link. Replaced backfill scope with: (1) make changelog — git log grouped by conventional commit prefix, (2) make release VERSION=x.y.z — one-command release pipeline, (3) remove broken make bump, (4) optional minimal CHANGELOG.md pointing to releases page.
