---
name: gotenberg-implement
description: Implement backward-compatible Go modules and HTTP API changes in Gotenberg, with Makefile checks and Godog integration scenarios.
---

# Gotenberg implementation

Read `AGENTS.md` before working and follow its compatibility and validation rules.

1. Trace the relevant route, form fields, flags and module interfaces in `pkg/`. Keep business logic out of `cmd/` and reuse existing module helpers.
2. Identify compatibility risks before editing. Preserve existing defaults, flags, environment variables, API form fields and error behavior unless a breaking change was explicitly requested.
3. For feature or route changes, start with the corresponding Godog scenario in `test/integration/features/`. Reuse step definitions in `test/integration/scenario/` and add fixtures only when necessary.
4. Implement the smallest complete change. Handle malformed input and cancellation without panics; maintain module lifecycle and concurrency guarantees.
5. Add table-driven unit tests for pure logic and GoDoc for exported symbols. Use the `gotenberg-testing` skill for the repository's verification workflow.
6. Report the behavior changed, compatibility impact and actual check results, including any blocked checks. Never publish images, merge a PR or alter deployment workflows without an explicit request.
