---
id: task-v6f
title: Add Cursor rules directory support to tl agents --write-files
status: open
priority: low
type: feature
created_at: 2026-05-30T18:38:33Z
updated_at: 2026-05-30T18:38:33Z
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

Add support for Cursor's .cursor/rules/ directory format to tl agents --write-files (and --file).

Context: Cursor supports two formats — (1) a single .cursorrules file in the project root, and (2) a .cursor/rules/ directory with individual .mdc rule files (newer format). The .cursor/rules/ format is the recommended one as of Cursor 0.45+.

Behavior:
- 'tl agents --write-files' auto-detects the Cursor format to use:
  1. If .cursor/rules/ directory exists → write .cursor/rules/tl.mdc (the .mdc file format Cursor uses for individual rules).
  2. If .cursorrules file exists (and .cursor/rules/ doesn't) → merge into .cursorrules (same managed block approach as AGENTS.md).
  3. If neither exists → skip (same as current behavior for missing files — don't create).
- The .cursor/rules/tl.mdc file uses Cursor's .mdc format: a markdown file with optional YAML frontmatter for rule metadata (globs, description). The tl workflow content goes in the body.
- The default agentInstructionFiles list should include both '.cursorrules' and '.cursor/rules/tl.mdc' (detected at write time, not a static list entry — the directory check happens during scanning).
- The --file flag (task-2q0) can be used to manually target either: 'tl agents --write-files --file .cursor/rules/tl.mdc'.

Implementation notes:
- Add a helper detectCursorFormat(ledger string) that returns which Cursor format to use (enum: rulesDir, singleFile, none).
- In the file-scanning loop (shared by --write-files and --dry-run), insert the detected Cursor path as an additional candidate.
- The .mdc file does NOT need the BEGIN/END TL WORKFLOW markers — Cursor expects whole-file rules. Write the full snippet directly (like --output behavior).
- Feature scenarios: write to .cursor/rules/ (directory exists), write to .cursorrules (no rules dir), dry-run with Cursor formats, --file targeting .cursor/rules/tl.mdc explicitly.
