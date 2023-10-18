Feature: login
    As a user
    I want to login to the system
    So that I can manage the files and folders

Scenario: login with valid username and valid password
    Given the user has browsed to the login page
    When user logs in with username "admin" and password "admin"
    Then user should redirect to the homepage