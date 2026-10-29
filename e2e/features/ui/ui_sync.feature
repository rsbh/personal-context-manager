@ui
Feature: Sync
  As a user I want to see the git status and sync my contexts in the web UI.

  Scenario: Header badge shows pending changes
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI
    Then the Sync badge shows "1"

  Scenario: Sync page shows the status
    Given the server is running
    And a git remote
    And the remote is set
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI
    And I click the Sync link
    Then the URL path is "/sync"
    And I see "main"
    And the "Changes" card shows "1"
    And the "Ahead" card shows "0"
    And the "Behind" card shows "0"
    And the changed files show "global/"

  Scenario: Sync now pushes the changes
    Given the server is running
    And a git remote
    And the remote is set
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/sync"
    And I click the "Sync now" button
    Then I see the toast "Synced"
    And the "Changes" card shows "0"
    And the Sync badge is not shown
    And the recent commits show "pcm sync"
    And the remote has the file "global/style/go.md"

  Scenario: Sync now pulls changes
    Given the server is running
    And a git remote
    And the remote is set
    And a global section "style" with a context "go" and content "Use gofmt."
    And I run "pcm sync"
    And another clone pushes "global/style/py.md" with content "Use black."
    When I open the UI at "/sync"
    And I click the "Sync now" button
    Then I see the toast "1 commit pulled"
    And the sidebar shows "py" under "style"

  Scenario: Sync conflict
    Given the server is running
    And a git remote
    And the remote is set
    And a global section "style" with a context "go" and content "Use gofmt."
    And I run "pcm sync"
    And another clone pushes "global/style/go.md" with content "remote rule"
    And the context "global/style/go" has content "local rule"
    When I open the UI at "/sync"
    And I click the "Sync now" button
    Then I see "Sync conflict, rebase aborted"
    And I see "global/style/go.md"

  Scenario: No remote
    Given the server is running
    When I open the UI at "/sync"
    Then I see "Remote not set"
    And I see "pcm config set remote <git-url>"
    And the "Sync now" button is disabled
