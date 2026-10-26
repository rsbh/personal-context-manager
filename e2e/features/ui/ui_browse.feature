@ui
Feature: Browse contexts
  As a user I want to browse, search and read contexts in the web UI.

  Scenario: Tree shows contexts under global sections
    Given the server is running
    And a global section "git" with a context "commits" and content "x"
    When I open the UI
    Then the sidebar shows "commits" under "Global"

  Scenario: Open a context from the tree
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

  Scenario: Project contexts after expand
    Given the server is running
    And a project "api" section "review" with a context "pr" and content "y"
    When I open the UI
    And I click "api" in the sidebar
    Then the sidebar shows "pr" under "Projects"

  Scenario: Search contexts
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    And a global section "style" with a context "py" and content "Use black."
    When I open the UI
    And I type "black" in "Search contexts"
    Then the search results show "py"
    And the search results do not show "go"
    When I click "py" in the search results
    Then the URL path is "/contexts/global/style/py"

  Scenario: Current contexts table
    Given the server is running
    And a global section "style" with a context "go" and content "global"
    And a project "api" section "style" with a context "go" and content "project"
    When I open the UI
    And I click the "Current contexts" link
    And I select "api" in "Project"
    Then the table has 1 row with "go"
    And the row with "go" contains "project"
