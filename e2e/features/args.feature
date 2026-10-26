Feature: Clear argument errors
  As a user I want clear errors when I give wrong arguments.

  Scenario Outline: Missing name
    When I run "<command>"
    Then the exit code is 1
    And the output contains "error: missing <resource> name"
    And the output contains "usage: <command> <name>"

    Examples:
      | command            | resource |
      | pcm create project | project  |
      | pcm delete project | project  |
      | pcm create section | section  |
      | pcm delete section | section  |

  Scenario: Too many arguments
    When I run "pcm create section a b"
    Then the exit code is 1
    And the output contains "error: too many arguments: want 1 (name), got 2"

  Scenario: Unexpected argument
    When I run "pcm get sections x"
    Then the exit code is 1
    And the output contains "error: unexpected argument \"x\""
