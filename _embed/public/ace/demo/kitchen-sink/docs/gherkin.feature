@these @_are_ @tags
Feature: Serve coffee
  Coffee should not be served until paid for
  Coffee should not be served until the button has been pressed
  If there is no coffee left then money should be refunded
  
  Scenario Outline: Eating
    Given there are <start> cucumbers
    When I eat <eat> cucumbers
    Then I should have <left> cucumbers

    Examples:
      | start | eat | left |
      |  12   |  5  |  7   |
      |  @20  |  5  |  15  |    

  Scenario: Buy last coffee
    Given there are 1 coffees left in the machine
    And I have deposited 1$ 
    When I press the coffee button
    Then I should be served a "coffee"
    
  # this a comment
  
  """
  this is a 
  pystring
  """