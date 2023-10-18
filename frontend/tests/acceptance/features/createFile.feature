Feature: Create a file
    As a user
    I want to manage a file
    So that I can save it

Background:
    Given the user has browsed to the login page
    And the user has logged in with username "admin" and password "admin"

Scenario: create a file
    When user has added file "demo.txt" with content "hello world"
    Then for user there should contain files "file.txt"