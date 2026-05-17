@implemented
Feature: List stale claims
  As a developer or agent
  I want to see claims past their expiry
  So that abandoned work can be reclaimed safely

  Background:
    Given an initialized TaskLedger repository

  Scenario: A claim past its expiry appears in the stale list
    Given a task "task-abc123" with an expired claim by "claude-code:main"
    When the developer runs `tl stale`
    Then the stale output contains "task-abc123"

  Scenario: A claim still within its lease does not appear in the stale list
    Given a task "task-abc123" claimed by "claude-code:main" with an active lease
    When the developer runs `tl stale`
    Then the stale output does not contain "task-abc123"

  Scenario: The stale list can be retrieved as JSON
    Given a task "task-abc123" with an expired claim by "claude-code:frontend"
    When the developer runs `tl stale --json`
    Then the JSON output is an array containing a stale claim for "task-abc123"
    And the JSON output contains actor "claude-code:frontend"
