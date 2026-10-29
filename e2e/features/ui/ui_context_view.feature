@ui
Feature: Context view
  As a user I want a clear context page with a way back to its scope.

  Scenario: Breadcrumb links to the project page
    Given the server is running
    And a project "api" section "review" with a context "pr" and content "Small PRs."
    When I open the UI at "/contexts/projects/api/review/pr"
    Then the breadcrumb shows "api / review"
    When I click "api" in the breadcrumb
    Then the URL path is "/projects/api"

  Scenario: Breadcrumb links to the global page
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/contexts/global/style/go"
    Then the breadcrumb shows "Global / style"
    When I click "Global" in the breadcrumb
    Then the URL path is "/global"

  Scenario: Edit and delete are icon buttons
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/contexts/global/style/go"
    Then the "Edit" button has no text
    And the "Delete" button has no text
    And the "Delete" button on the page is danger colour

  Scenario: Title block
    Given the server is running
    And a global section "style" with a context "go" and content "x" and description "Go style"
    When I open the UI at "/contexts/global/style/go"
    Then the context title is at least 26px and bold
    And the gap between the title and "Go style" is at most 8px
