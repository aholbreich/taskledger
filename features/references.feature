@implemented
Feature: Task references
  As a developer or agent
  I want to attach references to a task — file paths, URLs, tickets, ADRs, anything
  So that humans and LLMs can find the artefacts that go with the work

  # References are generic strings; tl does not validate at input time.
  # Validation happens later in `tl doctor` and only for entries that look
  # like repo-relative file paths.

  Background:
    Given an initialized task ledger repository

  # -------------------------------------------------------------------------
  # Create — initial references via --ref (repeatable). Bare shorthand inherits.
  # -------------------------------------------------------------------------
  Scenario: Creating a task with a single reference
    When the developer runs `tl create "Add login form" --ref src/auth/login.go`
    Then a new task with title "Add login form" exists
    And the new task has references containing "src/auth/login.go"

  Scenario: Creating a task with multiple references of mixed kinds
    When the developer runs `tl create "Add login form" --ref src/auth/login.go --ref features/login.feature --ref https://github.com/aholbreich/tl/pull/42 --ref JIRA-1234`
    Then the new task has references containing "src/auth/login.go"
    And the new task has references containing "features/login.feature"
    And the new task has references containing "https://github.com/aholbreich/tl/pull/42"
    And the new task has references containing "JIRA-1234"

  Scenario: The bare shorthand inherits --ref
    When the developer runs `tl "Add login form" --ref src/auth/login.go --tag auth`
    Then a new task with title "Add login form" exists
    And the new task has references containing "src/auth/login.go"
    And the new task has tag "auth"

  Scenario: Creating a task without --ref produces an empty references list
    When the developer runs `tl create "Add login form"`
    Then the new task has no references

  # -------------------------------------------------------------------------
  # Refine — add and remove are idempotent (mirror tl dep add/remove).
  # -------------------------------------------------------------------------
  Scenario: refine --add-ref appends a new reference
    Given a task "task-abc" with no references
    When the developer runs `tl refine task-abc --add-ref src/auth/login.go`
    Then "task-abc" has references containing "src/auth/login.go"

  Scenario: refine --add-ref is idempotent
    Given a task "task-abc" with reference "src/auth/login.go"
    When the developer runs `tl refine task-abc --add-ref src/auth/login.go`
    Then "task-abc" has exactly one reference matching "src/auth/login.go"

  Scenario: refine --remove-ref removes a reference
    Given a task "task-abc" with reference "src/auth/login.go"
    When the developer runs `tl refine task-abc --remove-ref src/auth/login.go`
    Then "task-abc" does not have reference "src/auth/login.go"

  Scenario: refine --remove-ref on a missing reference is a no-op
    Given a task "task-abc" with no references
    When the developer runs `tl refine task-abc --remove-ref src/auth/login.go`
    Then the command exits with code 0
    And "task-abc" has no references

  Scenario: refine can add and remove multiple references in one invocation
    Given a task "task-abc" with reference "old/path.go"
    When the developer runs `tl refine task-abc --add-ref features/login.feature --add-ref JIRA-1234 --remove-ref old/path.go`
    Then "task-abc" has references containing "features/login.feature"
    And "task-abc" has references containing "JIRA-1234"
    And "task-abc" does not have reference "old/path.go"

  # -------------------------------------------------------------------------
  # Display — frontmatter-only field, between Depends On and Claim.
  # -------------------------------------------------------------------------
  Scenario: tl show renders the References field between Depends On and Claim
    Given a task "task-abc" with references "src/auth/login.go" and "https://github.com/aholbreich/tl/pull/42"
    When the developer runs `tl show task-abc`
    Then the output contains "References:"
    And the output contains "src/auth/login.go"
    And the output contains "https://github.com/aholbreich/tl/pull/42"
    And the "References" line appears after the "Depends On" line
    And the "References" line appears before the "Claim" line

  Scenario: tl show with no references reports "none"
    Given a task "task-abc" with no references
    When the developer runs `tl show task-abc`
    Then the output contains "References: none"

  Scenario: JSON output exposes references as an array
    Given a task "task-abc" with references "src/auth/login.go" and "JIRA-1234"
    When the developer runs `tl show task-abc --json`
    Then the JSON output has a "references" array containing "src/auth/login.go"
    And the JSON output has a "references" array containing "JIRA-1234"

  Scenario: JSON output emits an empty array when there are no references
    Given a task "task-abc" with no references
    When the developer runs `tl show task-abc --json`
    Then the JSON output has an empty "references" array

  # -------------------------------------------------------------------------
  # Events — one event per reference added or removed.
  # -------------------------------------------------------------------------
  Scenario: Creating a task with references records a reference_added event per reference
    When the developer runs `tl create "Add login form" --ref src/auth/login.go --ref features/login.feature`
    Then a "reference_added" event is recorded for the new task with value "src/auth/login.go"
    And a "reference_added" event is recorded for the new task with value "features/login.feature"

  Scenario: refine --add-ref records a reference_added event
    Given a task "task-abc" with no references
    When the developer runs `tl refine task-abc --add-ref src/auth/login.go`
    Then a "reference_added" event is recorded for "task-abc" with value "src/auth/login.go"

  Scenario: refine --remove-ref records a reference_removed event
    Given a task "task-abc" with reference "src/auth/login.go"
    When the developer runs `tl refine task-abc --remove-ref src/auth/login.go`
    Then a "reference_removed" event is recorded for "task-abc" with value "src/auth/login.go"

  Scenario: An idempotent add does not record a duplicate event
    Given a task "task-abc" with reference "src/auth/login.go"
    When the developer runs `tl refine task-abc --add-ref src/auth/login.go`
    Then no "reference_added" event is recorded for "task-abc" in this invocation
