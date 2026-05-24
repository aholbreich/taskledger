@implemented
Feature: Refine a task
  As a developer or agent
  I want to change a task's editable fields after it is created
  So that I can correct mistakes without editing task files by hand

  Background:
    Given an initialized task ledger repository

  Scenario: Refineing the title replaces it and records an event
    Given a task "task-abc123" titled "Add login form"
    When the agent or developer runs `tl refine task-abc123 --title "Add login form validation"`
    Then "task-abc123" has title "Add login form validation"
    And an event "refined" is recorded for "task-abc123"

  Scenario: Refining the description replaces the stored description
    Given a task "task-abc123" exists
    When the developer runs `tl refine task-abc123 --description "Validate email format and require a password."`
    Then "task-abc123" has the description "Validate email format and require a password."

  Scenario: Refining editable fields does not change the task status
    Given a task "task-abc123" with status "in_progress"
    When the developer runs `tl refine task-abc123 --title "Add login form validation"`
    Then "task-abc123" still has status "in_progress"

  Scenario: Refining a task that does not exist is rejected
    When the developer runs `tl refine task-zzz999 --title "Add login form validation"`
    Then the command exits with code 3
    And the output reports that "task-zzz999" was not found

  Scenario: Refining with no editable fields specified is rejected
    Given a task "task-abc123" exists
    When the developer runs `tl refine task-abc123`
    Then the command exits with code 2
    And the output reports that no fields were given to refine

