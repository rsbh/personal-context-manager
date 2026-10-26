Feature: Slow sync
  As a user I want sync to work when git is slow.

  Scenario: Slow push still succeeds
    Given the server is running
    And a slow git remote
    And the remote is set
    And a global section "style" with a context "go" and content "Use gofmt."
    When I run "pcm sync"
    Then the exit code is 0
    And the line with "pushed:" contains "yes"
    And the remote has the file "global/style/go.md"

  Scenario: Status before first push
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    And I run "pcm sync"
    And a git remote
    And the remote is set
    When I run "pcm sync status"
    Then the exit code is 0
    And the line with "ahead:" contains "1"
    And the output contains "run pcm sync to push"
