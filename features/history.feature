Feature: Inspect the event history of the ledger
  As a reviewer or auditor
  I want to see events recorded for a task or recently across all tasks
  So that I can reconstruct who did what and when

  Background:
    Given an initialized TaskLedger repository

  Scenario: History accumulates events for one task in chronological order
    Given a ready task "task-abc123"
    When the agent runs `tl claim task-abc123 --actor claude-code:main`
    And the agent runs `tl note task-abc123 --actor claude-code:main --message "WIP"`
    And the agent runs `tl close task-abc123 --actor claude-code:main`
    And the developer runs `tl history task-abc123`
    Then the history output lists events in this order:
      | event       |
      | created     |
      | claimed     |
      | note_added  |
      | closed      |

  Scenario: Each history entry records an actor and a timestamp
    Given a task "task-abc123" claimed by "claude-code:main" with an active lease
    When the developer runs `tl history task-abc123`
    Then the history output for the "claimed" event shows actor "claude-code:main"
    And the history output for the "claimed" event has a timestamp

  Scenario: History as JSON returns the raw event objects for the task
    Given a ready task "task-abc123"
    When the developer runs `tl history task-abc123 --json`
    Then the JSON output is an array of event objects for "task-abc123"
    And each event object contains a type, a timestamp, and an actor

  Scenario: History for an unknown task is rejected
    When the developer runs `tl history task-xyz`
    Then the command exits with code 3

  Scenario: History without TASK_ID or --since is rejected
    When the developer runs `tl history`
    Then the command exits with code 2
    And the output reports that a task ID or --since duration is required

  Scenario: History --since lists events across all tasks within the window
    Given the following events were recorded:
      | task        | event   | when         |
      | task-aaa111 | created | 3 hours ago  |
      | task-bbb222 | claimed | 1 hour ago   |
      | task-ccc333 | closed  | 30 hours ago |
    When the developer runs `tl history --since 6h`
    Then the history output contains an event for "task-aaa111"
    And the history output contains an event for "task-bbb222"
    And the history output does not contain an event for "task-ccc333"

  Scenario: History --since combined with TASK_ID filters to that task within the window
    Given the following events were recorded for "task-abc123":
      | event      | when           |
      | created    | 30 hours ago   |
      | claimed    | 2 hours ago    |
      | note_added | 30 minutes ago |
    When the developer runs `tl history task-abc123 --since 6h`
    Then the history output contains the event "claimed"
    And the history output contains the event "note_added"
    And the history output does not contain the event "created"
