@ui
Feature: Sidebar
  As a user I want a flat sidebar: one scope at a time, its sections and contexts.

  Scenario: Scope select shows one scope at a time
    Given the server is running
    And a global section "style" with a context "go" and content "x"
    And a project "api" section "review" with a context "pr" and content "y"
    When I open the UI
    Then the "Scope" select shows "Global"
    And the sidebar shows "go" under "style"
    And the sidebar does not show "pr"
    When I select "api" in "Scope"
    Then the sidebar shows "pr" under "review"
    And the sidebar does not show "go"

  Scenario: Overview link opens the scope page
    Given the server is running
    When I open the UI
    And I click "Global overview" in the sidebar
    Then the URL path is "/global"

  Scenario: Opening a context selects its scope
    Given the server is running
    And a project "api" section "review" with a context "pr" and content "y"
    When I open the UI at "/contexts/projects/api/review/pr"
    Then the "Scope" select shows "api"
    And the sidebar shows "pr" under "review"

  Scenario: Create a project from the scope select
    Given the server is running
    When I open the UI
    And I select "New project…" in "Scope"
    Then I see the dialog "New project"
    When I type "web" in "Name"
    And I click the "Create" button in the dialog
    Then I see the toast "Project created"
    And the URL path is "/projects/web"

  Scenario: Context row menu
    Given the server is running
    And a global section "style" with a context "go" and content "x"
    When I open the UI
    And I open the sidebar menu of context "go"
    Then I see the menu items "Edit, Delete"

  Scenario: Section row menu
    Given the server is running
    And a global section "style" with a context "go" and content "x"
    When I open the UI
    And I open the sidebar menu of section "style"
    Then I see the menu items "New context, Delete section"

  Scenario: Edit from the context row menu
    Given the server is running
    And a global section "style" with a context "go" and content "x"
    When I open the UI
    And I open the sidebar menu of context "go"
    And I click "Edit" in the menu
    Then the URL path is "/edit/global/style/go"

  Scenario: Rows are 40px high
    Given the server is running
    And a global section "style" with a context "go" and content "x"
    And a global section "style" with a context "naming" and content "y"
    When I open the UI
    Then all item rows are 40px high

  Scenario: Sidebar items stay at the top
    Given the server is running
    And a global section "style" with a context "go" and content "x"
    When I open the UI
    Then the "Global overview" sidebar link is 40px high
    And the sidebar shows the section "style" right below the overview link
