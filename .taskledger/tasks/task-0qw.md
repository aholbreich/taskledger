---
id: task-0qw
title: make repository push work like it works with adr-tool
status: open
priority: medium
created_at: 2026-05-17T20:49:54Z
updated_at: 2026-05-22T21:04:26Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags: []
---

## Description

The README's "RPM (Fedora / Red Hat)" section tells users to add the Holbreich
RPM repo and run `sudo dnf install taskledger`. This currently does **not** work:
no CI job builds an `.rpm` or publishes one to https://aholbreich.github.io/rpm-repo/,
so the package is never actually available via dnf. The documented install is a
promise the project doesn't yet keep.

**Goal:** make `dnf install taskledger` work as the README describes, by porting
the release/publish pipeline from the sibling project `aholbreich/adr-tool` (this
repo's Makefile is already "adapted from adr-tool").

**Deliverable:**
- On a release tag, build the `.rpm` and publish it to the `aholbreich/rpm-repo`
  GitHub Pages dnf repository, using the same mechanism adr-tool uses.
- Copy/adapt the relevant GitHub Actions workflow(s) from adr-tool into
  `.github/workflows/`.
- Verify end-to-end: a clean Fedora host can add the repo and successfully run
  `sudo dnf install taskledger`.

**References:**
- adr-tool: https://github.com/aholbreich/adr-tool (source of the working pipeline)
- rpm-repo: https://github.com/aholbreich/rpm-repo (the published dnf repo)

**Heads-up (separable):** there are two conflicting release workflows —
`.github/workflows/release.yaml` (fires on any tag, currently works) and
`release.yml` (fires only on `v*` tags, never runs because tags are unprefixed,
e.g. `0.4.1`). Worth reconciling, but track separately from the RPM work.
