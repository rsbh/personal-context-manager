Feature: Health check
  As a user I want to check the server from the CLI.

  Scenario: Server not running
    Given the server is not running
    When I run "pcm health"
    Then the exit code is 1
    And the output contains "server: not reachable"
    And the output contains "hint:   run `pcm server start`"

  Scenario: Server running
    Given the server is running
    When I run "pcm health"
    Then the exit code is 0
    And the output contains "server: ok"
