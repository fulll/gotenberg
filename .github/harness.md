# Shared Copilot and Claude harness

`AGENTS.md` is the canonical repository guidance. Its original instructions are preserved; Gotenberg-specific additions belong there or in distinctly named local agents/skills, never in upstream files.

`CLAUDE.md` contains only `@AGENTS.md`. Copilot loads the root `AGENTS.md` and `.github/copilot-instructions.md` points to the same file. Agents and skills have one source under `.github/`; `.claude/` exposes them through relative symbolic links.

## Inventory

| Kind  | Name                  | Source                                |
| ----- | --------------------- | ------------------------------------- |
| Agent | `gotenberg-implement` | `agents/gotenberg-implement.agent.md` |
| Skill | `create-agentsmd`     | `skills/create-agentsmd/SKILL.md`     |
| Skill | `gotenberg-testing`   | `skills/gotenberg-testing/SKILL.md`   |

The Go implementation agent and testing skill are repository-specific: `gotenberg-implement` covers backward-compatible module/API changes and `gotenberg-testing` covers the Makefile/Godog workflow. `create-agentsmd` is the mandatory common skill.

Only the common scope of the tracking issue applies. This is a Go document-conversion service, not a reusable Terraform AWS module repository. The conditional Terraform section does not apply: `terraform-aws-implement`, `terraform-iac-reviewer`, `terraform.instructions.md` and Terraform validation workflows are neither pre-existing nor added.

GitHub issues are disabled on `fulll/gotenberg` (`has_issues: false`, verified through the repository API). The `refine-issue` agent and its Claude link, initially added by this PR, are removed rather than preserved as pre-existing customizations. No issue-writing, epic-writing or create-issue-from-prompt components exist to deduplicate. Other optional upstream skills (`agentic-eval`, `aws-well-architected-review`, `devops-rollout-plan`, `github-actions-hardening`) are not installed because this change establishes no concrete need for them.

No cloud setup workflow is introduced for this Markdown/symlink change. Application tooling and Docker prerequisites remain documented in `AGENTS.md`; this harness does not install conversion engines on the host or modify build/deployment workflows. Any future cloud setup must be justified separately, use a single `copilot-setup-steps` job, `permissions: contents: read`, SHA-pinned actions and no AWS credentials or secrets.

## Third-party provenance

The following paths are complete, unmodified copies (copie intégrale non modifiée), byte for byte, from [github/awesome-copilot](https://github.com/github/awesome-copilot) at commit [`5269f7ea2d6c3a9bfda5ed878b6726cf509121fd`](https://github.com/github/awesome-copilot/commit/5269f7ea2d6c3a9bfda5ed878b6726cf509121fd):

| Local path                | Pinned upstream source                                                                                                             | Policy                                                                  |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| `skills/create-agentsmd/` | [create-agentsmd/](https://github.com/github/awesome-copilot/tree/5269f7ea2d6c3a9bfda5ed878b6726cf509121fd/skills/create-agentsmd) | Complete, unmodified directory, including all files and subdirectories. |

The retained upstream skill uses the [MIT license](https://github.com/github/awesome-copilot/blob/5269f7ea2d6c3a9bfda5ed878b6726cf509121fd/LICENSE), reproduced in [harness-LICENSE.txt](harness-LICENSE.txt). Keep this attribution and license when updating the copy. Do not translate, reformat, normalize frontmatter or otherwise edit upstream files. Local agents and skills are maintained separately and do not claim upstream provenance.

Design references: [fulll/chapter-secops#1946](https://github.com/fulll/chapter-secops/issues/1946), [fulll/security-stack#967](https://github.com/fulll/security-stack/pull/967) and [fulll/dwh.infra#149](https://github.com/fulll/dwh.infra/pull/149). Infrastructure-specific content and workflow changes from those examples were not copied.

## Validation

From the repository root:

```sh
make lint-prettier
find .claude -type l -exec test ! -e {} \; -print
git ls-tree -r HEAD .claude
```

The `find` command must print nothing; the committed tree must show three entries with mode `120000`. Check that every `.github/agents/*.agent.md` and `.github/skills/*/SKILL.md` has a matching Claude link. Check local agent frontmatter separately; upstream frontmatter must remain byte-identical to the pinned source, not reduced to `name` and `description`.

Against a clone of `https://github.com/github/awesome-copilot.git`, checked out at the pinned commit, run (replace `/path/to/awesome-copilot` with that clone's path):

```sh
git -C /path/to/awesome-copilot rev-parse HEAD
diff -r /path/to/awesome-copilot/skills/create-agentsmd .github/skills/create-agentsmd
```

The commit must be `5269f7ea2d6c3a9bfda5ed878b6726cf509121fd`; the diff must be empty with exit status 0. Directory comparison checks the full skill, including any references and other files. If a formatter reports differences in an upstream copy, report them without changing that copy.

Repeat the link checks in a fresh macOS/Linux clone after checkout. Use a filesystem and Git configuration that preserve symbolic links.

For local Copilot discovery without an AI prompt, run `copilot --trust-folder skill list` and `copilot --trust-folder instruction list` in a repository you trust. The flag trusts the working directory for that invocation only. These commands check skill and instruction discovery, not custom-agent execution.

With the relevant CLI installed and authenticated, open Claude Code in the repository and inspect `/agents` and `/skills`; open Copilot CLI and inspect `/agent` and `/skills`. Confirm both expose the local agent `gotenberg-implement` and the skills `create-agentsmd` and `gotenberg-testing`. Ask each tool to identify the repository instruction source and confirm `AGENTS.md` is loaded. CLI versions may expose different discovery commands; record the actual commands and results rather than treating static link checks as proof of runtime discovery. Cloud-agent loading requires a separate cloud session and is not proven by local checks.

## Validation blockers and proposal

Copilot CLI static discovery lists the skills and instruction files, including `AGENTS.md`; `.github/copilot-instructions.md` was absent before this PR, so its pointer replaces no existing rules. Effective instruction loading in cloud mode and the Claude import, plus functional agent/skill parity in both tools, remain unverified. A pointer and valid symlinks do not resolve these acceptance blockers.

The official `create-agentsmd` skill instructs the agent to create a complete `AGENTS.md`, while existing Gotenberg instructions must not be overwritten. The upstream skill is intentionally unchanged. Suspend any operation that would replace existing guidance. Proposed resolution for maintainer validation: explicitly request additive maintenance, retain the original instruction bytes and review the resulting diff before accepting it. This proposal is not evidence that the conflict is resolved.

Validate interactive discovery and instruction loading in isolated Claude Code, Copilot CLI and cloud sessions, and repeat clone checks on Linux. If a tool rejects an official resource or symbolic links, record its exact error and submit a local wrapper or a generated, CI-checked synchronization alternative for approval; do not modify the official resource or silently install an alternative.

Include real check results and any unavailable runtime/cloud checks in the PR. Reference `fulll/chapter-secops#1946` without closing it.
