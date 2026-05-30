---
id: task-obe
title: Add tl agents --compact flag for constrained context
status: open
priority: medium
type: feature
created_at: 2026-05-30T18:38:18Z
updated_at: 2026-05-30T18:38:18Z
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
  - cmd/agents_snippet.md
  - features/agents.feature
---

## Description

Add a --compact flag to 'tl agents' that emits a condensed ~12-15 line version of the agent workflow guide instead of the full ~70 line snippet.

Use case: agents with tight context windows (Claude Code long sessions, appended to system prompts, Copilot instructions) don't need the full 6-step narrative — they just need the command reference and key rules.

Behavior:
- 'tl agents --compact' prints a condensed snippet to stdout (does NOT write files).
- 'tl agents --write-files --compact' writes the compact version to agent instruction files instead of the full version. (Combined with --file to target specific files.)
- The compact format: bare commands with one-line explanations, no narrative paragraphs, no markdown headings. Example:

[
  {
    "id": "task-4sh",
    "title": "Design and implement tl doctor command",
    "status": "in_progress",
    "priority": "high",
    "type": "feature",
    "created_at": "2026-05-29T11:16:22Z",
    "updated_at": "2026-05-30T17:07:32Z",
    "created_by": "human",
    "assignee": null,
    "depends_on": [],
    "claim": {
      "actor": "claude:doctor",
      "claimed_at": "2026-05-30T17:07:32Z",
      "expires_at": "2026-05-30T18:07:32Z",
      "heartbeat_at": "2026-05-30T17:07:32Z"
    },
    "tags": [
      "cli"
    ],
    "description": "Design and implement the `tl doctor` command — a ledger diagnostic tool that scans task files, events, config, and filesystem for integrity issues.",
    "notes": [
      {
        "time": "2026-05-29T11:16:43Z",
        "actor": "main-pc",
        "kind": "note",
        "message": "Drafted initial Gherkin feature file at features/doctor.feature with 31 scenarios covering all agreed diagnostic categories, JSON output shape, and --fix behavior. Key design points captured in the task body. Ready for human review and agreement on the Gherkin before implementation."
      },
      {
        "time": "2026-05-29T11:23:48Z",
        "actor": "main-pc",
        "kind": "note",
        "message": "Deep analysis: the artifact-traceability gap and the proposed references field"
      },
      {
        "time": "2026-05-29T11:23:51Z",
        "actor": "main-pc",
        "kind": "note",
        "message": "\"Deep analysis: the artifact-traceability gap. See analysis in parent message.\""
      },
      {
        "time": "2026-05-29T13:50:41Z",
        "actor": "claude-code",
        "kind": "note",
        "message": "Refined features/doctor.feature per review. Applied: (1) compressed repetitive 3-liners into Scenario Outlines — frontmatter/dependency/timestamps/body-markers/config/scale now use Examples tables, file went from 35 scenarios in 193 lines to 24 declared scenarios in ~210 lines covering ~35 effective test runs; (2) removed 'tasks with no events' scenario — too prone to false positives for pre-journal or imported tasks; (3) added explicit severity model (error vs warning) declared in a feature-header comment and asserted on every category — expired-claim and open-with-claim-data are warnings (recoverable), in_progress-with-no-claim is an error (state inconsistent), claims category now has mixed severity; (4) scale thresholds now severity warning with header comment explaining the 100/1000 rationale (where filesystem/journal scans become noticeable); (5) added scenario for --fix on an expired claim returning it to open, plus an inline comment noting lock protection makes the racy-actor case safe. Skipped --fix --dry-run per human direction."
      },
      {
        "time": "2026-05-29T14:04:59Z",
        "actor": "main-pc",
        "kind": "note",
        "message": "Created docs/import-sync-PRD.md — comprehensive PRD covering JSON pipe import, markdown import, GitHub Issues import, JIRA/Linear/Trello import, and Trello bidirectional sync with shallow tasks. Includes 20 open questions marked Q1-Q20 for discussion."
      },
      {
        "time": "2026-05-29T14:05:17Z",
        "actor": "claude-code",
        "kind": "note",
        "message": "Cross-cutting note from task-reg refinement: when task-reg lands, doctor.feature should grow ~5 scenarios for reference validation (URL skip, bare identifier skip, path exists, path missing -\u003e warning, --fix removes dead path refs). Whichever of {task-4sh, task-reg} ships second is responsible for adding the reference-validation scenarios to the relevant feature file."
      },
      {
        "time": "2026-05-29T18:06:33Z",
        "actor": "claude-code:references",
        "kind": "note",
        "message": "When implementing tl doctor, extend features/doctor.feature with references validation (now that task-reg has shipped the field). Heuristic per task-reg design: URL-shaped (matches ^[a-z][a-z0-9+.-]*:) -\u003e skip; path-shaped (contains '/' no scheme) -\u003e treat as repo-relative path (relative to parent of .tl/), warn if missing; bare identifier/free text -\u003e skip. doctor --fix removes dead file-path refs; URL/identifier cases reported fixable:false."
      }
    ],
    "sections": {
      "Scope (agreed with human)": "### Diagnostics categories (all in v1):\n- **Frontmatter**: malformed YAML, unknown status/priority/type values, missing required fields\n- **Identity**: duplicate task IDs across the ledger\n- **Dependencies**: missing depends_on targets, cyclic deps (A→B→A), self-dependency\n- **Events**: orphaned events (ref nonexistent task), tasks with no events in journal\n- **Claims**: in_progress with no claim, expired claim not released, open with stale claim data\n- **Timestamps**: created_at \u003e updated_at, timestamps in the future, claim expiry before claim time\n- **Filesystem**: orphaned .md.tmp files, corrupted task files that can't be parsed\n- **Config**: invalid or missing config.yaml\n- **Body**: merge conflict markers (\u003c\u003c\u003c\u003c\u003c\u003c\u003c, =======, \u003e\u003e\u003e\u003e\u003e\u003e\u003e), malformed notes format\n- **Scale warning**: warn when \u003e100 tasks or \u003e1000 events exist\n\n### Output model:\n- Human-readable grouped report (grouped by category)\n- Exit 0 always (doctor is diagnostic, not a failure); non-zero only if doctor itself fails\n- --json support: emit array of diagnostic objects [{severity, category, task_id, message, fixable}]\n- --fix mode: attempt auto-repair where possible\n\n### Relationship to other commands:\n- Does NOT subsume tl stale — orthogonal\n\n### Process:\n1. First: write and agree the Gherkin feature file (features/doctor.feature)\n2. Second: implement the cmd/doctor.go and any needed internal/ packages\n3. Third: add BDD step definitions\n4. Tag @implemented when done"
    }
  },
  {
    "id": "task-z31",
    "title": "Overhaul README installation section: reorder hooks, clean up RPM, add tl init + agents flow",
    "status": "open",
    "priority": "high",
    "type": "chore",
    "created_at": "2026-05-30T18:38:12Z",
    "updated_at": "2026-05-30T18:38:12Z",
    "created_by": "human",
    "assignee": null,
    "depends_on": [],
    "claim": {
      "actor": null,
      "claimed_at": null,
      "expires_at": null,
      "heartbeat_at": null
    },
    "tags": [
      "docs",
      "promotion"
    ],
    "description": "The README installation section needs a promotion-focused rewrite. Current problems:\n\n1. LEAD PROBLEM: The 'Why tl cli?' section leads with chat/TODO/Issues frustration — these are true but generic. The real hook is the agent coordination story, which is tl's differentiator from Beads/Backlog.md. Rewrite to lead with the agent coordination value prop, then layer on the broader benefits.\n\n2. EMPTY GAP: There's a stray '---' line between 'How tl cli compares' and 'Installation Options' — looks like a removed section that wasn't cleaned up.\n\n3. RPM OVERPROMINENCE: RPM/Fedora instructions are ~15 lines of repo config for a pre-launch tool. Move to a collapsible details/summary block or to a separate docs/install.md page. Keep brew + install script + source as the mainline flows.\n\n4. NO AGENTS ONBOARDING: After 'tl init', the next logical step is 'tl agents --write-files' to bootstrap agent instructions — but the README doesn't mention this. Add a 'Setup for agent collaboration' subsection after Quickstart or Installation.\n\n5. NO ONE-LINER FLOW: The ideal onboarding sequence 'brew install tl \u0026\u0026 tl init \u0026\u0026 tl agents --write-files' isn't surfaced anywhere. Add a 'Quick start' box right after the badge bar.\n\n6. MISSING TL AGENTS IN COMMAND TABLE: The 'Commands' section lists 'tl agents' but doesn't show --write-files or other flags. Add the flags inline.\n\n7. COMPARISON SECTION BURIED: The comparison to Beads/Backlog.md is deep in the page. Consider moving the comparison table earlier, or link to it more prominently from the lead section.\n\nImplementation plan:\n- Restructure sections to: (1) Hero + badges + demo, (2) Why tl? (lead with agent coordination), (3) Quickstart (with init + agents one-liner), (4) Installation Options (brew first, RPM collapsed), (5) Commands, (6) How it compares, (7) Development, (8) Further reading\n- Add 'tl completion --install' alongside 'tl init' in Quickstart\n- Ensure all tl agents flag changes from task-xb3 (--write-files rename) are reflected"
  },
  {
    "id": "task-609",
    "title": "Add tl agents --dry-run flag for --write-files",
    "status": "open",
    "priority": "medium",
    "type": "feature",
    "created_at": "2026-05-30T18:24:20Z",
    "updated_at": "2026-05-30T18:24:20Z",
    "created_by": "human",
    "assignee": null,
    "depends_on": [],
    "claim": {
      "actor": null,
      "claimed_at": null,
      "expires_at": null,
      "heartbeat_at": null
    },
    "tags": [
      "agents"
    ],
    "description": "Add a --dry-run flag to 'tl agents --write-files' that reports what would be changed without modifying any files.\n\nBehavior:\n- 'tl agents --write-files --dry-run' scans the known agent instruction files (AGENTS.md, CLAUDE.md, GEMINI_RULES.md, etc.) and prints which files exist and would be updated, and which don't exist and would be skipped.\n- Output format: one line per file, e.g. 'Would update AGENTS.md (managed block found)' / 'Would update CLAUDE.md (no managed block yet, would append)' / 'Would skip GEMINI_RULES.md (file not found)'.\n- If --dry-run is passed without --write-files, return an error: '--dry-run requires --write-files'.\n- Exit code 0 even if no files would change (diagnostic, not failure).\n- The dry-run scan should reuse the same file-detection logic as the real update path to avoid drift.\n\nImplementation notes:\n- Add a 'dryRun' bool flag alongside 'writeFiles' in newAgentsCmd().\n- Extract file-scanning into a helper that returns a list of (path, action string) pairs.\n- The existing updateAgentInstructionFiles() function calls that helper and then acts on it.\n- Add feature scenarios in features/agents.feature for: dry-run with existing files, dry-run with no files, dry-run without --write-files (error)."
  },
  {
    "id": "task-phl",
    "title": "Launch checklist: distribution channels",
    "status": "open",
    "priority": "medium",
    "type": "feature",
    "created_at": "2026-05-29T16:52:45Z",
    "updated_at": "2026-05-29T16:52:45Z",
    "created_by": "human",
    "assignee": null,
    "depends_on": [],
    "claim": {
      "actor": null,
      "claimed_at": null,
      "expires_at": null,
      "heartbeat_at": null
    },
    "tags": [
      "promotion"
    ],
    "description": "GATED on green CI + demo + 'why tl' section landing first (one shot per channel). Sequence: (1) awesome-go + awesome-cli-apps PRs, (2) Show HN with the demo gif, (3) r/golang, r/commandline, (4) Lobste.rs, (5) agentic-coding communities — Claude Code / Cursor / aider Discords \u0026 subreddits (the real target audience). Track which channel converted. Do NOT fire before the storefront converts."
  },
  {
    "id": "task-xb3",
    "title": "Rename tl agents --update flag to --write-files (alias old name for compat)",
    "status": "open",
    "priority": "medium",
    "type": "chore",
    "created_at": "2026-05-30T18:24:15Z",
    "updated_at": "2026-05-30T18:24:15Z",
    "created_by": "human",
    "assignee": null,
    "depends_on": [],
    "claim": {
      "actor": null,
      "claimed_at": null,
      "expires_at": null,
      "heartbeat_at": null
    },
    "tags": [
      "agents"
    ],
    "description": "The --update flag on tl agents is misleading — the command doesn't update tl itself, it writes/refreshes agent instruction files. Rename to --write-files, which is self-documenting.\n\nChanges needed:\n1. In cmd/agents.go: rename the flag variable from 'update' to 'writeFiles', change flag name from 'update' to 'write-files', add 'update' as a hidden alias for backward compatibility: c.Flags().BoolVar(\u0026writeFiles, 'write-files', false, 'Write or refresh the tl workflow block in existing agent instruction files') + c.Flags().BoolVar(\u0026writeFiles, 'update', false, '(deprecated: use --write-files)') with MarkHidden on 'update'.\n2. In cmd/agents_test.go: update test names/comments that reference 'update'.\n3. In features/agents.feature: change '--update' references to '--write-files' (keep one scenario testing the old --update alias for backward compat).\n4. Update docs/usage.md 'tl agents' line in the Commands section if it references --update.\n5. Update README.md if it shows --update anywhere.\n6. Update AGENTS.md if it has inline tl agents --update mentions.\n\nThe old --update flag must continue to work (via hidden alias) — this is not a breaking change."
  },
  {
    "id": "task-zz0",
    "title": "Revise agents snippet content: remove make bdd leak, add --actor pattern, soften rigid language, add missing commands",
    "status": "open",
    "priority": "medium",
    "type": "chore",
    "created_at": "2026-05-30T18:24:09Z",
    "updated_at": "2026-05-30T18:24:09Z",
    "created_by": "human",
    "assignee": null,
    "depends_on": [],
    "claim": {
      "actor": null,
      "claimed_at": null,
      "expires_at": null,
      "heartbeat_at": null
    },
    "tags": [
      "agents"
    ],
    "description": "The tl agents snippet (cmd/agents_snippet.md) has several content issues:\n\n1. 'Do not begin implementation from chat instructions alone if there is no matching tl task' — too rigid, fights real workflow. Reword to guide rather than forbid.\n2. No mention of --actor flag. Agents that can't set env vars need --actor on each call. Add examples of both patterns (TL_ACTOR env vs --actor flag).\n3. 'check the current @implemented set with make bdd before relying on them' — leaks implementation detail into agent workflow. Remove or replace with a simple 'tl \u003ccmd\u003e --help to verify availability.'\n4. Missing commands: tl history \u003cid\u003e (mentioned in steps but should be listed explicitly), tl dep add/remove, tl stale, tl unblock, tl resolve.\n5. Step 1 says 'run tl show \u003ctask-id\u003e and tl history \u003ctask-id\u003e' but 'history' is missing from the command listing in Step 6's preamble.\n6. No context-economy mode — snippet is ~70 lines. Consider adding --compact flag (separate task scope, but the base snippet should be tighter).\n\nSee the critical review in session history for full context."
  },
  {
    "id": "task-2q0",
    "title": "Add tl agents --file flag for targeted updates (scope control)",
    "status": "open",
    "priority": "low",
    "type": "feature",
    "created_at": "2026-05-30T18:24:25Z",
    "updated_at": "2026-05-30T18:24:25Z",
    "created_by": "human",
    "assignee": null,
    "depends_on": [],
    "claim": {
      "actor": null,
      "claimed_at": null,
      "expires_at": null,
      "heartbeat_at": null
    },
    "tags": [
      "agents"
    ],
    "description": "Add a --file flag to 'tl agents --write-files' that limits which agent instruction files get updated, instead of always scanning the hardcoded list.\n\nBehavior:\n- --file can be repeated: 'tl agents --write-files --file CLAUDE.md --file .cursorrules'\n- If --file is given, only those files are considered for update. Files not in the list are ignored.\n- If --file is not given, the current default list applies (backward compatible).\n- The default list should be updated to include: AGENTS.md, CLAUDE.md, GEMINI_RULES.md (note: renamed from GEMINI.md), .cursorrules, .aider-rules.md, .github/copilot-instructions.md.\n- If --file points to a file that doesn't exist, skip it with a note (same as current behavior), don't create it.\n- The hardcoded agentInstructionFiles variable in cmd/agents.go becomes the default list; --file overrides it.\n- No error if --file names a nonexistent file — just output 'Skipped \u003cfile\u003e (not found)' and continue.\n- Add feature scenarios: --file single target, --file multiple targets (repeatable), --file with nonexistent file, --file with --dry-run."
  },
  {
    "id": "task-cys",
    "title": "Add tl agents --remove flag to strip managed blocks from agent files",
    "status": "open",
    "priority": "low",
    "type": "feature",
    "created_at": "2026-05-30T18:24:31Z",
    "updated_at": "2026-05-30T18:24:31Z",
    "created_by": "human",
    "assignee": null,
    "depends_on": [],
    "claim": {
      "actor": null,
      "claimed_at": null,
      "expires_at": null,
      "heartbeat_at": null
    },
    "tags": [
      "agents"
    ],
    "description": "Add a --remove flag to 'tl agents' that strips the managed tl workflow block (\u003c!-- BEGIN TL WORKFLOW --\u003e ... \u003c!-- END TL WORKFLOW --\u003e) from agent instruction files.\n\nBehavior:\n- 'tl agents --remove' scans the known agent instruction files and removes the managed block from each file that contains one.\n- Also handles the legacy block markers (\u003c!-- BEGIN TASKLEDGER WORKFLOW --\u003e ...).\n- Output format: one line per file modified, e.g. 'Removed tl workflow block from AGENTS.md'.\n- If no files contain a managed block, output 'No managed tl workflow blocks found'.\n- --remove can be combined with --file to target specific files: 'tl agents --remove --file CLAUDE.md'.\n- --remove combined with --dry-run shows what would be removed without modifying files: 'Would remove tl workflow block from CLAUDE.md'.\n- --remove and --write-files are mutually exclusive (error if both passed).\n- Exit code 0 even when nothing is removed (diagnostic only).\n- After removal, the surrounding whitespace should be cleaned up (no trailing blank lines left by the removed block).\n\nImplementation notes:\n- Implement removeAgentBlocks() function in cmd/agents.go.\n- Refactor mergeAgentsBlock() into a shared helper — the pattern of scanning files and applying an operation is shared by --write-files, --remove, and --dry-run.\n- Add feature scenarios in features/agents.feature: remove from one file, remove from multiple, remove with no blocks found, remove with --dry-run, remove + --write-files mutual exclusion, remove with --file targeting."
  },
  {
    "id": "task-hz9",
    "title": "Add a CHANGELOG.md",
    "status": "open",
    "priority": "low",
    "type": "chore",
    "created_at": "2026-05-29T16:52:45Z",
    "updated_at": "2026-05-29T16:52:45Z",
    "created_by": "human",
    "assignee": null,
    "depends_on": [],
    "claim": {
      "actor": null,
      "claimed_at": null,
      "expires_at": null,
      "heartbeat_at": null
    },
    "tags": [
      "docs"
    ],
    "description": "Backfill a Keep-a-Changelog style CHANGELOG from existing tags (0.1.0 -\u003e 0.6.0). Helps users trust release cadence."
  }
]

- The compact snippet should be embedded as a separate constant (agentsSnippetCompact) in cmd/agents.go alongside the existing agentsSnippet.
- The mergeAgentsBlock() function needs a param or variant for choosing compact vs full.
- Feature scenarios: --compact stdout, --compact + --write-files, --compact + --file targeting, verify compact content is shorter than full.
