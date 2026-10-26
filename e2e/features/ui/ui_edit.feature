@ui
Feature: Edit contexts
  As a user I want to create, edit and delete contexts in the web UI.

  Scenario: Edit and save a context
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt." and description "Go style"
    When I open the UI at "/contexts/global/style/go"
    And I click the "Edit" button
    Then the URL path is "/edit/global/style/go"
    When I type "Go rules" in "Title"
    And I add the tag "fmt"
    And I type "Use **goimports**." in the editor
    Then I see "Unsaved changes"
    And I see "goimports" in bold
    When I click the "Save" button
    Then the URL path is "/contexts/global/style/go"
    And I see the toast "Saved"
    And I see "Go rules"
    And I see the raw file "- fmt"

  Scenario: Discard changes
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/edit/global/style/go"
    And I type "Changed." in the editor
    And I click the "Discard" button
    Then the editor contains "Use gofmt."
    And I do not see "Unsaved changes"

  Scenario: Warn before leaving with unsaved changes
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/edit/global/style/go"
    And I type "Changed." in the editor
    And I click "go" in the sidebar
    Then I see the dialog "Unsaved changes"
    When I click the "Stay" button in the dialog
    Then the URL path is "/edit/global/style/go"
    And the editor contains "Changed."
    When I click "go" in the sidebar
    And I click the "Leave" button in the dialog
    Then the URL path is "/contexts/global/style/go"
    And I see "Use gofmt."

  Scenario: Delete a context
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/contexts/global/style/go"
    And I click the "Delete" button
    Then I see the dialog "Delete context?"
    When I click the "Delete" button in the dialog
    Then the URL path is "/"
    And I see the toast "Deleted"
    And the sidebar does not show "go"

  Scenario: New context from the sidebar
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI
    And I click the "New context in style" button
    Then the URL path is "/new"
    And the "Scope" select shows "Global"
    And the "Section" select shows "style"
    When I type "py" in "Name"
    And I type "Use black." in the editor
    And I click the "Save" button
    Then the URL path is "/contexts/global/style/py"
    And I see the toast "Saved"
    And the sidebar shows "py" under "Global"

  Scenario: Create a context that exists
    Given the server is running
    And a global section "style" with a context "go" and content "Use gofmt."
    When I open the UI at "/new?scope=global&section=style"
    And I type "go" in "Name"
    And I type "Other." in the editor
    And I click the "Save" button
    Then I see the toast "already exists"
    And the URL path is "/new"
    And the editor contains "Other."
