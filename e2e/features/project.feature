Feature: Projects
  As a user I want to manage projects from the CLI.

  Scenario: List with no projects
    Given the server is running
    When I run "pcm get projects"
    Then the exit code is 0
    And the output contains "No projects found."

  Scenario: Create project
    Given the server is running
    When I run "pcm create project pcm"
    Then the exit code is 0
    And the output contains "project/pcm created"

  Scenario: Create existing project
    Given the server is running
    And a project "pcm"
    When I run "pcm create project pcm"
    Then the exit code is 1
    And the output contains "already exists"

  Scenario: Create with invalid name
    Given the server is running
    When I run "pcm create project Bad/Name"
    Then the exit code is 1
    And the output contains "invalid"

  Scenario: List projects
    Given the server is running
    And a project "api"
    And a project "web"
    When I run "pcm get projects"
    Then the exit code is 0
    And the output contains "NAME"
    And the output contains "api"
    And the output contains "web"

  Scenario: Get one project
    Given the server is running
    And a project "api"
    When I run "pcm get project api"
    Then the exit code is 0
    And the output contains "api"

  Scenario: Get missing project
    Given the server is running
    When I run "pcm get project none"
    Then the exit code is 1
    And the output contains "not found"

  Scenario: Delete empty project
    Given the server is running
    And a project "api"
    When I run "pcm delete project api"
    Then the exit code is 0
    And the output contains "project \"api\" deleted"

  Scenario: Delete project with sections needs force
    Given the server is running
    And a project "api" with a section "style"
    When I run "pcm delete project api"
    Then the exit code is 1
    And the output contains "use --force"
    When I run "pcm delete project api --force"
    Then the exit code is 0

  Scenario: Server not running
    Given the server is not running
    When I run "pcm get projects"
    Then the exit code is 1
    And the output contains "hint:   run `pcm server start`"
