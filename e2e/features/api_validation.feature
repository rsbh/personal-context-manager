Feature: Request body validation
  As an API client I want clear 400 errors when a request body is not valid.

  Background:
    Given the server is running

  Scenario: Create project without name
    When I POST "/api/v1/projects" with body:
      """
      {}
      """
    Then the status is 400
    And the error is "invalid body: name is required"

  Scenario: Create section with unknown scope
    When I POST "/api/v1/sections" with body:
      """
      {"scope": "team", "name": "style"}
      """
    Then the status is 400
    And the error is "invalid body: scope must be one of: global, project"

  Scenario: Create section reports all missing fields
    When I POST "/api/v1/sections" with body:
      """
      {}
      """
    Then the status is 400
    And the error is "invalid body: scope is required; name is required"

  Scenario: Create section in project scope without project
    When I POST "/api/v1/sections" with body:
      """
      {"scope": "project", "name": "style"}
      """
    Then the status is 400
    And the error is "invalid body: project is required for scope project"

  Scenario: Create section in global scope with project
    When I POST "/api/v1/sections" with body:
      """
      {"scope": "global", "project": "app", "name": "style"}
      """
    Then the status is 400
    And the error is "invalid body: project is not allowed for scope global"

  Scenario: Create context without section and name
    When I POST "/api/v1/contexts" with body:
      """
      {"scope": "global"}
      """
    Then the status is 400
    And the error is "invalid body: section is required; name is required"

  Scenario: Update context with empty body is allowed
    Given a global section "style" with a context "go"
    When I PUT "/api/v1/contexts/global/style/go" with body:
      """
      {}
      """
    Then the status is 200

  Scenario: Malformed JSON
    When I POST "/api/v1/projects" with body:
      """
      {"name":
      """
    Then the status is 400
    And the body contains "invalid JSON body"
