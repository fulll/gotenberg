# Shared Copilot and Claude harness

`AGENTS.md` is the canonical repository guidance. Its original instructions are preserved; the repository context and harness sections were added using the adapted `create-agentsmd` skill.

`CLAUDE.md` contains only `@AGENTS.md`. Copilot loads the root `AGENTS.md` and `.github/copilot-instructions.md` points to the same file. Agents and skills have one source under `.github/`; `.claude/` exposes them through relative symbolic links.

## Inventory

| Kind  | Name                  | Source                                |
| ----- | --------------------- | ------------------------------------- |
| Agent | `gotenberg-implement` | `agents/gotenberg-implement.agent.md` |
| Agent | `refine-issue`        | `agents/refine-issue.agent.md`        |
| Skill | `create-agentsmd`     | `skills/create-agentsmd/SKILL.md`     |
| Skill | `gotenberg-testing`   | `skills/gotenberg-testing/SKILL.md`   |

The Go implementation agent and testing skill are repository-specific. Terraform and AWS agents from the reference harnesses are intentionally excluded: this is a Go document-conversion service, not an infrastructure repository.

No cloud setup workflow is introduced for this Markdown/symlink change. Application tooling and Docker prerequisites remain documented in `AGENTS.md`; this harness does not install conversion engines on the host or modify build/deployment workflows. Any future cloud setup must be justified separately, use a single `copilot-setup-steps` job, `permissions: contents: read`, SHA-pinned actions and no AWS credentials or secrets.

## Third-party provenance

The following files are adapted from [github/awesome-copilot](https://github.com/github/awesome-copilot) at commit [`5269f7ea2d6c3a9bfda5ed878b6726cf509121fd`](https://github.com/github/awesome-copilot/commit/5269f7ea2d6c3a9bfda5ed878b6726cf509121fd):

| Local file                        | Pinned upstream source                                                                                                                              | Adaptation                                                                     |
| --------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| `agents/refine-issue.agent.md`    | [refine-issue.agent.md](https://github.com/github/awesome-copilot/blob/5269f7ea2d6c3a9bfda5ed878b6726cf509121fd/agents/refine-issue.agent.md)       | Portable frontmatter, Gotenberg scope, propose changes before publishing.      |
| `skills/create-agentsmd/SKILL.md` | [create-agentsmd/SKILL.md](https://github.com/github/awesome-copilot/blob/5269f7ea2d6c3a9bfda5ed878b6726cf509121fd/skills/create-agentsmd/SKILL.md) | Preserve existing guidance, Makefile evidence and single-source harness rules. |

Both use the upstream [MIT license](https://github.com/github/awesome-copilot/blob/5269f7ea2d6c3a9bfda5ed878b6726cf509121fd/LICENSE), reproduced in [harness-LICENSE.txt](harness-LICENSE.txt). Keep this attribution and license when updating these adaptations.

Design references: [fulll/chapter-secops#1946](https://github.com/fulll/chapter-secops/issues/1946), [fulll/security-stack#967](https://github.com/fulll/security-stack/pull/967) and [fulll/dwh.infra#149](https://github.com/fulll/dwh.infra/pull/149). Infrastructure-specific content and workflow changes from those examples were not copied.

## Validation

From the repository root:

```sh
make lint-prettier
find .claude -type l -exec test ! -e {} \; -print
git ls-tree -r HEAD .claude
```

The `find` command must print nothing; the committed tree must show four entries with mode `120000`. Check that every `.github/agents/*.agent.md` and `.github/skills/*/SKILL.md` has a matching Claude link, and that agent frontmatter contains only `name` and `description`.

Repeat the link checks in a fresh macOS/Linux clone after checkout. Use a filesystem and Git configuration that preserve symbolic links.

For local Copilot discovery without an AI prompt, run `copilot --trust-folder skill list` and `copilot --trust-folder instruction list` in a repository you trust. The flag trusts the working directory for that invocation only. These commands check skill and instruction discovery, not custom-agent execution.

With the relevant CLI installed and authenticated, open Claude Code in the repository and inspect `/agents` and `/skills`; open Copilot CLI and inspect `/agent` and `/skills`. Confirm both expose `gotenberg-implement`, `refine-issue`, `create-agentsmd` and `gotenberg-testing`. Ask each tool to identify the repository instruction source and confirm `AGENTS.md` is loaded. CLI versions may expose different discovery commands; record the actual commands and results rather than treating static link checks as proof of runtime discovery. Cloud-agent loading requires a separate cloud session and is not proven by local checks.

Include real check results and any unavailable runtime/cloud checks in the PR. Reference `fulll/chapter-secops#1946` without closing it.
