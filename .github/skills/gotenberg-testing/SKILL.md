---
name: gotenberg-testing
description: Select and run Gotenberg Makefile checks, Godog integration scenarios and Go unit tests, with explicit prerequisites and failure reporting.
---

# Verify a Gotenberg change

Read `AGENTS.md` and the Makefile first. Run commands from the repository root; do not replace the Makefile with an invented Go or Docker workflow.

## Select coverage

- Features and routes: start with a scenario in `test/integration/features/`, reuse steps in `test/integration/scenario/` and fixtures in `test/integration/testdata/`.
- Pure Go logic: add table-driven `*_test.go` tests beside the implementation in `pkg/`.
- Harness-only documentation: check formatting, instruction consistency, frontmatter, relative links and Git symlink modes. Do not claim application checks were run.

## Prerequisites and commands

- Use the Go toolchain required by `go.mod`, `golangci-lint` compatible with `.golangci.yml`, and a working Docker daemon.
- For non-Go formatting, use the Node version in `.node-version` and `npm ci` with `package-lock.json` when dependencies are missing.
- `make fmt` also runs `go fix` and `go mod tidy`; review all changes.
- `make lint` checks Go; `make lint-prettier` checks non-Go files.
- `make build` builds the image configured in the tracked `.env`, without publishing it.
- With that image available, start integration coverage with a relevant tag, for example `make test-integration TAGS=health NO_CONCURRENCY=true`. Tags are listed in the Makefile.
- Complete application verification with `make test-integration` and `make test-unit`, following the definition of done in `AGENTS.md`.

## Report evidence

Record each command, exit result and meaningful failure. Distinguish regressions from pre-existing failures. Missing tooling or an unavailable Docker daemon is a blocker, not a successful check. Do not install cloud credentials, publish images, alter deployment workflows or mutate infrastructure to bypass a blocker.
