Feature: Update a task
  As a developer or agent
  I want to change a task's editable fields after it is created
  So that I can correct mistakes without editing task files by hand

  Background:
    Given an initialized TaskLedger repository

  Scenario: Updating the title replaces it and records an event
    Given a task "task-abc123" titled "Add login form valdiation"
    When the developer runs `tl update task-abc123 --title "Add login form validation"`
    Then "task-abc123" has title "Add login form validation"
    And an event "updated" is recorded for "task-abc123"

  Scenario: Updating the description replaces the stored description
    Given a task "task-abc123" exists
    When the developer runs `tl update task-abc123 --description "Validate email format and require a password."`
    Then "task-abc123" has the description "Validate email format and require a password."

  Scenario: Updating the priority changes the task priority
    Given a task "task-abc123" with status "open" and priority "low"
    When the developer runs `tl update task-abc123 --priority high`
    Then "task-abc123" has priority "high"

  Scenario: Adding a tag leaves existing tags in place
    Given a task "task-abc123" with tags "frontend"
    When the developer runs `tl update task-abc123 --add-tag auth`
    Then "task-abc123" has tags "frontend" and "auth"

  Scenario: Removing a tag drops only that tag
    Given a task "task-abc123" with tags "frontend" and "auth"
    When the developer runs `tl update task-abc123 --remove-tag auth`
    Then "task-abc123" has tags "frontend"
    And "task-abc123" does not have the tag "auth"

  Scenario: Updating one field leaves the others unchanged
    Given a task "task-abc123" titled "Add login form validation" with priority "low"
    When the developer runs `tl update task-abc123 --priority high`
    Then "task-abc123" has priority "high"
    And "task-abc123" has title "Add login form validation"

  Scenario: Updating editable fields does not change the task status
    Given a task "task-abc123" with status "in_progress"
    When the developer runs `tl update task-abc123 --title "Add login form validation"`
    Then "task-abc123" still has status "in_progress"

  Scenario: Updating with JSON output returns the updated task
    Given a task "task-abc123" titled "Add login form validation"
    When the developer runs `tl update task-abc123 --priority high --json`
    Then the JSON output contains title "Add login form validation"
    And the JSON output contains priority "high"

  Scenario: Updating with an invalid priority is rejected
    Given a task "task-abc123" exists
    When the developer runs `tl update task-abc123 --priority med`
    Then the command exits with code 2
    And the output reports that the priority is invalid

  Scenario: Updating a task that does not exist is rejected
    When the developer runs `tl update task-zzz999 --title "Add login form validation"`
    Then the command exits with code 3
    And the output reports that "task-zzz999" was not found

  Scenario: Updating with no editable fields specified is rejected
    Given a task "task-abc123" exists
    When the developer runs `tl update task-abc123`
    Then the command exits with code 2
    And the output reports that no fields were given to update
