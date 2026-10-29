@ui
Feature: UI header
  As a user I want a clear main navigation.

  Scenario: Header is 64px high
    Given the server is running
    When I open the UI
    Then the header is 64px high

  Scenario: Navigation tabs are easy to hit
    Given the server is running
    When I open the UI
    Then each main navigation link is at least 32px high
    And the gap between main navigation links is at least 4px

  Scenario: Active tab is marked
    Given the server is running
    When I open the UI
    Then the "Browse" navigation link is the current page
    When I click the "Sync" link
    Then the "Sync" navigation link is the current page
    And the "Browse" navigation link is not the current page
