@ui
Feature: Sections and projects
  As a user I want to create and delete sections and projects in the web UI.

  Scenario: Create a global section
    Given the server is running
    When I open the UI
    And I click the "New global section" button
    Then I see the dialog "New section"
    When I type "style" in "Name"
    And I click the "Create" button in the dialog
    Then I see the toast "Section created"
    And the sidebar shows "style" under "Global"

  Scenario: Create a project and a section in it
    Given the server is running
    When I open the UI
    And I click the "New project" button
    And I type "api" in "Name"
    And I click the "Create" button in the dialog
    Then I see the toast "Project created"
    And the sidebar shows "api" under "Projects"
    When I click the "New section in api" button
    And I type "review" in "Name"
    And I click the "Create" button in the dialog
    Then I see the toast "Section created"
    And the sidebar shows "review" under "Projects"

  Scenario: Create a section that exists
    Given the server is running
    And a global section "style"
    When I open the UI
    And I click the "New global section" button
    And I type "style" in "Name"
    And I click the "Create" button in the dialog
    Then I see the toast "already exists"
    And I see the dialog "New section"
    And the "Name" field is "style"

  Scenario: Delete an empty section
    Given the server is running
    And a global section "style"
    When I open the UI
    And I click the "Section style actions" button
    And I click "Delete" in the menu
    Then I see the dialog "Delete section?"
    When I click the "Delete" button in the dialog
    Then I see the toast "Deleted"
    And the sidebar does not show "style"

  Scenario: Delete a section that is not empty
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI
    And I click the "Section style actions" button
    And I click "Delete" in the menu
    And I click the "Delete" button in the dialog
    Then I see the dialog "Section is not empty"
    When I click the "Delete all" button in the dialog
    Then I see the toast "Deleted"
    And the sidebar does not show "style"
    And the sidebar does not show "go"

  Scenario: Delete a project that is not empty
    Given the server is running
    And a project "api" with a section "review"
    When I open the UI
    And I click the "Project api actions" button
    And I click "Delete" in the menu
    And I click the "Delete" button in the dialog
    Then I see the dialog "Project is not empty"
    When I click the "Delete all" button in the dialog
    Then I see the toast "Deleted"
    And the sidebar does not show "api"

  Scenario: Delete the section of the open context
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/contexts/global/style/go"
    And I click the "Section style actions" button
    And I click "Delete" in the menu
    And I click the "Delete" button in the dialog
    And I click the "Delete all" button in the dialog
    Then the URL path is "/"
