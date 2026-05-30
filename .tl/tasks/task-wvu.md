---
id: task-wvu
title: 'Extend agent auto-detection: aider, windsurf, pi env vars'
status: open
priority: low
type: feature
created_at: 2026-05-30T18:38:46Z
updated_at: 2026-05-30T18:38:49Z
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
  - cmd/actor.go
  - cmd/actor_test.go
  - features/actor.feature
---

## Description

Extend the agent auto-detection in cmd/actor.go (defaultDetectActor) to cover more coding harnesses.

Current detection (cmd/actor.go):
1. CLAUDE_CODE_SESSION_ID → 'claude-code' ✓ (reliable, Claude Code sets this per-session)
2. .codex file → 'codex' ✓ (OpenAI Codex CLI)
3. PI_AGENT_ID → 'pi' (inconsistently set)
4. hostname → fallback

Research findings on reliable env var markers:

| Agent            | Reliable Env Var(s)                      | Reliability |
|------------------|------------------------------------------|-------------|
| Claude Code      | CLAUDE_CODE_SESSION_ID (UUID)            | HIGH — set per interactive session |
| Aider            | AIDER_MODEL or AIDER_ARCH               | HIGH — always set when aider runs |
| Windsurf/Codeium | CODEIUM_API_KEY or CODEIUM_SESSION_ID    | HIGH — set in Windsurf's integrated terminal |
| Pi               | PI_AGENT_ID or PI_CODING_AGENT=true      | MEDIUM — PI_CODING_AGENT=true is reliable; PI_AGENT_ID varies |
| Cursor           | No reliable single env var               | LOW — VSCODE_IPC_HOOK exists but many tools set it |
| GitHub Copilot   | GITHUB_COPILOT_TOKEN                     | LOW — auth token, not an agent marker |

Changes:
1. Add 'AIDER_MODEL' check → return 'aider' (check this first in the aider section since it's most specific).
2. Add 'CODEIUM_API_KEY' check → return 'windsurf' (Codeium/Windsurf).
3. Add 'PI_CODING_AGENT=true' check → return 'pi' (as a complement to PI_AGENT_ID, checked first since it's more reliably set).
4. Reorder the detection chain: PI_CODING_AGENT → CLAUDE_CODE_SESSION_ID → AIDER_MODEL → CODEIUM_API_KEY → PI_AGENT_ID → .codex file → hostname.

Note about Cursor: No reliable env var exists. Cursor is a VS Code fork and sets VSCODE_* vars that are also set by regular VS Code. False positive risk is too high. Skip Cursor for env-var detection.

Note about Pi: PI_CODING_AGENT=true is set by the pi harness. This is more reliable than PI_AGENT_ID which is optional. Check PI_CODING_AGENT first, then PI_AGENT_ID as fallback.

Also update the ActorEnvChain to include PI_CODING_AGENT? No — PI_CODING_AGENT is a boolean flag (true/false), not an actor name. Keep it in auto-detection only.

Update cmd/actor_test.go with tests for the new detection paths.
Update features/actor.feature with scenarios for each new agent marker.

## Notes

- 2026-05-30T18:38:49Z [pi:planning] note: Behavior details: - No --actor flag, no TL_ACTOR/ACTOR_NAME/BEADS_ACTOR, but AIDER_MODEL is set → actor becomes 'aider' - No --actor flag, no TL_ACTOR/ACTOR_NAME/BEADS_ACTOR, but CODEIUM_API_KEY is set → actor becomes 'windsurf' - No --actor flag, no TL_ACTOR/ACTOR_NAME/BEADS_ACTOR, but PI_CODING_AGENT=true → actor becomes 'pi' (checked before PI_AGENT_ID) - Detection order: PI_CODING_AGENT → CLAUDE_CODE_SESSION_ID → AIDER_MODEL → CODEIUM_API_KEY → PI_AGENT_ID → .codex → hostname - Cursor intentionally skipped: VSCODE_IPC_HOOK gives false positives with regular VS Code
