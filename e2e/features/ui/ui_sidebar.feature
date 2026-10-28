@ui
Feature: Sidebar tree
  As a user I want a clear, consistent sidebar tree.

  Scenario: Collapse and expand a section
    Given the server is running
    And a global section "style" with a context "tabs" and content "x"
    When I open the UI
    Then the sidebar shows "tabs" under "style"
    When I click the "style" section heading
    Then the sidebar does not show "tabs"
    When I click the "style" section heading
    Then the sidebar shows "tabs" under "style"

  Scenario: Section heading is larger than its items
    Given the server is running
    And a global section "style" with a context "tabs" and content "x"
    When I open the UI
    Then the "style" heading font size is larger than the "tabs" item

  Scenario: Rows have the same height and spacing
    Given the server is running
    And a global section "style" with a context "tabs" and content "x"
    And a global section "style" with a context "naming" and content "y"
    When I open the UI
    Then all item rows have the same height
    And the gaps between rows are equal

  Scenario: Heading and children have the same spacing
    Given the server is running
    And a global section "style" with a context "tabs" and content "x"
    And a global section "style" with a context "naming" and content "y"
    When I open the UI
    Then the gap below the "style" heading equals the gaps between rows

  Scenario: Action buttons are easy to hit
    Given the server is running
    And a global section "style"
    When I open the UI
    Then the "New context in style" button is at least 24px wide
