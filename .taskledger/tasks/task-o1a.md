---
id: task-o1a
title: Specify editor mode for tl refine
status: done
priority: medium
type: task
created_at: 2026-05-24T16:31:22Z
updated_at: 2026-05-24T16:38:56Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags: []
---

## Description

Add Gherkin scenarios for tl refine --edit opening a system editor to update editable task fields.

## Notes

- 2026-05-24T16:32:45Z [pi:refine-edit-feature] note: Drafted unimplemented Gherkin contract in features/refine-edit.feature for tl refine --edit. Kept existing implemented refine.feature unchanged so make bdd still passes. Awaiting review.
- 2026-05-24T16:35:08Z [pi:refine-edit] resolved: Approved; implement editor-mode refine contract as written.
- 2026-05-24T16:38:55Z [pi:refine-edit] note: Implemented tl refine --edit with VISUAL/EDITOR temp-buffer workflow, validation, no-op handling, BDD coverage, and docs. Verification: make bdd and go test ./... passed.
