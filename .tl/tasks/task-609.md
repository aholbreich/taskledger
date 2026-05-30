---
id: task-609
title: Add tl agents --dry-run flag for --write-files
status: open
priority: medium
type: feature
created_at: 2026-05-30T18:24:20Z
updated_at: 2026-05-30T18:24:20Z
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

Add a --dry-run flag to 'tl agents --write-files' that reports what would be changed without modifying any files.

Behavior:
- 'tl agents --write-files --dry-run' scans the known agent instruction files (AGENTS.md, CLAUDE.md, GEMINI_RULES.md, etc.) and prints which files exist and would be updated, and which don't exist and would be skipped.
- Output format: one line per file, e.g. 'Would update AGENTS.md (managed block found)' / 'Would update CLAUDE.md (no managed block yet, would append)' / 'Would skip GEMINI_RULES.md (file not found)'.
- If --dry-run is passed without --write-files, return an error: '--dry-run requires --write-files'.
- Exit code 0 even if no files would change (diagnostic, not failure).
- The dry-run scan should reuse the same file-detection logic as the real update path to avoid drift.

Implementation notes:
- Add a 'dryRun' bool flag alongside 'writeFiles' in newAgentsCmd().
- Extract file-scanning into a helper that returns a list of (path, action string) pairs.
- The existing updateAgentInstructionFiles() function calls that helper and then acts on it.
- Add feature scenarios in features/agents.feature for: dry-run with existing files, dry-run with no files, dry-run without --write-files (error).
