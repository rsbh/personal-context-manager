Feature: Current contexts table
  As a user I want to see all contexts that apply in the current repo.

  Scenario: Global first, then detected project
    Given the server is running
    And a global section "git" with a context "commits" and content "x"
    And I am in a git repo named "api"
    And a project "api" section "review" with a context "pr" and content "y"
    When I run "pcm get contexts"
    Then the exit code is 0
    And the output contains "project: api (from git repo)"
    And the line with "commits" comes before the line with "pr"
    And the line with "commits" contains "global"
    And the line with "pr" contains "project"

  Scenario: Project overrides global
    Given the server is running
    And a global section "style" with a context "go" and content "global rule"
    And I am in a git repo named "api"
    And a project "api" section "style" with a context "go" and content "project rule"
    When I run "pcm get contexts"
    Then the output contains one line with "go"
    And the line with "go" contains "project"

  Scenario: Not in a git repo
    Given the server is running
    And a global section "git" with a context "commits" and content "x"
    And I am not in a git repo
    When I run "pcm get contexts"
    Then the exit code is 0
    And the output contains "project: none (not in a git repo)"
    And the output contains "commits"

  Scenario: Global only with -g
    Given the server is running
    And a global section "git" with a context "commits" and content "x"
    And I am in a git repo named "api"
    And a project "api" section "review" with a context "pr" and content "y"
    When I run "pcm get contexts -g"
    Then the output contains "commits"
    And the output does not contain "pr "
    And the output does not contain "project:"
