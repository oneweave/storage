---
name: go-handler-contract-kit-pattern
description: Implement HTTP handlers using go-api-kit and typed models. Use this when adding or refactoring API handlers.
---

# Purpose

Use this skill to keep handlers contract-first, typed, and consistent with go-api-kit patterns.

# Rules

- Use predefined api or domain models for request and response bodies.
- Do not use anonymous structs or map-based request and response contracts.
- Use stfsy/go-api-kit handler patterns, including validating handlers for request validation.
- Define dependency interfaces for storage and collaborating modules for testability.

# Required Workflow

1. Identify the endpoint contract and choose existing typed models, or add new ones.
2. Define required dependency interfaces in or near the handler.
3. Implement the endpoint with go-api-kit validating flow.
4. Ensure request validation and error responses are explicit.
5. Ensure response payloads remain typed and follow API response wrapper conventions.
6. Add or update handler tests for happy path, validation failure, dependency failure, and edge cases.

# Completion Criteria

Only declare completion when handlers are go-api-kit based, model-typed, interface-driven, and fully tested for success and failure paths.
