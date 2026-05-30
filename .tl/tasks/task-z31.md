---
id: task-z31
title: 'Overhaul README installation section: reorder hooks, clean up RPM, add tl init + agents flow'
status: open
priority: high
type: chore
created_at: 2026-05-30T18:38:12Z
updated_at: 2026-05-30T18:38:12Z
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
  - promotion
references:
  - README.md
  - docs/comparison.md
  - docs/PRD.md
  - docs/usage.md
---

## Description

The README installation section needs a promotion-focused rewrite. Current problems:

1. LEAD PROBLEM: The 'Why tl cli?' section leads with chat/TODO/Issues frustration — these are true but generic. The real hook is the agent coordination story, which is tl's differentiator from Beads/Backlog.md. Rewrite to lead with the agent coordination value prop, then layer on the broader benefits.

2. EMPTY GAP: There's a stray '---' line between 'How tl cli compares' and 'Installation Options' — looks like a removed section that wasn't cleaned up.

3. RPM OVERPROMINENCE: RPM/Fedora instructions are ~15 lines of repo config for a pre-launch tool. Move to a collapsible details/summary block or to a separate docs/install.md page. Keep brew + install script + source as the mainline flows.

4. NO AGENTS ONBOARDING: After 'tl init', the next logical step is 'tl agents --write-files' to bootstrap agent instructions — but the README doesn't mention this. Add a 'Setup for agent collaboration' subsection after Quickstart or Installation.

5. NO ONE-LINER FLOW: The ideal onboarding sequence 'brew install tl && tl init && tl agents --write-files' isn't surfaced anywhere. Add a 'Quick start' box right after the badge bar.

6. MISSING TL AGENTS IN COMMAND TABLE: The 'Commands' section lists 'tl agents' but doesn't show --write-files or other flags. Add the flags inline.

7. COMPARISON SECTION BURIED: The comparison to Beads/Backlog.md is deep in the page. Consider moving the comparison table earlier, or link to it more prominently from the lead section.

Implementation plan:
- Restructure sections to: (1) Hero + badges + demo, (2) Why tl? (lead with agent coordination), (3) Quickstart (with init + agents one-liner), (4) Installation Options (brew first, RPM collapsed), (5) Commands, (6) How it compares, (7) Development, (8) Further reading
- Add 'tl completion --install' alongside 'tl init' in Quickstart
- Ensure all tl agents flag changes from task-xb3 (--write-files rename) are reflected
