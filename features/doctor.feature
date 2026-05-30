@implemented
Feature: Diagnose ledger integrity
  As a developer or agent
  I want to scan the ledger for structural and data integrity issues
  So that problems like broken dependencies, malformed files, and stale state are found (and optionally fixed)

  # Each diagnostic has a category and a severity:
  #   error   — breaks commands or corrupts data; needs attention
  #   warning — recoverable or cosmetic; informational but not blocking
  # `tl doctor` always exits 0 regardless of findings (it's diagnostic, not a test runner).
  # Non-zero exit means doctor itself failed (e.g. could not read the ledger).

  Background:
    Given an initialized task ledger repository

  Scenario: A clean ledger reports zero issues
    Given a task "task-abc" with title "Working task" and status "open"
    When the developer runs `tl doctor`
    Then the doctor reports no issues
    And the command exits with code 0

  Scenario: doctor exits 0 even when issues are found
    Given a task "task-bad" with status "invalid-status"
    When the developer runs `tl doctor`
    Then the command exits with code 0

  # -------------------------------------------------------------------------
  # Frontmatter — bad data inside the task's YAML header.
  # -------------------------------------------------------------------------
  Scenario Outline: Frontmatter problems are detected (errors)
    Given <setup>
    When the developer runs `tl doctor`
    Then the doctor reports a "frontmatter" issue for "<task>" with severity "error"

    Examples:
      | setup                                                          | task     |
      | a task file "task-bad.md" whose content is not valid YAML       | task-bad |
      | a task file "task-xyz.md" with frontmatter missing the "title" field | task-xyz |
      | a task "task-abc" with status "super-duper"                     | task-abc |
      | a task "task-abc" with priority "urgent"                        | task-abc |

  Scenario Outline: An empty type field is a fixable warning
    Given <setup>
    When the developer runs `tl doctor`
    Then the doctor reports a "frontmatter" issue for "<task>" with severity "warning"

    Examples:
      | setup                                | task     |
      | a task "task-abc" with type "" (empty) | task-abc |

  # -------------------------------------------------------------------------
  # Identity — two task files claiming the same ID.
  # -------------------------------------------------------------------------
  Scenario: Duplicate task IDs are detected
    Given a task "task-dup" exists
    And a second task file claiming id "task-dup" exists in the tasks directory
    When the developer runs `tl doctor`
    Then the doctor reports an "identity" issue for "task-dup" with severity "error"

  # -------------------------------------------------------------------------
  # Dependencies — broken or self-referential depends_on edges.
  # -------------------------------------------------------------------------
  Scenario Outline: Dependency problems are detected
    Given <setup>
    When the developer runs `tl doctor`
    Then the doctor reports a "dependency" issue for "<task>" with severity "error"

    Examples:
      | setup                                                  | task     |
      | a task "task-abc" with dependency "task-nonexistent"    | task-abc |
      | a task "task-abc" with dependency "task-abc"            | task-abc |

  Scenario: Cyclic dependencies are detected and reported for every task in the cycle
    Given a task "task-a" with dependency "task-b"
    And a task "task-b" with dependency "task-a"
    When the developer runs `tl doctor`
    Then the doctor reports a "dependency" issue for "task-a" with severity "error"
    And the doctor reports a "dependency" issue for "task-b" with severity "error"

  # -------------------------------------------------------------------------
  # Events — journal references a task that does not exist.
  # Not detected: tasks with no events. Pre-journal tasks and imported tasks
  # produce false positives, so the absence of events is not flagged.
  # -------------------------------------------------------------------------
  Scenario: Orphaned events referencing a nonexistent task are detected
    Given an event in the journal referencing task "task-ghost"
    And no task file for "task-ghost" exists
    When the developer runs `tl doctor`
    Then the doctor reports an "events" issue for "task-ghost" with severity "warning"

  # -------------------------------------------------------------------------
  # Claims — claim state inconsistent with task status. Severity varies:
  #   error   — status/claim disagree such that commands will misbehave
  #   warning — recoverable (stale claim, leftover claim data on open task)
  # -------------------------------------------------------------------------
  Scenario: A task marked in_progress with no claim data is an error
    Given a task "task-abc" with status "in_progress" and no active claim
    When the developer runs `tl doctor`
    Then the doctor reports a "claims" issue for "task-abc" with severity "error"

  Scenario: An expired claim that was never released is a warning
    Given a task "task-abc" claimed by "claude-code:main" with an expired lease
    When the developer runs `tl doctor`
    Then the doctor reports a "claims" issue for "task-abc" with severity "warning"

  Scenario: An open task still carrying claim data is a warning
    Given a task "task-abc" with status "open" that still has claim data set
    When the developer runs `tl doctor`
    Then the doctor reports a "claims" issue for "task-abc" with severity "warning"

  # -------------------------------------------------------------------------
  # Timestamps — out-of-order or future-dated timestamps. All warnings:
  # they signal data weirdness but don't break command behaviour.
  # -------------------------------------------------------------------------
  Scenario Outline: Timestamp anomalies are detected
    Given <setup>
    When the developer runs `tl doctor`
    Then the doctor reports a "timestamps" issue for "task-abc" with severity "warning"

    Examples:
      | setup                                                          |
      | a task "task-abc" with created_at in the year 2099              |
      | a task "task-abc" with created_at after updated_at              |
      | a task "task-abc" with claim_expires_at before claim_claimed_at |

  # -------------------------------------------------------------------------
  # Filesystem — orphaned .tmp files (warning, recoverable) and unreadable
  # task files (error, data lost).
  # -------------------------------------------------------------------------
  Scenario: An orphaned .tmp file is a warning
    Given an orphaned file "task-abc.md.tmp" exists in the tasks directory
    When the developer runs `tl doctor`
    Then the doctor reports a "filesystem" issue with severity "warning"

  Scenario: A task file that cannot be read is an error
    Given a task file "task-corrupted.md" that cannot be read
    When the developer runs `tl doctor`
    Then the doctor reports a "filesystem" issue for "task-corrupted" with severity "error"

  # -------------------------------------------------------------------------
  # Body — merge-conflict markers (error: next write will overwrite the
  # conflict) and malformed notes lines (warning: display only).
  # -------------------------------------------------------------------------
  Scenario Outline: Merge conflict markers in the body are detected
    Given a task "task-abc" whose body contains "<marker>"
    When the developer runs `tl doctor`
    Then the doctor reports a "body" issue for "task-abc" with severity "error"

    Examples:
      | marker              |
      | <<<<<<< HEAD        |
      | =======             |
      | >>>>>>> branch-name |

  Scenario: Malformed notes lines in the body are a warning
    Given a task "task-abc" whose Notes section contains lines that do not match the canonical format
    When the developer runs `tl doctor`
    Then the doctor reports a "body" issue for "task-abc" with severity "warning"

  # -------------------------------------------------------------------------
  # Config — config.yaml problems break command loading.
  # -------------------------------------------------------------------------
  Scenario Outline: Config problems are detected
    Given <setup>
    When the developer runs `tl doctor`
    Then the doctor reports a "config" issue with severity "error"

    Examples:
      | setup                                                  |
      | the ledger has no config.yaml file                      |
      | the config.yaml contains content that is not valid YAML |

  # -------------------------------------------------------------------------
  # Scale — informational warnings when the ledger grows large. The
  # thresholds (>100 tasks, >1000 events) reflect where filesystem scans
  # and journal reads start to become noticeable on slower disks. Tune as
  # the tool's performance characteristics evolve.
  # -------------------------------------------------------------------------
  Scenario Outline: Scale warnings are emitted for large ledgers
    Given <setup>
    When the developer runs `tl doctor`
    Then the doctor output contains a scale warning about "<dimension>" with severity "warning"

    Examples:
      | setup                        | dimension |
      | a ledger with 101 tasks      | tasks     |
      | a ledger with 1001 events    | events    |

  # -------------------------------------------------------------------------
  # JSON output — diagnostics serialise as an array of objects with a
  # stable shape so consumers can pipe to jq.
  # -------------------------------------------------------------------------
  Scenario: A clean ledger produces an empty JSON array
    Given a task "task-abc" with title "Working task" and status "open"
    When the developer runs `tl doctor --json`
    Then the JSON output is an empty array

  Scenario: JSON diagnostics carry severity, category, message, and fixable fields
    Given a task "task-bad" with status "invalid-status"
    When the developer runs `tl doctor --json`
    Then the JSON output is an array of diagnostic objects
    And a diagnostic object has category "frontmatter" and severity "error"
    And a diagnostic object has a non-empty "message" field
    And a diagnostic object has a "fixable" boolean field

  # -------------------------------------------------------------------------
  # --fix — opt-in auto-repair for the specific issues we know how to mend.
  # Anything else is reported as `fixable: false` and left alone.
  # -------------------------------------------------------------------------
  Scenario: --fix repairs a self-referencing dependency
    Given a task "task-abc" with dependency "task-abc"
    When the developer runs `tl doctor --fix`
    Then the doctor reports the "dependency" issue for "task-abc" as fixed
    And "task-abc" no longer depends on "task-abc"

  Scenario: --fix removes orphaned .tmp files
    Given an orphaned file "task-abc.md.tmp" exists in the tasks directory
    When the developer runs `tl doctor --fix`
    Then the doctor reports the orphaned file as removed
    And the file "task-abc.md.tmp" no longer exists

  Scenario: --fix clears leftover claim data from an open task
    Given a task "task-abc" with status "open" that still has claim data set
    When the developer runs `tl doctor --fix`
    Then the doctor reports the claim data as cleared
    And "task-abc" has no claim data

  Scenario: --fix releases an expired claim and returns the task to open
    # Lock-protected: another actor reclaiming after this `--fix` runs the
    # normal claim path and gets a fresh lease — no data race possible.
    Given a task "task-abc" claimed by "claude-code:main" with an expired lease
    When the developer runs `tl doctor --fix`
    Then the doctor reports the stale claim as released
    And "task-abc" has status "open"

  Scenario: --fix does not attempt to repair unfixable issues
    Given a task file "task-bad.md" whose content is not valid YAML
    When the developer runs `tl doctor --fix`
    Then the doctor reports the "frontmatter" issue for "task-bad" as not fixable
