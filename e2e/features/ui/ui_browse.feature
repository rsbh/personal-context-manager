@ui
Feature: Browse contexts
  As a user I want to browse and read contexts in the web UI.

  Scenario: Sidebar shows contexts under global sections
    Given the server is running
    And a global section "git" with a context "commits" and content "x"
    When I open the UI
    Then the sidebar shows "commits" under "git"

  Scenario: Open a context from the sidebar
    Given the server is running
    And a global section "style" with a context "go" and content "Use **gofmt**." and description "Go style"
    When I open the UI
    And I click "go" in the sidebar
    Then the URL path is "/contexts/global/style/go"
    And I see "Go style"
    And I see "gofmt" in bold
    And I see the raw file "global/style/go.md"

  Scenario: Deep link to a context
    Given the server is running
    And a global section "git" with a context "commits" and content "Write short messages."
    When I open the UI at "/contexts/global/git/commits"
    Then I see "Write short messages."

  Scenario: Missing context
    Given the server is running
    When I open the UI at "/contexts/global/git/none"
    Then I see "Context not found"

  Scenario: Project contexts after scope select
    Given the server is running
    And a project "api" section "review" with a context "pr" and content "y"
    When I open the UI
    And I select "api" in "Scope"
    Then the sidebar shows "pr" under "review"

  Scenario: Current contexts table
    Given the server is running
    And a global section "style" with a context "go" and content "global"
    And a project "api" section "style" with a context "go" and content "project"
    When I open the UI
    And I click the "Current contexts" link
    And I select "api" in "Project"
    Then the table has 1 row with "go"
    And the row with "go" contains "project"

  Scenario: Wide table does not widen the page
    Given the server is running
    And a project "api" section "style" with a context "go" and content "project"
    And the window is 900px wide
    When I open the UI at "/current?project=api"
    Then the page does not scroll sideways
