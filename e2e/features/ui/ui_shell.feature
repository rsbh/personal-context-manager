@ui
Feature: UI shell
  As a user I want to open the pcm web UI from the server.

  Scenario: UI loads from pcm-server
    Given the server is running
    When I open the UI
    Then the page title is "pcm"
    And I see "server ok"

  Scenario: Sidebar lists projects and global sections
    Given the server is running
    And a global section "git" with a context "commits" and content "x"
    And a project "api" section "review" with a context "pr" and content "y"
    When I open the UI
    Then the sidebar shows "git" under "Global"
    And the sidebar shows "api" under "Projects"

  Scenario: Deep link loads the app
    Given the server is running
    When I open the UI at "/contexts/global/git/commits"
    Then the page title is "pcm"

  Scenario: Page has no outer margin
    Given the server is running
    When I open the UI
    Then the page body has no margin
