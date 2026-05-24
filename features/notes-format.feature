@implemented
Feature: Notes section format
  As an agent consuming task JSON and Markdown
  I want task notes to use one predictable format
  So that progress and lifecycle context are easy to parse

  Background:
    Given an initialized task ledger repository

  Scenario: Appending a note writes a canonical bullet
    Given a task "task-abc123" titled "Add login form validation"
    When the agent runs `tl note task-abc123 --actor claude-code:frontend --message "Verified locally."`
    Then "task-abc123" has a canonical "note" note from "claude-code:frontend" with message "Verified locally."

  Scenario: Lifecycle commands write canonical note kinds
    Given a task "task-abc123" with status "open"
    When the developer runs `tl block task-abc123 --actor human --message "Waiting for API decision."`
    Then "task-abc123" has a canonical "blocked" note from "human" with message "Waiting for API decision."

  Scenario: Task JSON exposes parsed notes
    Given a task "task-abc123" titled "Add login form validation"
    When the agent runs `tl note task-abc123 --actor claude-code:frontend --message "Verified locally."`
    And the developer asks for JSON for "task-abc123"
    Then the JSON output contains a parsed "note" note from "claude-code:frontend" with message "Verified locally."
