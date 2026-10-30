---
name: create-agentsmd
description: Create or update evidence-based AGENTS.md instructions while preserving existing repository guidance and the shared Copilot and Claude harness.
---

# Create or update AGENTS.md

Use this skill when creating or maintaining repository instructions. `AGENTS.md` is a README for coding agents, not a replacement for human documentation.

## Inspect before writing

1. Read the existing `AGENTS.md` files, `CLAUDE.md`, Copilot instructions, agents and skills. Preserve existing guidance and reconcile contradictions explicitly rather than replacing it with a generic template.
2. Inspect the repository structure, dependency manifests, Makefile, tests and CI definitions. Derive commands from those sources, not from assumptions about other repositories.
3. Identify the project purpose, architecture, toolchain requirements, test strategy, formatting rules, public contracts and security boundaries.

## Write actionable instructions

- Describe the actual project and important directories.
- Document setup prerequisites, dependency installation, local development, build and test commands, including how to select focused tests.
- Explain style, error handling, documentation and backward-compatibility requirements.
- Describe build outputs and CI boundaries without authorizing deployment or infrastructure mutations.
- Include PR guidance only when supported by repository conventions or an explicit request.
- Use concise Markdown and exact commands. Mark unverified commands as unverified, and describe missing tooling honestly.
- Add nested instructions only when a subtree genuinely needs distinct guidance.

## Maintain the shared harness

- Keep repository rules in `AGENTS.md` and `CLAUDE.md` as only `@AGENTS.md`.
- Keep Copilot instructions as a pointer to `AGENTS.md`.
- Maintain agents and skills in `.github/`, with relative symbolic links in `.claude/`.
- Preserve existing customizations and third-party attribution; verify every link and check Git mode `120000` before publishing.
- Review the final instructions against the source files and execute applicable checks when available. Do not change unrelated application code to make a documentation check pass.

For Gotenberg, retain the established Makefile workflow and module/API compatibility rules. Do not import Terraform or AWS-specific implementation guidance from another project's harness.

Adapted from `github/awesome-copilot`; see [provenance and license](../../harness.md#third-party-provenance).
