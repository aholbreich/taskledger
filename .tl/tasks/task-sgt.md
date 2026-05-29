---
id: task-sgt
title: 'Modernize CI: bump actions to Node24-compatible versions'
status: done
priority: medium
type: chore
created_at: 2026-05-29T16:52:29Z
updated_at: 2026-05-29T17:00:07Z
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
---

## Description

CI emits a deprecation warning: actions/checkout@v4 and actions/setup-go@v5 run on Node20, forced to Node24 from June 2 2026. Bump to current major versions and confirm green.

## Notes

- 2026-05-29T17:00:07Z [claude-code:promotion] note: Bumped actions/checkout@v4->v6 and actions/setup-go@v5->v6 in ci.yaml and release.yaml (both Node24 runtimes; latest majors confirmed via GitHub API). Resolves the Node20 deprecation warning. upload/download-artifact@v4 and softprops@v2 left as-is — still current majors, no Node24 successor yet. YAML validated. Note: release.back is a stale tracked backup with old pins but is not a .yml/.yaml so it's inert.
