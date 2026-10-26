Feature: Sections
  As a user I want to manage sections per scope from the CLI.

  Scenario: Create global section
    Given the server is running
    When I run "pcm create section style -g"
    Then the exit code is 0
    And the output contains "section/style created in global"

  Scenario: Create section in a project
    Given the server is running
    And a project "api"
    When I run "pcm create section style -p api"
    Then the exit code is 0
    And the output contains "section/style created in project \"api\""

  Scenario: Create section in missing project with -p
    Given the server is running
    When I run "pcm create section style -p none"
    Then the exit code is 1
    And the output contains "not found"

  Scenario: Current project from git repo, auto-created
    Given the server is running
    And I am in a git repo named "My.Repo"
    When I run "pcm create section style"
    Then the exit code is 0
    And the output contains "section/style created in project \"my-repo\""
    When I run "pcm get projects"
    Then the output contains "my-repo"

  Scenario: Get sections of current project
    Given the server is running
    And I am in a git repo named "api"
    And a project "api" with a section "style"
    When I run "pcm get sections"
    Then the exit code is 0
    And the output contains "NAME"
    And the output contains "style"

  Scenario: Get sections when current project is missing
    Given the server is running
    And I am in a git repo named "api"
    When I run "pcm get sections"
    Then the exit code is 0
    And the output contains "No sections found."

  Scenario: Not in a git repo and no flag
    Given the server is running
    And I am not in a git repo
    When I run "pcm get sections"
    Then the exit code is 1
    And the output contains "not in a git repo"
    And the output contains "use -p <project> or -g"

  Scenario: Both -p and -g
    Given the server is running
    When I run "pcm get sections -p api -g"
    Then the exit code is 1
    And the output contains "-p and -g"

  Scenario: Scopes are separate
    Given the server is running
    And a global section "style"
    And a project "api" with a section "review"
    When I run "pcm get sections -g"
    Then the output contains "style"
    And the output does not contain "review"

  Scenario: Create existing section
    Given the server is running
    And a global section "style"
    When I run "pcm create section style -g"
    Then the exit code is 1
    And the output contains "already exists"

  Scenario: Delete empty section
    Given the server is running
    And a global section "style"
    When I run "pcm delete section style -g"
    Then the exit code is 0
    And the output contains "section \"style\" deleted from global"

  Scenario: Delete section with contexts needs force
    Given the server is running
    And a global section "style" with a context "go"
    When I run "pcm delete section style -g"
    Then the exit code is 1
    And the output contains "use --force"
    When I run "pcm delete section style -g --force"
    Then the exit code is 0
