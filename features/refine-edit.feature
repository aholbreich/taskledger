@implemented
Feature: Refine a task in an editor
  As a developer or agent
  I want to update a task's editable fields in my system editor
  So that larger refinements are easier than passing many CLI flags

  Background:
    Given an initialized task ledger repository

  Scenario: Refining in editor mode updates fields from the saved buffer
    Given a task "task-abc123" titled "Add login form"
    And "task-abc123" has a description "Build the first version."
    And the developer's editor will save:
      """
      title: Add login form validation
      priority: high
      type: feature

      ## Description

      Validate email format and require a password.
      """
    When the developer runs `tl refine task-abc123 --edit`
    Then "task-abc123" has title "Add login form validation"
    And "task-abc123" has priority "high"
    And "task-abc123" has type "feature"
    And "task-abc123" has the description "Validate email format and require a password."
    And an event "refined" is recorded for "task-abc123"

  Scenario: Refining in editor mode with no changes is a successful no-op
    Given a task "task-abc123" titled "Add login form"
    And "task-abc123" has a description "Build the first version."
    And the developer's editor saves the buffer unchanged
    When the developer runs `tl refine task-abc123 --edit`
    Then the command exits with code 0
    And "task-abc123" has title "Add login form"
    And "task-abc123" has the description "Build the first version."
    And no event "refined" is recorded for "task-abc123"

  Scenario: Refining in editor mode validates edited priority
    Given a task "task-abc123" titled "Add login form"
    And the developer's editor will save:
      """
      title: Add login form validation
      priority: urgent
      type: feature

      ## Description

      Validate email format.
      """
    When the developer runs `tl refine task-abc123 --edit`
    Then the command exits with code 2
    And the output reports that the priority is invalid
    And "task-abc123" has title "Add login form"

  Scenario: Refining in editor mode requires an editor
    Given a task "task-abc123" exists
    And no system editor is configured
    When the developer runs `tl refine task-abc123 --edit`
    Then the command exits with code 2
    And the output reports that no editor is configured
