Feature: Version
  As a user I want to see which pcm version I run.

  Scenario: Show the CLI version
    When I run "pcm --version"
    Then the exit code is 0
    And the output contains "pcm dev (commit none, built unknown)"

  Scenario: Version needs no server
    Given the server is not running
    When I run "pcm --version"
    Then the exit code is 0
    And the output contains "pcm dev"
