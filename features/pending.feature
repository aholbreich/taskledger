@implemented
Feature: Mark a task as pending human input
  As an agent that needs a human decision
  I want to put a task into a waiting state with a question
  So that work stops cleanly and the human knows what is being asked

  Background:
    Given an initialized TaskLedger repository

  Scenario: An agent marks a task pending_human with a question
    Given a task "task-abc123" claimed by "claude-code:frontend" with an active lease
    When the agent runs `tl pending task-abc123 --actor claude-code:frontend --question "Which auth provider first?"`
    Then "task-abc123" has status "pending_human"
    And "task-abc123" records the question "Which auth provider first?"
    And "task-abc123" records the requester "claude-code:frontend"
    And an event "pending_requested" is recorded for "task-abc123"

  Scenario: A pending_human task is removed from the ready queue
    Given a task "task-abc123" with status "pending_human"
    When the agent runs `tl ready`
    Then the ready output does not contain "task-abc123"

  Scenario: Marking a task pending_human releases the active claim
    Given a task "task-abc123" claimed by "claude-code:frontend" with an active lease
    When the agent runs `tl pending task-abc123 --actor claude-code:frontend --question "Which auth provider first?"`
    Then "task-abc123" has status "pending_human"
    And "task-abc123" is not claimed
    And "task-abc123" records the requester "claude-code:frontend"
