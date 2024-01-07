Feature: Login
  As an admin
  I want to login into the application
  So I can manage my files

  Background: 
    Given admin has browsed to the login page

  Scenario: admin logs in with correct credentials
    When admin logs in with username as 'admin' and password as 'admin'
    Then admin should be navigated to homescreen

  Scenario Outline: admin logs in with incorrect credentials
    When admin logs in with username as "<username>" and password as "<password>"
    Then admin should see "Wrong credentials" message

    Examples: 
      | username | password |
      | user99   | admin    |
      | admin    | user99   |
      |          | admin    |
      | admin    |          |
      |          |          |
