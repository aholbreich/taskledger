---
id: task-afx
title: Write docs/usage.md (by-example guide); delete COMMANDS.md
status: done
priority: medium
type: feature
created_at: 2026-05-29T17:15:29Z
updated_at: 2026-05-29T17:19:33Z
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
---

## Description

Replace the per-command COMMANDS.md dump with a narrative by-example guide at docs/usage.md, organized as flows: starting a ledger (genesis), the work loop, off-ramps, two-agents/multi-agent coordination, reading the ledger, setup & housekeeping. References flow stubbed with a TODO until task-reg ships. Delete docs/COMMANDS.md and fix the 3 inbound links (README x2, docs/tech-docs.md x1). Transcripts captured from real command runs. Voice: terse, human, no marketing filler.

## Notes

- 2026-05-29T17:19:33Z [claude-code:promotion] note: Wrote docs/usage.md (~1465 words): 7 flows — starting a ledger, the work loop, when work stalls (block/pending/cancel/release), two agents one ledger (claim contention exit 5, heartbeat re-claim, handoff via note+release), reading the ledger (list filters, --json, history audit trail, stale, agents), references (stubbed -> task-reg), setup & housekeeping (completion, actor order, exit-code table, locking). All transcripts captured from real runs in a scratch ledger, not invented. Deleted docs/COMMANDS.md; repointed 3 links (README x2, tech-docs x1) to usage.md. No dangling COMMANDS.md refs remain. Not committed (session: local-only, .tl excluded).
