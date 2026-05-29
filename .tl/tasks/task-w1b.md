---
id: task-w1b
title: 'Fix red CI: fish completion install test fails on runner'
status: done
priority: high
type: bug
created_at: 2026-05-29T16:52:29Z
updated_at: 2026-05-29T16:58:38Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - ci
  - promotion
---

## Description

The CI run on main is failing on TestFeatures/tl_completion_--install_fish_writes_the_script_to_the_canonical_fish_path. The test passes locally but fails on the GitHub runner, indicating a test-isolation/env bug (HOME or XDG_* differs on the runner). Make the test set/override HOME and XDG_CONFIG_HOME deterministically so the canonical fish path resolves identically everywhere. Goal: green CI badge. This is the #1 credibility blocker for promotion.

## Notes

- 2026-05-29T16:58:38Z [claude-code:promotion] note: Root cause: bdd Before hook isolated CWD+HOME but not XDG_CONFIG_HOME/XDG_DATA_HOME/ZDOTDIR. installPath() consults those before $HOME, so the GitHub runner's XDG_CONFIG_HOME sent the fish script outside the temp HOME. Reproduced locally with XDG_CONFIG_HOME set. Fix: clear those three vars in the bdd Before hook (bdd/bdd_test.go). Verified: full gofmt/vet/build/test gate green even with all three leaked.
