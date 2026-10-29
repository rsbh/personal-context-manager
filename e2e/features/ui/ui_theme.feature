@ui
Feature: UI theme
  As a user I want calm colours and readable text.

  Scenario: Indigo accent and slate gray
    Given the server is running
    When I open the UI
    Then the theme accent color is "indigo"
    And the theme gray color is "slate"

  Scenario: Body text is 16px
    Given the server is running
    When I open the UI
    Then the body text size is 16px

  Scenario: Text and background use the gray scale
    Given the server is running
    When I open the UI
    Then the main text color is gray step 12
    And the page background is gray step 1
    And the secondary text color is gray step 11
