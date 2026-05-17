@implemented
Feature: List tasks in the ledger
  As a developer or agent
  I want to see active tasks in the ledger
  So that I can understand the work that still needs attention

  Background:
    Given an initialized TaskLedger repository

  Scenario: Listing tasks shows the default table columns for active tasks
    Given the following tasks exist:
      | id          | status      | priority | claimed by | title                        |
      | task-abc123 | open        | high     |            | Add login form validation    |
      | task-def456 | in_progress | medium   | codex      | Refactor auth error messages |
      | task-ghi789 | done        | low      |            | Document login flow          |
    When the developer runs `tl list`
    Then the list output columns are exactly:
      | column     |
      | ID         |
      | Status     |
      | Priority   |
      | Claimed By |
      | Title      |
    And the output lists "task-abc123" with status "open", priority "high", claimed by "-", and title "Add login form validation"
    And the output lists "task-def456" with status "in_progress", priority "medium", claimed by "codex", and title "Refactor auth error messages"
    And the output does not list "task-ghi789"

  Scenario: Listing tasks with JSON output hides closed tasks by default
    Given the following tasks exist:
      | id          | status    | title                     |
      | task-abc123 | open      | Add login form validation |
      | task-def456 | cancelled | Remove obsolete auth flow |
      | task-ghi789 | done      | Document login flow       |
    When the developer runs `tl list --json`
    Then the JSON output is an array of 1 tasks
    And the JSON output contains a task with identifier "task-abc123"
    And the JSON output does not contain a task with identifier "task-def456"
    And the JSON output does not contain a task with identifier "task-ghi789"

  Scenario: Listing tasks with --all includes closed tasks
    Given the following tasks exist:
      | id          | status    | priority | title                     |
      | task-abc123 | open      | high     | Add login form validation |
      | task-def456 | cancelled | medium   | Remove obsolete auth flow |
      | task-ghi789 | done      | low      | Document login flow       |
    When the developer runs `tl list --all`
    Then the output lists "task-abc123" with status "open", priority "high", claimed by "-", and title "Add login form validation"
    And the output lists "task-def456" with status "cancelled", priority "medium", claimed by "-", and title "Remove obsolete auth flow"
    And the output lists "task-ghi789" with status "done", priority "low", claimed by "-", and title "Document login flow"

  Scenario: Listing tasks orders active tasks by priority and identifier
    Given the following tasks exist:
      | id          | status | priority | claimed by | title                        |
      | task-ccc333 | open   | low      | human      | Document login flow          |
      | task-bbb222 | open   | high     | codex      | Add login form validation    |
      | task-aaa111 | open   | high     | pi:1       | Refactor auth error messages |
      | task-ddd444 | open   | medium   | codex      | Review login copy            |
    When the developer runs `tl list`
    Then the listed task identifiers appear in this order:
      | id          |
      | task-aaa111 |
      | task-bbb222 |
      | task-ddd444 |
      | task-ccc333 |

  Scenario: Listing tasks can be filtered by claim actor
    Given the following tasks exist:
      | id          | status      | priority | claimed by | title                        |
      | task-aaa111 | in_progress | high     | pi:1       | Refactor auth error messages |
      | task-bbb222 | in_progress | high     | codex      | Add login form validation    |
      | task-ccc333 | open        | low      |            | Document login flow          |
      | task-ddd444 | in_progress | medium   | codex      | Review login copy            |
      | task-eee555 | done        | high     | codex      | Closed codex task            |
    When the developer runs `tl list --claimed-by codex`
    Then the listed task identifiers appear in this order:
      | id          |
      | task-bbb222 |
      | task-ddd444 |
    And the output does not list "task-aaa111"
    And the output does not list "task-ccc333"
    And the output does not list "task-eee555"

  Scenario: Listing tasks can be filtered by status
    Given the following tasks exist:
      | id          | status        | title         |
      | task-aaa111 | open          | First         |
      | task-bbb222 | pending_human | Second        |
      | task-ccc333 | blocked       | Third         |
      | task-ddd444 | in_progress   | Fourth        |
    When the developer runs `tl list --status pending_human`
    Then the output lists "task-bbb222"
    And the output does not list "task-aaa111"
    And the output does not list "task-ccc333"
    And the output does not list "task-ddd444"

  Scenario: Filtering by a closed status reveals closed tasks without --all
    Given the following tasks exist:
      | id          | status    | title              |
      | task-aaa111 | open      | Active             |
      | task-bbb222 | cancelled | Abandoned          |
      | task-ccc333 | done      | Finished           |
    When the developer runs `tl list --status cancelled`
    Then the output lists "task-bbb222"
    And the output does not list "task-aaa111"
    And the output does not list "task-ccc333"

  Scenario: --mine filters tasks claimed by the resolved actor
    Given environment variable "TL_ACTOR" is "codex"
    And the following tasks exist:
      | id          | status      | claimed by | title    |
      | task-aaa111 | in_progress | codex      | Mine A   |
      | task-bbb222 | in_progress | pi:1       | Theirs   |
      | task-ccc333 | in_progress | codex      | Mine B   |
    When the developer runs `tl list --mine`
    Then the output lists "task-aaa111"
    And the output lists "task-ccc333"
    And the output does not list "task-bbb222"

  Scenario: Listing tasks can be filtered by tag
    Given the following tasks exist:
      | id          | status      | tags        | title              |
      | task-aaa111 | open        | review      | Review auth flow   |
      | task-bbb222 | in_progress | docs        | Document login     |
      | task-ccc333 | open        | review,arch | Review architecture|
      | task-ddd444 | open        |             | No tags            |
    When the developer runs `tl list --tag review`
    Then the output lists "task-aaa111"
    And the output lists "task-ccc333"
    And the output does not list "task-bbb222"
    And the output does not list "task-ddd444"
