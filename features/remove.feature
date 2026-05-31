@implemented
Feature: Remove a mistaken task from the active ledger
  As a developer cleaning up accidental or invalid task entries
  I want to remove a task file with an audited reason
  So that the active ledger stays clean while Git and the event journal preserve history

  Background:
    Given an initialized task ledger repository

  Scenario: Removing a cancelled task deletes the task file and records a reason
    Given a task "task-abc123" with status "cancelled"
    When the developer runs `tl remove task-abc123 -m "Created by mistake" --actor human`
    Then the task file for "task-abc123" does not exist
    And an event "removed" is recorded for "task-abc123" with value "Created by mistake"
    And the output contains "Removed task-abc123"

  Scenario: Removing without a reason is rejected
    Given a task "task-abc123" with status "cancelled"
    When the developer runs `tl remove task-abc123 --actor human`
    Then the command exits with code 2
    And the output reports that a reason is required
    And the task file for "task-abc123" exists

  Scenario: Removing a non-cancelled task requires force
    Given a task "task-abc123" with status "open"
    When the developer runs `tl remove task-abc123 -m "wrong task" --actor human`
    Then the command exits with code 4
    And the output reports that force is required
    And the task file for "task-abc123" exists

  Scenario: Force-removing a non-cancelled task succeeds
    Given a task "task-abc123" with status "open"
    When the developer runs `tl remove task-abc123 -m "wrong task" --actor human --force`
    Then the task file for "task-abc123" does not exist
    And an event "removed" is recorded for "task-abc123" with value "wrong task"

  Scenario: Removing a task with an active claim requires force
    Given a task "task-abc123" claimed by "agent-a" with an active lease
    When the developer runs `tl remove task-abc123 -m "wrong task" --actor human`
    Then the command exits with code 5
    And the command reports the claim is held by a different actor
    And the task file for "task-abc123" exists

  Scenario: Removing a task that other tasks depend on requires force
    Given a task "task-abc123" with status "cancelled"
    And a task "task-def456" with dependency "task-abc123"
    When the developer runs `tl remove task-abc123 -m "wrong task" --actor human`
    Then the command exits with code 4
    And the output reports that force is required
    And the task file for "task-abc123" exists

  Scenario: History remains readable after removing a task file
    Given a ready task "task-abc123"
    When the developer runs `tl remove task-abc123 -m "wrong task" --actor human --force`
    And the developer runs `tl history task-abc123`
    Then the output contains "created"
    And the output contains "removed"
