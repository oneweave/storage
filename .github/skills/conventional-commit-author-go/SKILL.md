---
name: conventional-commit-author-go
description: Draft Conventional Commit messages for Go repository changes. Use this when preparing commits after implementing code updates.
---

# Purpose

Use this skill to produce concise, standards-compliant commit messages.

# Rules

- Follow format: type(scope): description.
- Keep description imperative and concise.
- Add body only when it helps explain what and why.
- Add footer references when issues or PR links are relevant.

# Required Workflow

1. Inspect changed files and classify change intent.
2. Select the most accurate commit type:
   - feat, fix, docs, style, refactor, perf, test, chore
3. Choose a focused scope reflecting the subsystem.
4. Draft a concise imperative subject line.
5. Add optional body and footer only if needed for clarity.

# Completion Criteria

Only declare completion when the proposed commit message follows Conventional Commits and accurately reflects the code changes.
