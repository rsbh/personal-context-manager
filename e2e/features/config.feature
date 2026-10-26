Feature: Config
  As a user I want to read and change pcm config from the CLI.

  Scenario: Set and get remote
    When I run "pcm config set remote git@example.com:me/ctx.git"
    Then the exit code is 0
    And the output contains "config/remote set"
    When I run "pcm config get remote"
    Then the exit code is 0
    And the output contains "git@example.com:me/ctx.git"

  Scenario: Get all keys
    When I run "pcm config get"
    Then the exit code is 0
    And the output contains "path:"
    And the output contains "remote:"
    And the output contains "port:"

  Scenario: Set port shows restart hint
    When I run "pcm config set port 9999"
    Then the exit code is 0
    And the output contains "config/port set"
    And the output contains "pcm server stop && pcm server start"

  Scenario: Unknown key
    When I run "pcm config set color red"
    Then the exit code is 1
    And the output contains "error: unknown key \"color\": use path, remote or port"

  Scenario: Invalid port
    When I run "pcm config set port abc"
    Then the exit code is 1
    And the output contains "invalid port"
