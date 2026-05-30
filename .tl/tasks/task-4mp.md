---
id: task-4mp
title: Add tl agents --output flag to write snippet to custom path
status: open
priority: low
type: feature
created_at: 2026-05-30T18:38:25Z
updated_at: 2026-05-30T18:38:25Z
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

Add a --output flag to 'tl agents' that writes the workflow snippet to a custom file path instead of stdout.

Behavior:
- 'tl agents --output docs/agent-guide.md' writes the full snippet (same as default stdout) to docs/agent-guide.md.
- 'tl agents --compact --output docs/agent-guide.md' writes the compact snippet.
- 'tl agents --output /path/to/file --write-files' is an error ('--output and --write-files are mutually exclusive').
- The output file is created if it doesn't exist, overwritten if it does (same as --write-files behavior).
- The output file does NOT get managed markers (BEGIN/END TL WORKFLOW) — it's a pure write, not a merge. If users want managed blocks, they use --write-files which targets known files with merge behavior.
- If --output points to a directory, return an error.

Implementation:
- Add 'output' string flag in newAgentsCmd().
- Validate mutual exclusion with --write-files at the top of RunE.
- Reuse agentsSnippet (or agentsSnippetCompact) as the content, write directly with os.WriteFile.
- Add feature scenarios: --output to new file, --output overwrite, --output + --write-files (error), --output + --compact, --output pointing to a directory (error).
