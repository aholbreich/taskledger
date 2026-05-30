---
id: task-xb3
title: Rename tl agents --update flag to --write-files (alias old name for compat)
status: open
priority: medium
type: chore
created_at: 2026-05-30T18:24:15Z
updated_at: 2026-05-30T18:24:15Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - agents
references:
  - cmd/agents.go
  - cmd/agents_test.go
  - features/agents.feature
---

## Description

The --update flag on tl agents is misleading — the command doesn't update tl itself, it writes/refreshes agent instruction files. Rename to --write-files, which is self-documenting.

Changes needed:
1. In cmd/agents.go: rename the flag variable from 'update' to 'writeFiles', change flag name from 'update' to 'write-files', add 'update' as a hidden alias for backward compatibility: c.Flags().BoolVar(&writeFiles, 'write-files', false, 'Write or refresh the tl workflow block in existing agent instruction files') + c.Flags().BoolVar(&writeFiles, 'update', false, '(deprecated: use --write-files)') with MarkHidden on 'update'.
2. In cmd/agents_test.go: update test names/comments that reference 'update'.
3. In features/agents.feature: change '--update' references to '--write-files' (keep one scenario testing the old --update alias for backward compat).
4. Update docs/usage.md 'tl agents' line in the Commands section if it references --update.
5. Update README.md if it shows --update anywhere.
6. Update AGENTS.md if it has inline tl agents --update mentions.

The old --update flag must continue to work (via hidden alias) — this is not a breaking change.
