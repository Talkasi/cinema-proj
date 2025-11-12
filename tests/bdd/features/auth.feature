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
    And a valid email 2FA code is generated
    When the user provides the correct 2FA code
    Then the user should be fully authenticated
    And should receive a valid JWT token