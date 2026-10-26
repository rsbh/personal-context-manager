Feature: pcm ui command
  As a user I want to open the web UI from the CLI.

  Scenario: Server not running
    Given the server is not running
    When I run "pcm ui"
    Then the exit code is 1
    And the output contains "hint:   run `pcm server start`"

  Scenario: Server running
    Given the server is running
    When I run "pcm ui"
    Then the exit code is 0
    And the output contains "ui: http://127.0.0.1:"
