@implemented
Feature: Claim a task with a lease
  As an agent about to start work
  I want to claim a task for a bounded time
  So that other actors know not to pick up the same work

  Background:
    Given an initialized TaskLedger repository

  Scenario: Claiming a ready task records the claim and marks it in progress
    Given a ready task "task-abc123" titled "Add login form validation"
    When the agent runs `tl claim task-abc123 --actor claude-code:main`
    Then "task-abc123" is claimed by "claude-code:main"
    And "task-abc123" has status "in_progress"
    And "task-abc123" has a non-empty claim expiry
    And an event "claimed" is recorded for "task-abc123"

  Scenario: Claiming a task already claimed by another actor is rejected
    Given a task "task-abc123" claimed by "claude-code:frontend" with an active lease
    When the agent runs `tl claim task-abc123 --actor claude-code:main`
    Then the command exits with code 5
    And "task-abc123" is still claimed by "claude-code:frontend"

  Scenario: Claiming a task whose dependencies are not done is rejected
    Given a task "task-def456" with status "open"
    And a task "task-abc123" with status "open"
    And "task-abc123" depends on "task-def456"
    When the agent runs `tl claim task-abc123 --actor claude-code:main`
    Then the command exits with code 4
    And "task-abc123" is not claimed

  Scenario: Claiming with a custom lease and JSON output returns the claim
    Given a ready task "task-abc123"
    When the agent runs `tl claim task-abc123 --actor claude-code:main --ttl 120m --json`
    Then the JSON output contains actor "claude-code:main"
    And the JSON output contains a claim expiry 120 minutes after the claim time
