@ui
Feature: Sections and projects
  As a user I want to create and delete sections and projects in the web UI.

  Scenario: Create a section that exists
    Given the server is running
    And a global section "style"
    When I open the UI at "/global"
    And I click the "New section" button in the overview
    And I type "style" in "Name"
    And I click the "Create" button in the dialog
    Then I see the toast "already exists"
    And I see the dialog "New section"
    And the "Name" field is "style"

  Scenario: Create a project and a section in it
    Given the server is running
    When I open the UI
    And I select "New project…" in "Scope"
    And I type "api" in "Name"
    And I click the "Create" button in the dialog
    Then I see the toast "Project created"
    When I click the "New section" button in the overview
    And I type "review" in "Name"
    And I click the "Create" button in the dialog
    Then I see the toast "Section created"
    And the sidebar shows the section "review"

  Scenario: Delete an empty section from the sidebar
    Given the server is running
    And a global section "style"
    When I open the UI
    And I open the sidebar menu of section "style"
    And I click "Delete section" in the menu
    Then I see the dialog "Delete section?"
    When I click the "Delete" button in the dialog
    Then I see the toast "Deleted"
    And the sidebar does not show the section "style"

  Scenario: Delete a context from the sidebar
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI
    And I open the sidebar menu of context "go"
    And I click "Delete" in the menu
    Then I see the dialog "Delete context?"
    When I click the "Delete" button in the dialog
    Then I see the toast "Deleted"
    And the sidebar does not show "go"

  Scenario: Delete the section of the open context
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/contexts/global/style/go"
    And I open the sidebar menu of section "style"
    And I click "Delete section" in the menu
    And I click the "Delete" button in the dialog
    And I click the "Delete all" button in the dialog
    Then the URL path is "/"
