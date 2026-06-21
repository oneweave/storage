---
name: mongodb-query-safety-and-input-validation
description: Secure MongoDB query construction for user input. Use this when writing or changing storage queries that include user-provided values.
---

# Purpose

Use this skill to prevent NoSQL injection and invalid query execution.

# Rules

- Validate user-provided query inputs before building filters.
- Reject empty strings and nil pointers for required query fields.
- Use strict equality matching with $eq for user input values.
- Do not pass user values directly as unconstrained MongoDB filter expressions.

# Required Workflow

1. Identify all user-controlled inputs used in query filters.
2. Add explicit guard checks for empty and nil values before query construction.
3. Return wrapped errors with actionable context when validation fails.
4. Build filters with strict equality, for example:
   - bson.M{"field": bson.M{"$eq": value}}
5. Add or update tests for:
   - Valid input paths.
   - Empty or nil input rejection.
   - Correct safe filter behavior.

# Completion Criteria

Only declare completion when all affected MongoDB filters validate input first and enforce strict equality matching for user-provided values.
