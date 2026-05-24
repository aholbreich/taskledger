@implemented
Feature: Compact bulk JSON task output
  As an agent listing candidate work
  I want bulk JSON output to omit raw Markdown bodies
  So that queue responses stay compact while preserving parsed task context

  Background:
    Given an initialized task ledger repository

  Scenario: List JSON omits body and includes parsed fields
    Given a task "task-abc123" titled "Add login form validation"
    And "task-abc123" has a description "Validate login inputs."
    When the agent runs `tl note task-abc123 --actor claude-code:frontend --message "Verified locally."`
    And the developer asks for list JSON
    Then the JSON task "task-abc123" does not include field "body"
    And the JSON task "task-abc123" has description "Validate login inputs."
    And the JSON task "task-abc123" contains a parsed "note" note from "claude-code:frontend" with message "Verified locally."

  Scenario: Ready JSON omits body and includes parsed fields
    Given a task "task-abc123" titled "Add login form validation"
    And "task-abc123" has a description "Validate login inputs."
    When the agent runs `tl note task-abc123 --actor claude-code:frontend --message "Verified locally."`
    And the developer asks for ready JSON
    Then the JSON task "task-abc123" does not include field "body"
    And the JSON task "task-abc123" has description "Validate login inputs."
    And the JSON task "task-abc123" contains a parsed "note" note from "claude-code:frontend" with message "Verified locally."
