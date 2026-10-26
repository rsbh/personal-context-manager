Feature: Server serves the UI
  As a user I want the UI and the API on one port.

  Scenario: Root serves HTML
    Given the server is running
    When I GET "/"
    Then the status is 200
    And the header "Content-Type" is "text/html; charset=utf-8"

  Scenario: Unknown UI path falls back to the app
    Given the server is running
    When I GET "/contexts/global/git/commits"
    Then the status is 200
    And the header "Content-Type" is "text/html; charset=utf-8"

  Scenario: Unknown API path is a JSON 404
    Given the server is running
    When I GET "/api/v1/nope"
    Then the status is 404
    And the header "Content-Type" is "application/json"
