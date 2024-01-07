Feature: Create a new resource
  As a admin 
  I want to be able to create a new files and folders
  So that I can organize my files and folders

  Background: 
    Given "admin" has logged in
    And admin has navigated to the homepage

  Scenario: Create a new folder
    When admin creates a new folder named "myFolder"
    Then admin should be able to see a folder named "myFolder"

  Scenario: Create a new file with content
    When admin creates a new file named "myFile.txt" with content "Hello World"
    Then admin should be able to see a file named "myFile.txt" with content "Hello World"

  Scenario: Rename a file
    Given admin has created a file named "oldfile.txt" with content "Hello World"
    When admin renames a file "oldfile.txt" to "newfile.txt"
    Then admin should be able to see file with "newfile.txt" name

  Scenario: Delete a file
    Given admin creates a new file named "delMyFile.txt" using API
    When admin deletes a file named "delMyFile.txt"
    Then admin shouln't see "delMyFile" in the UI
