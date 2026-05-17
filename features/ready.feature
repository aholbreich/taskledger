@implemented
Feature: List ready tasks
  As an agent picking up work
  I want to see which open tasks are unblocked and unclaimed
  So that I can choose what to work on next safely

  Background:
    Given an initialized TaskLedger repository

  Scenario: An open task with no dependencies is ready
    Given a task "task-abc123" with status "open" and no dependencies
    When the agent runs `tl ready`
    Then the ready output contains "task-abc123"

  Scenario: An open task whose dependencies are all done is ready
    Given a task "task-def456" with status "done"
    And a task "task-abc123" with status "open"
    And "task-abc123" depends on "task-def456"
    When the agent runs `tl ready`
    Then the ready output contains "task-abc123"

  Scenario: An open task with a not-done dependency is not ready
    Given a task "task-def456" with status "open"
    And a task "task-abc123" with status "open"
    And "task-abc123" depends on "task-def456"
    When the agent runs `tl ready`
    Then the ready output does not contain "task-abc123"

  Scenario: A blocked task is not ready
    Given a task "task-abc123" with status "blocked"
    When the agent runs `tl ready`
    Then the ready output does not contain "task-abc123"

  Scenario: A task awaiting human input is not ready
    Given a task "task-abc123" with status "pending_human"
    When the agent runs `tl ready`
    Then the ready output does not contain "task-abc123"

  Scenario: A claimed in-progress task is not ready
    Given a task "task-abc123" claimed by "claude-code:main" with an active lease
    When the agent runs `tl ready`
    Then the ready output does not contain "task-abc123"

  Scenario: A task with a stale expired claim is ready
    Given a task "task-abc123" with an expired claim by "claude-code:main"
    When the agent runs `tl ready`
    Then the ready output contains "task-abc123"

  Scenario: A done task is not ready
    Given a task "task-abc123" with status "done"
    When the agent runs `tl ready`
    Then the ready output does not contain "task-abc123"

  Scenario: A cancelled task is not ready
    Given a task "task-abc123" with status "cancelled"
    When the agent runs `tl ready`
    Then the ready output does not contain "task-abc123"

  Scenario: The ready queue can be retrieved as JSON
    Given a task "task-abc123" titled "Add login form validation" with status "open" and no dependencies
    When the agent runs `tl ready --json`
    Then the JSON output is an array containing a task with identifier "task-abc123"
    And the JSON output contains title "Add login form validation"
    And the JSON output contains a priority for "task-abc123"

  Scenario: Ready can be filtered to tasks carrying a specific tag
    Given the following tasks exist:
      | id          | status | tags        |
      | task-aaa111 | open   | review      |
      | task-bbb222 | open   | docs        |
      | task-ccc333 | open   | review,arch |
      | task-ddd444 | open   |             |
    When the agent runs `tl ready --tag review`
    Then the ready output contains "task-aaa111"
    And the ready output contains "task-ccc333"
    And the ready output does not contain "task-bbb222"
    And the ready output does not contain "task-ddd444"

  Scenario: A tag filter does not override the ready criteria
    Given a task "task-abc123" with status "blocked" and tag "review"
    When the agent runs `tl ready --tag review`
    Then the ready output does not contain "task-abc123"
