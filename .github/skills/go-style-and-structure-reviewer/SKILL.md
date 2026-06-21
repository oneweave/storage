---
name: go-style-and-structure-reviewer
description: Review Go changes for idiomatic style and maintainable structure. Use this when performing a final quality pass on Go code.
---

# Purpose

Use this skill to improve readability and maintainability without changing behavior.

# Review Checklist

1. Verify naming is clear and descriptive.
2. Verify functions are small and focused.
3. Verify package boundaries are logical and avoid circular dependencies.
4. Verify exported symbols and complex logic have helpful comments.
5. Verify dependency injection is used instead of avoidable globals.
6. Verify code remains explicit and readable over clever shortcuts.
7. Verify error wrapping and context handling follow project rules.

# Output Format

When reporting, provide:

1. Findings ordered by severity with file and line references.
2. Any residual risks or testing gaps.
3. A brief change summary only after findings.

# Completion Criteria

Only declare completion when review findings are explicit, actionable, and aligned with repository coding conventions.
