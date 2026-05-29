@implemented
Feature: Task ID autocompletion
  As a developer or agent
  I want pressing TAB on a task-ID argument to suggest the actual task IDs
  So that I do not have to memorise or copy-paste identifiers

  Background:
    Given an initialized task ledger repository

  Scenario: Completing a task-id argument suggests every active task with status and title as description
    Given a task "task-abc123" titled "Add login form validation" with status "open"
    And a task "task-def456" titled "Refactor auth errors" with status "blocked"
    When the developer runs `tl __complete claim ""`
    Then the completion suggestion "task-abc123" appears with description "[open] Add login form validation"
    And the completion suggestion "task-def456" appears with description "[blocked] Refactor auth errors"
    And the completion directive is "ShellCompDirectiveNoFileComp"
    And the completion directive is "ShellCompDirectiveKeepOrder"

  Scenario: Completion filters suggestions to the typed prefix
    Given a task "task-abc123" titled "Add login form validation" with status "open"
    And a task "task-xyz789" titled "Refactor auth errors" with status "open"
    When the developer runs `tl __complete show "task-ab"`
    Then the completion suggestion "task-abc123" is present
    And the completion suggestion "task-xyz789" is absent

  Scenario: Completion supports bare short codes
    Given a task "task-abc123" titled "Add login form validation" with status "open"
    And a task "task-xyz789" titled "Refactor auth errors" with status "open"
    When the developer runs `tl __complete show "ab"`
    Then the completion suggestion "abc123" is present
    And the completion suggestion "xyz789" is absent

  Scenario Outline: Completion is wired up for every command that takes a task-id positional argument
    Given a task "task-abc123" titled "Add login form validation" with status "open"
    When the developer runs `tl __complete <command> ""`
    Then the completion suggestion "task-abc123" is present

    Examples:
      | command |
      | show    |
      | claim   |
      | close   |
      | note    |
      | block   |
      | cancel  |
      | unblock |
      | pending |
      | resolve |
      | release |
      | refine  |
      | history |

  Scenario: Completing the source argument of dep add suggests task IDs
    Given a task "task-abc123" titled "Add login form validation" with status "open"
    When the developer runs `tl __complete dep add ""`
    Then the completion suggestion "task-abc123" is present

  Scenario: Completing the source argument of dep remove suggests task IDs
    Given a task "task-abc123" titled "Add login form validation" with status "open"
    When the developer runs `tl __complete dep remove ""`
    Then the completion suggestion "task-abc123" is present

  Scenario: Completing the --on flag of dep add suggests task IDs
    Given a task "task-abc123" titled "Add login form validation" with status "open"
    And a task "task-def456" titled "Refactor auth errors" with status "open"
    When the developer runs `tl __complete dep add task-abc123 --on ""`
    Then the completion suggestion "task-def456" is present

  Scenario: Completing a task-id argument when no tasks exist yields no suggestions
    When the developer runs `tl __complete show ""`
    Then no task-ID completion suggestions are returned
    And the completion directive is "ShellCompDirectiveNoFileComp"

  Scenario: Completion hides done and cancelled tasks across every command
    Given a task "task-abc123" titled "Active task" with status "open"
    And a task "task-old111" titled "Finished task" with status "done"
    And a task "task-old222" titled "Abandoned task" with status "cancelled"
    When the developer runs `tl __complete claim ""`
    Then the completion suggestion "task-abc123" is present
    And the completion suggestion "task-old111" is absent
    And the completion suggestion "task-old222" is absent

  Scenario: Completion hides done tasks from show too — type the full ID to inspect archives
    Given a task "task-abc123" titled "Active task" with status "open"
    And a task "task-old111" titled "Finished task" with status "done"
    When the developer runs `tl __complete show ""`
    Then the completion suggestion "task-abc123" is present
    And the completion suggestion "task-old111" is absent
