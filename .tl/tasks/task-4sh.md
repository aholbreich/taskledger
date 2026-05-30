---
id: task-4sh
title: Design and implement tl doctor command
status: done
priority: high
type: feature
created_at: 2026-05-29T11:16:22Z
updated_at: 2026-05-30T19:47:27Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - cli
references:
  - file:features/doctor.feature
---

## Description


Design and implement the `tl doctor` command — a ledger diagnostic tool that scans task files, events, config, and filesystem for integrity issues.

## Scope (agreed with human)

### Diagnostics categories (all in v1):
- **Frontmatter**: malformed YAML, unknown status/priority/type values, missing required fields
- **Identity**: duplicate task IDs across the ledger
- **Dependencies**: missing depends_on targets, cyclic deps (A→B→A), self-dependency
- **Events**: orphaned events (ref nonexistent task), tasks with no events in journal
- **Claims**: in_progress with no claim, expired claim not released, open with stale claim data
- **Timestamps**: created_at > updated_at, timestamps in the future, claim expiry before claim time
- **Filesystem**: orphaned .md.tmp files, corrupted task files that can't be parsed
- **Config**: invalid or missing config.yaml
- **Body**: merge conflict markers (<<<<<<<, =======, >>>>>>>), malformed notes format
- **Scale warning**: warn when >100 tasks or >1000 events exist

### Output model:
- Human-readable grouped report (grouped by category)
- Exit 0 always (doctor is diagnostic, not a failure); non-zero only if doctor itself fails
- --json support: emit array of diagnostic objects [{severity, category, task_id, message, fixable}]
- --fix mode: attempt auto-repair where possible

### Relationship to other commands:
- Does NOT subsume tl stale — orthogonal

### Process:
1. First: write and agree the Gherkin feature file (features/doctor.feature)
2. Second: implement the cmd/doctor.go and any needed internal/ packages
3. Third: add BDD step definitions
4. Tag @implemented when done

## Notes

- 2026-05-29T11:16:43Z [main-pc] note: Drafted initial Gherkin feature file at features/doctor.feature with 31 scenarios covering all agreed diagnostic categories, JSON output shape, and --fix behavior. Key design points captured in the task body. Ready for human review and agreement on the Gherkin before implementation.
- 2026-05-29T11:23:48Z [main-pc] note: Deep analysis: the artifact-traceability gap and the proposed references field
- 2026-05-29T11:23:51Z [main-pc] note: "Deep analysis: the artifact-traceability gap. See analysis in parent message."
- 2026-05-29T13:50:41Z [claude-code] note: Refined features/doctor.feature per review. Applied: (1) compressed repetitive 3-liners into Scenario Outlines — frontmatter/dependency/timestamps/body-markers/config/scale now use Examples tables, file went from 35 scenarios in 193 lines to 24 declared scenarios in ~210 lines covering ~35 effective test runs; (2) removed 'tasks with no events' scenario — too prone to false positives for pre-journal or imported tasks; (3) added explicit severity model (error vs warning) declared in a feature-header comment and asserted on every category — expired-claim and open-with-claim-data are warnings (recoverable), in_progress-with-no-claim is an error (state inconsistent), claims category now has mixed severity; (4) scale thresholds now severity warning with header comment explaining the 100/1000 rationale (where filesystem/journal scans become noticeable); (5) added scenario for --fix on an expired claim returning it to open, plus an inline comment noting lock protection makes the racy-actor case safe. Skipped --fix --dry-run per human direction.
- 2026-05-29T14:04:59Z [main-pc] note: Created docs/import-sync-PRD.md — comprehensive PRD covering JSON pipe import, markdown import, GitHub Issues import, JIRA/Linear/Trello import, and Trello bidirectional sync with shallow tasks. Includes 20 open questions marked Q1-Q20 for discussion.
- 2026-05-29T14:05:17Z [claude-code] note: Cross-cutting note from task-reg refinement: when task-reg lands, doctor.feature should grow ~5 scenarios for reference validation (URL skip, bare identifier skip, path exists, path missing -> warning, --fix removes dead path refs). Whichever of {task-4sh, task-reg} ships second is responsible for adding the reference-validation scenarios to the relevant feature file.
- 2026-05-29T18:06:33Z [claude-code:references] note: When implementing tl doctor, extend features/doctor.feature with references validation (now that task-reg has shipped the field). Heuristic per task-reg design: URL-shaped (matches ^[a-z][a-z0-9+.-]*:) -> skip; path-shaped (contains '/' no scheme) -> treat as repo-relative path (relative to parent of .tl/), warn if missing; bare identifier/free text -> skip. doctor --fix removes dead file-path refs; URL/identifier cases reported fixable:false.
- 2026-05-30T19:47:27Z [pi:doctor] note: Analysis complete: ClaudeCode actually finished the entire implementation before being interrupted. Verified: - cmd/doctor.go wired into root.go ✅ - internal/doctor/doctor.go covers all 10 diagnostic categories ✅ - 13 unit tests in internal/doctor/doctor_test.go all pass ✅ - Full BDD step definitions in bdd/doctor_test.go ✅ - features/doctor.feature correctly tagged @implemented ✅ - Live tested: 'go run . doctor' finds 10 real issues, --json outputs correctly Closing — no implementation work remaining.
