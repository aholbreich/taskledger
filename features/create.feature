@implemented
Feature: Create a task
  As a developer or agent
  I want to add a new task to the ledger
  So that the work item is visible, addressable, and tracked

  Background:
    Given an initialized TaskLedger repository

  Scenario: Creating a task records it with default status open
    Given no tasks exist
    When the developer runs `tl create "Add login form validation"`
    Then a new task with title "Add login form validation" exists
    And the new task has status "open"
    And the new task has no dependencies
    And an event "created" is recorded for the new task

  # Shortcut version for humans
  Scenario: Creating a task records without "create" or "add"
    Given no tasks exist
    When the developer runs `tl "Shortcut syntax"`
    Then a new task with title "Shortcut syntax" exists
    And the new task has status "open"
    And the new task has no dependencies
    And an event "created" is recorded for the new task
  
  Scenario: add is a synonym for create
    Given no tasks exist
    When the developer runs `tl add "Add login form validation"`
    Then a new task with title "Add login form validation" exists
    And the new task has status "open"
    And the new task has no dependencies
    And an event "created" is recorded for the new task

  Scenario: Creating a task applies type, priority, and tags
    When the developer runs `tl create "Refactor auth error messages" -t chore --priority low --tag frontend --tag auth`
    Then a new task with title "Refactor auth error messages" exists
    And the new task has type "chore"
    And the new task has priority "low"
    And the new task has tags "frontend" and "auth"

  Scenario: Creating a task with JSON output returns the new task object
    When the developer runs `tl create "Add login form validation" --json`
    Then the JSON output contains the new task identifier
    And the JSON output contains title "Add login form validation"
    And the JSON output contains status "open"

  Scenario: Creating a task with forced color highlights the new task identifier
    When the developer runs `tl --color=always create "Add login form validation"`
    Then the output colorizes the new task identifier

  Scenario: Creating a task with JSON output does not emit color
    When the developer runs `tl --color=always create "Add login form validation" --json`
    Then the output does not contain ANSI color
    And the JSON output contains the new task identifier

  Scenario: Creating a task records in it shortest version
    Given no tasks exist
    When the developer runs `tl create "Add login form validation"`
    Then a new task with title "Add login form validation" exists
    And the new task has status "open"
    And the new task has no dependencies
    And an event "created" is recorded for the new task

  Scenario: Creating a task with --title flag is equivalent to the positional form
    Given no tasks exist
    When the developer runs `tl create --title "Add login form validation"`
    Then a new task with title "Add login form validation" exists
    And the new task has status "open"
    And the new task has no dependencies
    And an event "created" is recorded for the new task

  Scenario: Creating a task records in short title version and priority
    Given no tasks exist
    When the developer runs `tl create "Add login form validation" --priority low`
    Then a new task with title "Add login form validation" exists
    And the new task has status "open"
    And the new task has no dependencies
    And an event "created" is recorded for the new task
    And the new task has priority "low"

  Scenario: Creating a task with a long-form description stores it under Description
    Given no tasks exist
    When the developer runs `tl create "Add login form validation" --description "Validate email format and require a password."`
    Then a new task with title "Add login form validation" exists
    And the new task description is "Validate email format and require a password."

  Scenario: Creating a task with the -d short flag stores the description
    Given no tasks exist
    When the developer runs `tl create "Add login telemetry" -d "Capture login attempts metric."`
    Then a new task with title "Add login telemetry" exists
    And the new task description is "Capture login attempts metric."

  Scenario: Creating a task with a description exposes the body in JSON
    When the developer runs `tl create "Add login form validation" -d "Validate email format." --json`
    Then the JSON output contains the new task identifier
    And the JSON output body contains "Validate email format."

  Scenario: Creating a task with short priority aliases stores the canonical form
    Given no tasks exist
    When the developer runs `tl create "short priority" --priority h`
    Then a new task with title "short priority" exists
    And the new task has priority "high"

  Scenario: Creating a task with an invalid priority is rejected
    When the developer runs `tl create "Bad priority" --priority med`
    Then the command exits with code 2
    And the output reports that the priority is invalid


