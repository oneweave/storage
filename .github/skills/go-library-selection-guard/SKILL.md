---
name: go-library-selection-guard
description: Enforce approved Go library choices. Use this when adding dependencies or selecting libraries for implementation tasks.
---

# Purpose

Use this skill to keep dependency choices aligned with repository standards.

# Preferred Libraries

- github.com/kelseyhightower/envconfig for environment configuration.
- github.com/stfsy/go-api-kit for HTTP handler configuration, middleware, and response handling.
- github.com/stfsy/go-api-key for API key creation and validation.
- github.com/stfsy/go-argon2id for secure hashing and hash verification.
- github.com/go-playground/validator for struct validation.
- github.com/stretchr/testify for test assertions.

# Required Workflow

1. Check whether requested functionality is already covered by approved libraries.
2. Prefer approved libraries before introducing new dependencies.
3. If a new dependency is still required, provide explicit rationale and tradeoffs.
4. Confirm import usage remains minimal and targeted.
5. Update tests to cover behavior introduced by selected dependencies.

# Completion Criteria

Only declare completion when dependency choices either use approved libraries or include a clear, justified exception.
