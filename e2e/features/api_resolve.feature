Feature: Resolve API formats
  As an API client I want merged contexts as JSON or markdown.

  Background:
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt." and description "Go style"

  Scenario: JSON by default
    When I GET "/api/v1/resolve"
    Then the status is 200
    And the header "Content-Type" is "application/json"
    And the body contains "\"id\":\"global/style/go\""

  Scenario: Markdown with format param
    When I GET "/api/v1/resolve?format=md"
    Then the status is 200
    And the header "Content-Type" is "text/markdown; charset=utf-8"
    And the body contains "## style/go (global)"

  Scenario: Markdown with Accept header
    When I GET "/api/v1/resolve" with header "Accept: text/markdown"
    Then the status is 200
    And the header "Content-Type" is "text/markdown; charset=utf-8"
    And the body contains "Use gofmt."

  Scenario: Format param wins over Accept
    When I GET "/api/v1/resolve?format=json" with header "Accept: text/markdown"
    Then the header "Content-Type" is "application/json"

  Scenario: Invalid format
    When I GET "/api/v1/resolve?format=xml"
    Then the status is 400
    And the header "Content-Type" is "application/json"
    And the error is "invalid format \"xml\": use json or md"

  Scenario: Empty markdown
    Given the global section "style" has no contexts
    When I GET "/api/v1/resolve?format=md"
    Then the status is 200
    And the body is "No contexts found.\n"
