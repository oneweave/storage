---
name: go-error-handling-enforcer
description: Enforce Go error handling conventions. Use this when adding or refactoring Go code that calls functions returning errors.
---

# Purpose

Use this skill to keep error handling explicit, wrapped, and actionable.

# Rules

- Always check returned errors immediately after the call.
- Always wrap propagated errors with fmt.Errorf and %w.
- Include useful context in error messages, without secrets.
- Return correct zero values with wrapped errors.
- Do not use pre-declare and assign-later error patterns.

# Required Workflow

1. Find all new or changed function calls that return an error.
2. Convert each call to the two-line pattern:
   - result, err := call(...)
   - if err != nil { return ..., fmt.Errorf("operation context: %w", err) }
3. Ensure every wrapper message states the failed operation and non-sensitive identifier when useful.
4. Verify no nested or compact error idioms reduce readability.
5. Verify returned non-error values are proper zero values for each error path.

# Completion Criteria

Only declare completion when all changed call sites either handle errors locally or return wrapped errors with %w and useful context.
