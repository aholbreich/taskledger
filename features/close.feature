Feature: Close a completed task
  As an agent or developer
  I want to mark a task as done once verification has passed
  So that the ledger reflects completed work and contributes to the audit trail

  Background:
    Given an initialized TaskLedger repository

  Scenario: The claim's actor closes a task that has no verification commands
    Given a task "task-abc123" claimed by "claude-code:main" with no verification commands
    When the agent runs `tl close task-abc123 --actor claude-code:main`
    Then "task-abc123" has status "done"
    And "task-abc123" was closed by "claude-code:main"
    And an event "closed" is recorded for "task-abc123"

  Scenario: The claim's actor closes a task whose verification passes
    Given a task "task-abc123" claimed by "claude-code:main"
    And verification for "task-abc123" passes
    When the agent runs `tl close task-abc123 --actor claude-code:main`
    Then "task-abc123" has status "done"
    And an event "closed" is recorded for "task-abc123"

  Scenario: An unclaimed task can be closed by any actor
    Given a task "task-abc123" with status "open"
    When the developer runs `tl close task-abc123 --actor human`
    Then "task-abc123" has status "done"
    And "task-abc123" was closed by "human"
    And an event "closed" is recorded for "task-abc123"

  Scenario: Closing a blocked task is rejected
    Given a task "task-abc123" with status "blocked"
    When the developer runs `tl close task-abc123 --actor claude-code:main`
    Then the command reports the task is blocked
    And "task-abc123" still has status "blocked"

  Scenario: Closing a task whose verification does not pass is rejected
    Given a task "task-abc123" claimed by "claude-code:main"
    And verification for "task-abc123" does not pass
    When the agent runs `tl close task-abc123 --actor claude-code:main`
    Then the command exits with code 6
    And "task-abc123" does not have status "done"

  Scenario: Closing a task claimed by another actor without force is rejected
    Given a task "task-abc123" claimed by "claude-code:frontend" with an active lease
    When the developer runs `tl close task-abc123 --actor human`
    Then the command reports the claim is held by a different actor
    And "task-abc123" does not have status "done"

  Scenario: Closing a task claimed by another actor with force succeeds
    Given a task "task-abc123" claimed by "claude-code:frontend" with an active lease
    When the developer runs `tl close task-abc123 --actor human --force`
    Then "task-abc123" has status "done"
    And "task-abc123" was closed by "human"
    And an event "closed" is recorded for "task-abc123"

  Scenario: Closing an already-done task is rejected
    Given a task "task-abc123" with status "done"
    When the developer runs `tl close task-abc123 --actor human`
    Then the command reports the task is already closed
    And "task-abc123" still has status "done"

  Scenario: Closing with JSON output returns the updated task
    Given a task "task-abc123" claimed by "claude-code:main" with no verification commands
    When the agent runs `tl close task-abc123 --actor claude-code:main --json`
    Then the JSON output contains status "done"
    And the JSON output contains identifier "task-abc123"
