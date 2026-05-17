@implemented
Feature: Actor identity resolution for claims
  As an agent picking up work
  I want my identity to be resolved without needing --actor every time
  So that claims are frictionless while still preventing collisions

  Background:
    Given an initialized TaskLedger repository

  Scenario: CLI --actor takes highest priority
    Given environment variable "TL_ACTOR" is "env-agent"
    And a ready task "task-abc123"
    When the agent runs `tl claim task-abc123 --actor cli-agent`
    Then "task-abc123" is claimed by "cli-agent"

  Scenario: TL_ACTOR env var when --actor is absent
    Given environment variable "TL_ACTOR" is "pi:main"
    And a ready task "task-abc123"
    When the agent runs `tl claim task-abc123`
    Then "task-abc123" is claimed by "pi:main"

  Scenario: ACTOR_NAME as second env fallback
    Given environment variable "ACTOR_NAME" is "fallback-agent"
    And a ready task "task-abc123"
    When the agent runs `tl claim task-abc123`
    Then "task-abc123" is claimed by "fallback-agent"

  Scenario: BEADS_ACTOR as third env fallback
    Given environment variable "BEADS_ACTOR" is "beads-agent"
    And a ready task "task-abc123"
    When the agent runs `tl claim task-abc123`
    Then "task-abc123" is claimed by "beads-agent"

  Scenario: TL_ACTOR takes precedence over ACTOR_NAME and BEADS_ACTOR
    Given environment variable "TL_ACTOR" is "primary"
    And environment variable "ACTOR_NAME" is "secondary"
    And environment variable "BEADS_ACTOR" is "tertiary"
    And a ready task "task-abc123"
    When the agent runs `tl claim task-abc123`
    Then "task-abc123" is claimed by "primary"

  Scenario: Auto-detect agent name when no explicit actor is set
    Given the detected agent is "claude"
    And a ready task "task-abc123"
    When the agent runs `tl claim task-abc123`
    Then "task-abc123" is claimed by "claude"

  Scenario: Reject claim from a different actor
    Given a task "task-abc123" claimed by "claude-code:frontend" with an active lease
    And environment variable "TL_ACTOR" is "claude-code:main"
    When the agent runs `tl claim task-abc123`
    Then the command exits with code 5
    And "task-abc123" is still claimed by "claude-code:frontend"

  Scenario: Same actor can renew its own claim
    Given a task "task-abc123" claimed by "claude-code:main" with an active lease
    And environment variable "TL_ACTOR" is "claude-code:main"
    When the agent runs `tl claim task-abc123`
    Then the claim expiry for "task-abc123" is extended
