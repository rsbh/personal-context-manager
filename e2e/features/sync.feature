Feature: Git sync
  As a user I want to sync my context repo with a git remote.

  Scenario: Sync without remote commits locally
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I run "pcm sync"
    Then the exit code is 0
    And the line with "committed:" contains "2 files"
    And the output contains "remote not set, contexts are local only"
    And the output contains "pcm config set remote <git-url> && pcm sync"

  Scenario: Sync pushes to remote
    Given the server is running
    And a git remote
    And the remote is set
    And a global section "style" with a context "go" and content "Use gofmt."
    When I run "pcm sync"
    Then the exit code is 0
    And the line with "pushed:" contains "yes"
    And the remote has the file "global/style/go.md"

  Scenario: Nothing to sync
    Given the server is running
    And a git remote
    And the remote is set
    And a global section "style" with a context "go" and content "Use gofmt."
    And I run "pcm sync"
    When I run "pcm sync"
    Then the exit code is 0
    And the line with "committed:" contains "0 files"
    And the line with "pushed:" contains "no"

  Scenario: Pull changes from remote
    Given the server is running
    And a git remote
    And the remote is set
    And a global section "style" with a context "go" and content "Use gofmt."
    And I run "pcm sync"
    And another clone pushes "global/style/py.md" with content "Use black."
    When I run "pcm sync"
    Then the exit code is 0
    And the line with "pulled:" contains "1 commit"
    When I run "pcm get context py -s style -g"
    Then the output contains "Use black."

  Scenario: Conflict aborts the sync
    Given the server is running
    And a git remote
    And the remote is set
    And a global section "style" with a context "go" and content "Use gofmt."
    And I run "pcm sync"
    And another clone pushes "global/style/go.md" with content "remote rule"
    And the context "global/style/go" has content "local rule"
    When I run "pcm sync"
    Then the exit code is 1
    And the output contains "sync conflict, rebase aborted"
    And the output contains "global/style/go.md"
    And the output contains "then pcm sync"

  Scenario: Status shows changes
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I run "pcm sync status"
    Then the exit code is 0
    And the line with "branch:" contains "main"
    And the line with "remote:" contains "none"
    And the output contains "global/"
    And the output contains "remote not set, contexts are local only"

  Scenario: Log
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    And I run "pcm sync"
    When I run "pcm sync log"
    Then the exit code is 0
    And the output contains "HASH"
    And the output contains "pcm sync: 2 files changed"

  Scenario: Log with no commits
    Given the server is running
    When I run "pcm sync log"
    Then the exit code is 0
    And the output contains "No commits yet."
