---
id: task-2q0
title: Add tl agents --file flag for targeted updates (scope control)
status: open
priority: low
type: feature
created_at: 2026-05-30T18:24:25Z
updated_at: 2026-05-30T18:24:25Z
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

Add a --file flag to 'tl agents --write-files' that limits which agent instruction files get updated, instead of always scanning the hardcoded list.

Behavior:
- --file can be repeated: 'tl agents --write-files --file CLAUDE.md --file .cursorrules'
- If --file is given, only those files are considered for update. Files not in the list are ignored.
- If --file is not given, the current default list applies (backward compatible).
- The default list should be updated to include: AGENTS.md, CLAUDE.md, GEMINI_RULES.md (note: renamed from GEMINI.md), .cursorrules, .aider-rules.md, .github/copilot-instructions.md.
- If --file points to a file that doesn't exist, skip it with a note (same as current behavior), don't create it.
- The hardcoded agentInstructionFiles variable in cmd/agents.go becomes the default list; --file overrides it.
- No error if --file names a nonexistent file — just output 'Skipped <file> (not found)' and continue.
- Add feature scenarios: --file single target, --file multiple targets (repeatable), --file with nonexistent file, --file with --dry-run.
