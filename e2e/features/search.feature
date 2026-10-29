Feature: Search
  As a user I want fast, ranked, fuzzy full-text search of contexts.

  Background:
    Given the server is running
    And a global section "style" with a context "gostyle" and content "Use gofmt for formatting." and description "Go style"

  Scenario: Exact hit
    When I run "pcm search gofmt"
    Then the exit code is 0
    And the output contains "SCOPE"
    And the line with "gostyle" contains "style"

  Scenario: Fuzzy hit with typo
    When I run "pcm search gofmat"
    Then the exit code is 0
    And the output contains "gostyle"

  Scenario: Multi-word query
    When I run "pcm search 'gofmt formatting'"
    Then the exit code is 0
    And the output contains "gostyle"

  Scenario: Plus means word must match
    Given a global section "review" with a context "gorev" and content "Check gofmt in review."
    When I run "pcm search '+gofmt +review'"
    Then the exit code is 0
    And the output contains "gorev"
    And the output does not contain "gostyle"

  Scenario: Minus means word must not match
    Given a global section "review" with a context "gorev" and content "Check gofmt in review."
    When I run "pcm search 'gofmt -review'"
    Then the exit code is 0
    And the output contains "gostyle"
    And the output does not contain "gorev"

  Scenario: Exact phrase
    Given a global section "review" with a context "gorev" and content "Check gofmt in review."
    When I GET "/api/v1/search?q=%22gofmt%20for%20formatting%22"
    Then the status is 200
    And the body contains "\"id\":\"global/style/gostyle\""
    And the body does not contain "gorev"

  Scenario: Title match ranks above content match
    Given a global section "notes" with a context "title-hit" and title "Testing" and content "x"
    And a global section "notes" with a context "body-hit" and content "Some testing tips."
    When I run "pcm search testing"
    Then the exit code is 0
    And the line with "title-hit" comes before the line with "body-hit"

  Scenario: Word in the name matches
    Given a global section "notes" with a context "error-handling" and content "x"
    When I run "pcm search handling"
    Then the exit code is 0
    And the output contains "error-handling"

  Scenario: Section name matches
    Given a global section "review" with a context "pr" and content "x"
    When I run "pcm search review"
    Then the exit code is 0
    And the line with "pr" contains "review"

  Scenario: Project name matches
    Given a project "billing" section "style" with a context "ctx" and content "x"
    When I run "pcm search billing --all"
    Then the exit code is 0
    And the output contains "ctx"

  Scenario: Name match ranks above content match
    Given a global section "notes" with a context "caching" and content "x"
    And a global section "notes" with a context "other" and content "Some caching tips."
    When I run "pcm search caching"
    Then the line with "caching" comes before the line with "other"

  Scenario: Filter by section
    Given a global section "review" with a context "gorev" and content "Check gofmt in review."
    When I run "pcm search gofmt -s review"
    Then the exit code is 0
    And the output contains "gorev"
    And the output does not contain "gostyle"

  Scenario: Project flag searches global and that project
    Given a project "api" section "style" with a context "apigo" and content "Run gofmt in api."
    And a project "web" section "style" with a context "webgo" and content "Run gofmt in web."
    When I run "pcm search gofmt -p api"
    Then the exit code is 0
    And the output contains "apigo"
    And the output contains "gostyle"
    And the output does not contain "webgo"

  Scenario: In a git repo search global and current project
    Given a project "api" section "style" with a context "apigo" and content "Run gofmt in api."
    And a project "web" section "style" with a context "webgo" and content "Run gofmt in web."
    And I am in a git repo named "api"
    When I run "pcm search gofmt"
    Then the exit code is 0
    And the output contains "apigo"
    And the output contains "gostyle"
    And the output does not contain "webgo"

  Scenario: Global flag searches global only
    Given a project "api" section "style" with a context "apigo" and content "Run gofmt in api."
    And I am in a git repo named "api"
    When I run "pcm search gofmt -g"
    Then the exit code is 0
    And the output contains "gostyle"
    And the output does not contain "apigo"

  Scenario: All flag searches all projects
    Given a project "api" section "style" with a context "apigo" and content "Run gofmt in api."
    And a project "web" section "style" with a context "webgo" and content "Run gofmt in web."
    And I am in a git repo named "api"
    When I run "pcm search gofmt --all"
    Then the exit code is 0
    And the output contains "apigo"
    And the output contains "gostyle"
    And the output contains "webgo"

  Scenario: Not in a git repo searches all projects
    Given a project "web" section "style" with a context "webgo" and content "Run gofmt in web."
    When I run "pcm search gofmt"
    Then the exit code is 0
    And the output contains "gostyle"
    And the output contains "webgo"

  Scenario: Only one scope flag
    When I run "pcm search gofmt -p api --all"
    Then the exit code is 1
    And the output contains "error: use only one of -p, -g and --all"

  Scenario: Limit results
    Given a global section "review" with a context "gorev" and content "Check gofmt in review."
    When I run "pcm search gofmt --limit 1"
    Then the exit code is 0
    And the output contains one line with "global"

  Scenario: Zero limit
    When I run "pcm search gofmt --limit 0"
    Then the exit code is 1
    And the output contains "error: invalid limit"

  Scenario: Negative limit
    When I run "pcm search gofmt --limit -1"
    Then the exit code is 1
    And the output contains "error: invalid limit"

  Scenario: No hit
    When I run "pcm search zzzqqq"
    Then the exit code is 0
    And the output contains "No results found."

  Scenario: Missing query
    When I run "pcm search"
    Then the exit code is 1
    And the output contains "error: missing query"

  Scenario: New context is searchable at once
    When I run "pcm create context pyblack -s style -g -m 'Use black.'"
    And I run "pcm search black"
    Then the output contains "pyblack"

  Scenario: Updated context is searchable at once
    Given the context "global/style/gostyle" has content "Use goimports."
    When I run "pcm search goimports"
    Then the output contains "gostyle"
    When I run "pcm search gofmt"
    Then the output contains "No results found."

  Scenario: Deleted context is not found
    When I run "pcm delete context gostyle -s style -g"
    And I run "pcm search gofmt"
    Then the output contains "No results found."

  Scenario: Rebuild finds file written outside pcm
    Given the file "~/.pcm/repo/global/style/rsfmt.md" has content "Use rustfmt."
    When I run "pcm search rustfmt"
    Then the output contains "No results found."
    When I run "pcm index rebuild"
    Then the exit code is 0
    And the output contains "index rebuilt: 2 contexts"
    When I run "pcm search rustfmt"
    Then the output contains "rsfmt"

  Scenario: Search finds file pulled by sync
    Given a git remote
    And the remote is set
    And I run "pcm sync"
    And another clone pushes "global/style/pyblack.md" with content "Use black."
    When I run "pcm sync"
    And I run "pcm search black"
    Then the output contains "pyblack"

  Scenario: Search works after server restart
    When I run "pcm server stop"
    And I run "pcm server start"
    And I run "pcm search gofmt"
    Then the exit code is 0
    And the output contains "gostyle"

  Scenario: API search returns hit fields
    When I GET "/api/v1/search?q=gofmt"
    Then the status is 200
    And the body contains "\"id\":\"global/style/gostyle\""
    And the body contains "\"title\":\"gostyle\""
    And the body contains "\"snippet\""
    And the body contains "\"score\""

  Scenario: API project with global
    Given a project "api" section "style" with a context "apigo" and content "Run gofmt in api."
    When I GET "/api/v1/search?q=gofmt&project=api&global=true"
    Then the status is 200
    And the body contains "\"id\":\"projects/api/style/apigo\""
    And the body contains "\"id\":\"global/style/gostyle\""

  Scenario: API missing q
    When I GET "/api/v1/search"
    Then the status is 400
    And the error is "missing q"

  Scenario: API bad limit
    When I GET "/api/v1/search?q=go&limit=abc"
    Then the status is 400
    And the error is "invalid limit"

  Scenario: API zero limit
    When I GET "/api/v1/search?q=go&limit=0"
    Then the status is 400
    And the error is "invalid limit"

  Scenario: API rebuild
    When I POST "/api/v1/index/rebuild"
    Then the status is 200
    And the body contains "\"indexed\":1"
