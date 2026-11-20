@outbound-security
Feature: Outbound request security

  @download-from
  Scenario: Reject a private downloadFrom destination when enabled
    Given I have a Gotenberg container with the following environment variable(s):
      | API_DOWNLOAD_FROM_DENY_PRIVATE_IPS | true |
    When I make a "POST" request to Gotenberg at the "/forms/pdfengines/flatten" endpoint with the following form data and header(s):
      | downloadFrom | [{"url":"http://127.0.0.1:3000/health"}] | field |
    Then the response status code should be 403

  @webhook
  Scenario: Reject a private webhook destination when enabled
    Given I have a Gotenberg container with the following environment variable(s):
      | WEBHOOK_DENY_PRIVATE_IPS | true |
    When I make a "POST" request to Gotenberg at the "/forms/pdfengines/flatten" endpoint with the following form data and header(s):
      | files                       | testdata/page_1.pdf          | file   |
      | Gotenberg-Webhook-Url       | http://127.0.0.1:3000/health | header |
      | Gotenberg-Webhook-Error-Url | http://127.0.0.1:3000/health | header |
    Then the response status code should be 403
