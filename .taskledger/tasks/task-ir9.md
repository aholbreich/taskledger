---
id: task-ir9
title: Omit raw body from list and ready JSON
status: done
priority: medium
type: task
created_at: 2026-05-24T16:15:02Z
updated_at: 2026-05-24T16:16:45Z
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

Make bulk JSON outputs compact by omitting the raw Markdown body while preserving parsed description and notes.

## Notes

- 2026-05-24T16:16:45Z [pi:compact-json] note: Omitted raw body from list/ready JSON via compact task DTO while preserving parsed description/notes; added BDD coverage and docs; make bdd and go test ./... passed.
