@ui
Feature: Search palette
  As a user I want to press Cmd+K and find a context fast.

  Scenario: Open the palette with the keyboard
    Given the server is running
    When I open the UI
    And I press "Control+k"
    Then I see the dialog "Search contexts"

  Scenario: Open the palette from the header
    Given the server is running
    When I open the UI
    And I click the "Search" button
    Then I see the dialog "Search contexts"

  Scenario: Find and open a context
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    And a global section "style" with a context "py" and content "Use black."
    When I open the UI
    And I press "Control+k"
    And I type "black" in "Search contexts"
    Then the palette shows "py"
    And the palette does not show "go"
    When I click "py" in the palette
    Then the URL path is "/contexts/global/style/py"
    And I do not see the dialog "Search contexts"

  Scenario: No results
    Given the server is running
    When I open the UI
    And I press "Control+k"
    And I type "zzz" in "Search contexts"
    Then the palette shows "No results"

  Scenario: Close with Escape
    Given the server is running
    When I open the UI
    And I press "Control+k"
    And I press "Escape"
    Then I do not see the dialog "Search contexts"
