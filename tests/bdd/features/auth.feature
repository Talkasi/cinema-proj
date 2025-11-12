Feature: Email 2FA Authentication
  As a user of the cinema management system
  I want to be able to authenticate with email-based two-factor authentication
  So that my account is more secure

  Scenario: Successful login without 2FA
    Given a user with valid credentials exists
    When the user logs in with correct email and password
    Then the user should be authenticated successfully
    And should receive a valid JWT token

  Scenario: Login with email 2FA enabled
    Given a user with valid credentials and email 2FA enabled exists
    When the user logs in with correct email and password
    Then the user should receive a temporary authentication status
    And should be prompted to enter the email 2FA code

  Scenario: Successful email 2FA verification
    Given a user with valid credentials and email 2FA enabled exists
    When the user logs in with correct email and password
    When the user provides the correct 2FA code
    Then the user should be fully authenticated
    And should receive a valid JWT token

  Scenario: Failed email 2FA verification
    Given a user with valid credentials and email 2FA enabled exists
    And a valid email 2FA code is generated
    When the user provides an incorrect 2FA code
    Then the user should not be authenticated
    And should receive an error message about invalid 2FA code

  Scenario: Lock account after failed login attempts
    Given a user with valid credentials exists
    When the user attempts to log in with incorrect password 5 times
    Then the user account should be temporarily locked
    And subsequent login attempts should fail with "account locked" message

  Scenario: Account recovery after lockout
    Given a user with valid credentials exists
    And the user's account is locked due to failed login attempts
    When 30 minutes have passed since the lockout
    Then the user should be able to log in with correct credentials

  Scenario: Enable email 2FA
    Given a user with valid credentials exists
    When the user enables email two-factor authentication
    Then the 2FA should be enabled in the user's account

  Scenario: Disable 2FA
    Given a user with valid credentials and 2FA enabled exists
    When the user disables two-factor authentication
    Then the 2FA should be disabled in the user's account