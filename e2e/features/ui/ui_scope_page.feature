@ui
Feature: Scope overview pages
  As a user I want one page per scope to see and manage its sections and contexts.

  Scenario: Global overview lists sections and contexts
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/global"
    Then I see the heading "Global"
    And I see "1 section · 1 context"
    And the overview section "style" lists "go"

  Scenario: Project overview lists sections and contexts
    Given the server is running
    And a project "api" section "review" with a context "pr" and content "Small PRs."
    When I open the UI at "/projects/api"
    Then I see the heading "api"
    And the overview section "review" lists "pr"

  Scenario: Unknown project
    Given the server is running
    When I open the UI at "/projects/nope"
    Then I see "Project not found"

  Scenario: Open a context from the overview
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/global"
    And I click "go" in the overview
    Then the URL path is "/contexts/global/style/go"

  Scenario: Edit a context from the overview
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/global"
    And I click the "Edit go" button in the overview
    Then the URL path is "/edit/global/style/go"

  Scenario: Create a section from the overview
    Given the server is running
    When I open the UI at "/global"
    And I click the "New section" button in the overview
    And I type "style" in "Name"
    And I click the "Create" button in the dialog
    Then I see the toast "Section created"
    And the overview shows the section "style"

  Scenario: New context in a section
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/global"
    And I click the "New context in style" button in the overview
    Then the URL path is "/new"
    And the "Section" select shows "style"

  Scenario: Delete a context from the overview
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/global"
    And I click the "Delete go" button in the overview
    Then I see the dialog "Delete context?"
    When I click the "Delete" button in the dialog
    Then I see the toast "Deleted"
    And the overview does not show "go"

  Scenario: Delete a section that is not empty
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/global"
    And I click the "Delete section style" button in the overview
    And I click the "Delete" button in the dialog
    And I click the "Delete all" button in the dialog
    Then I see the toast "Deleted"
    And the overview does not show the section "style"

  Scenario: Delete a project
    Given the server is running
    And a project "api" with a section "review"
    When I open the UI at "/projects/api"
    And I click the "Delete project" button in the overview
    And I click the "Delete" button in the dialog
    And I click the "Delete all" button in the dialog
    Then I see the toast "Deleted"
    And the URL path is "/"

  Scenario: Action icons have colour
    Given the server is running
    And a project "api" with a section "review"
    When I open the UI at "/projects/api"
    Then the "New section" button in the overview is success colour
    And the "Delete project" button in the overview is danger colour
    And the "Delete section review" button in the overview is danger colour
