---
id: task-cys
title: Add tl agents --remove flag to strip managed blocks from agent files
status: open
priority: low
type: feature
created_at: 2026-05-30T18:24:31Z
updated_at: 2026-05-30T18:24:31Z
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
  - features/agents.feature
---

## Description

Add a --remove flag to 'tl agents' that strips the managed tl workflow block (<!-- BEGIN TL WORKFLOW --> ... <!-- END TL WORKFLOW -->) from agent instruction files.

Behavior:
- 'tl agents --remove' scans the known agent instruction files and removes the managed block from each file that contains one.
- Also handles the legacy block markers (<!-- BEGIN TASKLEDGER WORKFLOW --> ...).
- Output format: one line per file modified, e.g. 'Removed tl workflow block from AGENTS.md'.
- If no files contain a managed block, output 'No managed tl workflow blocks found'.
- --remove can be combined with --file to target specific files: 'tl agents --remove --file CLAUDE.md'.
- --remove combined with --dry-run shows what would be removed without modifying files: 'Would remove tl workflow block from CLAUDE.md'.
- --remove and --write-files are mutually exclusive (error if both passed).
- Exit code 0 even when nothing is removed (diagnostic only).
- After removal, the surrounding whitespace should be cleaned up (no trailing blank lines left by the removed block).

Implementation notes:
- Implement removeAgentBlocks() function in cmd/agents.go.
- Refactor mergeAgentsBlock() into a shared helper — the pattern of scanning files and applying an operation is shared by --write-files, --remove, and --dry-run.
- Add feature scenarios in features/agents.feature: remove from one file, remove from multiple, remove with no blocks found, remove with --dry-run, remove + --write-files mutual exclusion, remove with --file targeting.
