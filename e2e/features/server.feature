Feature: Background server
  As a user I want to start and stop the server in the background.

  Scenario: Start server
    Given the server is not running
    When I run "pcm server start"
    Then the exit code is 0
    And the output contains "server: started"
    And the pid file exists

  Scenario: Start when already running
    Given the server is running
    When I run "pcm server start"
    Then the exit code is 0
    And the output contains "server: already running"

  Scenario: Stop server
    Given the server is running
    When I run "pcm server stop"
    Then the exit code is 0
    And the output contains "server: stopped"
    And the pid file does not exist

  Scenario: Stop when not running
    Given the server is not running
    When I run "pcm server stop"
    Then the exit code is 0
    And the output contains "server: not running"

  Scenario: Status when running
    Given the server is running
    When I run "pcm server status"
    Then the exit code is 0
    And the output contains "server: ok"
    And the output contains the server pid

  Scenario: Stale pid file
    Given a pid file with a dead pid
    When I run "pcm server status"
    Then the exit code is 1
    And the pid file does not exist
