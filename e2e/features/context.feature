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

  Scenario: Missing section flag
    When I run "pcm create context go -g -m x"
    Then the exit code is 1
    And the output contains "error: missing section: use -s <section>"

  Scenario: Missing content
    When I run "pcm create context go -s style -g"
    Then the exit code is 1
    And the output contains "error: missing content: use -m <text>"

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
