@implemented
Feature: Cancel a task that will not be done
  As an actor abandoning work
  I want to mark a task cancelled with a reason
  So that the ledger reflects intentional abandonment without falsely claiming the work was completed

  Background:
    Given an initialized TaskLedger repository

  Scenario: An unclaimed open task can be cancelled with a reason
    Given a task "task-abc123" with status "open"
    When the developer runs `tl cancel task-abc123 -m "Superseded by task-def456."`
    Then "task-abc123" has status "cancelled"
    And "task-abc123" has a note containing "Superseded by task-def456."
    And an event "cancelled" is recorded for "task-abc123"

  Scenario: The claim's actor can cancel a claimed task
    Given a task "task-abc123" claimed by "claude-code:main" with an active lease
    When the agent runs `tl cancel task-abc123 --actor claude-code:main -m "Approach is wrong; opening replacement."`
    Then "task-abc123" has status "cancelled"
    And "task-abc123" is not claimed
    And an event "cancelled" is recorded for "task-abc123"

  Scenario: Cancelling a task claimed by another actor without force is rejected
    Given a task "task-abc123" claimed by "claude-code:frontend" with an active lease
    When the developer runs `tl cancel task-abc123 --actor human -m "stale"`
    Then the command reports the claim is held by a different actor
    And "task-abc123" does not have status "cancelled"

  Scenario: Forcing cancel of a task claimed by another actor succeeds
    Given a task "task-abc123" claimed by "claude-code:frontend" with an active lease
    When the developer runs `tl cancel task-abc123 --actor human --force -m "stale"`
    Then "task-abc123" has status "cancelled"
    And "task-abc123" is not claimed
    And an event "cancelled" is recorded for "task-abc123"

  Scenario: Cancelling without a reason is rejected
    Given a task "task-abc123" with status "open"
    When the developer runs `tl cancel task-abc123`
    Then the command exits with code 2
    And the output reports that a reason is required

  Scenario: Cancelling a done task is rejected
    Given a task "task-abc123" with status "done"
    When the developer runs `tl cancel task-abc123 -m "rethink"`
    Then the command reports the task is already closed
    And "task-abc123" still has status "done"

  Scenario: Cancelling an already-cancelled task is rejected
    Given a task "task-abc123" with status "cancelled"
    When the developer runs `tl cancel task-abc123 -m "again"`
    Then the command reports the task is already cancelled
    And "task-abc123" still has status "cancelled"

  Scenario: Cancelling with JSON output returns the updated task
    Given a task "task-abc123" with status "open"
    When the developer runs `tl cancel task-abc123 -m "drop" --json`
    Then the JSON output contains status "cancelled"
    And the JSON output contains identifier "task-abc123"
