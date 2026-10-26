Feature: Claude Code skill
  As a user I want to install the pcm skill for Claude Code.

  Scenario: Install skill
    When I run "pcm skill install"
    Then the exit code is 0
    And the output contains "skill/pcm installed in"
    And the file "~/.claude/skills/pcm/SKILL.md" contains "name: pcm"

  Scenario: Install overwrites
    Given the file "~/.claude/skills/pcm/SKILL.md" has content "old"
    When I run "pcm skill install"
    Then the exit code is 0
    And the file "~/.claude/skills/pcm/SKILL.md" contains "name: pcm"

  Scenario: Show skill
    When I run "pcm skill show"
    Then the exit code is 0
    And the output contains "name: pcm"
    And the output contains "pcm get contexts -o md"

  Scenario: Uninstall skill
    Given I run "pcm skill install"
    When I run "pcm skill uninstall"
    Then the exit code is 0
    And the output contains "skill/pcm uninstalled from"
    And the file "~/.claude/skills/pcm/SKILL.md" does not exist

  Scenario: Uninstall when not installed
    When I run "pcm skill uninstall"
    Then the exit code is 0
    And the output contains "skill/pcm not installed"

  Scenario: Skill dumps session
    When I run "pcm skill show"
    Then the exit code is 0
    And the output contains "## 5. Dump session"
    And the output contains "pcm create context <name> -s <section> -f <file>"
    And the output does not contain "<<'EOF'"
    And the output contains "scratchpad"
    And the output contains "If there is no branch"
    And the output contains "--force"
    And the output contains "Bash(gh pr view:*)"
    And the output contains "Bash(git branch:*)"
    And the output contains ", Write"
