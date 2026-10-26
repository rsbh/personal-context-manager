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
