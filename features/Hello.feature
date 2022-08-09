Feature: Hello, World!

  Scenario: Happy path.
    Given there are no rows in table "greetings"
    When I request HTTP endpoint with method "GET" and URI "/hello?name=Jane&locale=en-US"
    And I concurrently request idempotent HTTP endpoint
    Then I should have response with body
    """
    {"message":"Hello, Jane!"}
    """
    And I should have response with status "OK"
    And only these rows are available in table "greetings":
      | message      |
      | Hello, Jane! |

  Scenario: Unhappy path.
    Given there are no rows in table "greetings"
    When I request HTTP endpoint with method "GET" and URI "/hello?name=Jane&locale=zz-ZZ"
    And I concurrently request idempotent HTTP endpoint
    Then I should have response with body
    """
    {
     "status":"INVALID_ARGUMENT","error":"invalid argument: validation failed",
     "context":{"query:locale":["#: value must be one of \"en-US\", \"ru-RU\""]}
    }
    """
    And I should have response with status "Bad Request"
    And no rows are available in table "greetings"

  Scenario: Buggy path.
    Given there are no rows in table "greetings"
    When I request HTTP endpoint with method "GET" and URI "/hello?name=Bug&locale=ru-RU"
    And I concurrently request idempotent HTTP endpoint
    Then I should have response with body
    """
    {
     "error":"#$@@^! %C 🤖",
     "context":{"trace.id":"<ignore-diff>","transaction.id":"<ignore-diff>"}
    }
    """
    And I should have response with status "Internal Server Error"
    And no rows are available in table "greetings"
