Feature: Contexts
  As a user I want to manage contexts from the CLI.

  Scenario: Create global context
    Given the server is running
    And a global section "style"
    When I run "pcm create context go -s style -g -m 'Use gofmt.' --title Go --description 'Go style' --tag go --tag fmt"
    Then the exit code is 0
    And the output contains "context/go created in global (section \"style\")"

  Scenario: Create context in current project
    Given the server is running
    And I am in a git repo named "api"
    And a project "api" with a section "style"
    When I run "pcm create context go -s style -m 'Use gofmt.'"
    Then the exit code is 0
    And the output contains "context/go created in project \"api\" (section \"style\")"

  Scenario: Create context in missing section
    Given the server is running
    When I run "pcm create context go -s none -g -m x"
    Then the exit code is 1
    And the output contains "not found"

  Scenario: Create existing context
    Given the server is running
    And a global section "style" with a context "go"
    When I run "pcm create context go -s style -g -m x"
    Then the exit code is 1
    And the output contains "already exists"

  Scenario: Force overwrites existing context
    Given the server is running
    And a global section "style" with a context "go" and content "Old rule." and description "Old desc"
    When I run "pcm create context go -s style -g -m 'New rule.' --title Go --force"
    Then the exit code is 0
    And the output contains "context/go updated in global (section \"style\")"
    When I run "pcm get context go -s style -g"
    Then the output contains "title: Go"
    And the output contains "New rule."
    And the output does not contain "Old rule."
    And the output does not contain "Old desc"

  Scenario: Force creates missing context
    Given the server is running
    And a global section "style"
    When I run "pcm create context go -s style -g -m x --force"
    Then the exit code is 0
    And the output contains "context/go created in global (section \"style\")"

  Scenario: Force in missing section
    Given the server is running
    When I run "pcm create context go -s none -g -m x --force"
    Then the exit code is 1
    And the output contains "not found"

  Scenario: Missing section flag
    When I run "pcm create context go -g -m x"
    Then the exit code is 1
    And the output contains "error: missing section: use -s <section>"

  Scenario: Missing content
    When I run "pcm create context go -s style -g"
    Then the exit code is 1
    And the output contains "error: missing content: use -m <text> or -f <file>"

  Scenario: Create context from file
    Given the server is running
    And a global section "style"
    And the file "~/rules.md" has content "Use gofmt."
    When I run "pcm create context go -s style -g -f rules.md"
    Then the exit code is 0
    And the output contains "context/go created in global (section \"style\")"
    When I run "pcm get context go -s style -g"
    Then the output contains "Use gofmt."

  Scenario: Create context from stdin
    Given the server is running
    And a global section "style"
    When I run "pcm create context go -s style -g -f -" with stdin:
      """
      # Go
      - Use gofmt.
      - Wrap errors.
      """
    Then the exit code is 0
    When I run "pcm get context go -s style -g"
    Then the output contains "- Use gofmt."
    And the output contains "- Wrap errors."

  Scenario: Create context from missing file
    Given the server is running
    And a global section "style"
    When I run "pcm create context go -s style -g -f none.md"
    Then the exit code is 1
    And the output contains "none.md"

  Scenario: Message and file together
    When I run "pcm create context go -s style -g -m x -f rules.md"
    Then the exit code is 1
    And the output contains "error: use -m or -f, not both"

  Scenario: List contexts
    Given the server is running
    And a global section "style" with a context "go"
    When I run "pcm get contexts -g"
    Then the exit code is 0
    And the output contains "NAME"
    And the output contains "SECTION"
    And the output contains "TITLE"
    And the output contains "DESCRIPTION"
    And the output contains "UPDATED"
    And the output contains "go"

  Scenario: List contexts filtered by section
    Given the server is running
    And a global section "style" with a context "go"
    And a global section "review" with a context "pr"
    When I run "pcm get contexts -g -s review"
    Then the output contains "pr"
    And the output does not contain "go "

  Scenario: List with no contexts
    Given the server is running
    When I run "pcm get contexts -g"
    Then the exit code is 0
    And the output contains "No contexts found."

  Scenario: Get one context
    Given the server is running
    And a global section "style"
    And I run "pcm create context go -s style -g -m 'Use gofmt.' --title Go"
    When I run "pcm get context go -s style -g"
    Then the exit code is 0
    And the output contains "title: Go"
    And the output contains "Use gofmt."

  Scenario: Get missing context
    Given the server is running
    And a global section "style"
    When I run "pcm get context none -s style -g"
    Then the exit code is 1
    And the output contains "not found"

  Scenario: Delete context
    Given the server is running
    And a global section "style" with a context "go"
    When I run "pcm delete context go -s style -g"
    Then the exit code is 0
    And the output contains "context \"go\" deleted from global (section \"style\")"
