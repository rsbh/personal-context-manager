Feature: Merged contexts
  As a coding agent I want all contexts that apply, in one markdown output.

  Scenario: Global and current project
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt." and description "Go style"
    And I am in a git repo named "api"
    And a project "api" section "review" with a context "pr" and content "Check errors."
    When I run "pcm get contexts -o md"
    Then the exit code is 0
    And the output is:
      """
      # Context

      ## style/go (global)
      > Go style

      Use gofmt.

      ## review/pr (project "api")

      Check errors.
      """

  Scenario: Project overrides global
    Given the server is running
    And a global section "style" with a context "go" and content "global rule"
    And a project "api" section "style" with a context "go" and content "project rule"
    When I run "pcm get contexts -o md -p api"
    Then the output contains "project rule"
    And the output does not contain "global rule"

  Scenario: Not in a git repo gives global only
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    And I am not in a git repo
    When I run "pcm get contexts -o md"
    Then the exit code is 0
    And the output contains "## style/go (global)"

  Scenario: Filter by section
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    And a global section "review" with a context "pr" and content "Check errors."
    When I run "pcm get contexts -o md -g -s review"
    Then the output contains "Check errors."
    And the output does not contain "Use gofmt."

  Scenario: Nothing found
    Given the server is running
    When I run "pcm get contexts -o md -g"
    Then the exit code is 0
    And the output contains "No contexts found."

  Scenario: Invalid output format
    When I run "pcm get contexts -o json"
    Then the exit code is 1
    And the output contains "error: invalid output format \"json\": use table or md"
