# Oneweave Storage - Agent Guidelines

This is a lightweight, generic, type-safe MongoDB repository library for Go. Use this guide to understand key conventions and patterns.

## Project Overview

**Storage** abstracts MongoDB boilerplate (CRUD, querying, coercion, validation) while providing compile-time type safety through Go generics. The library:
- Uses MongoDB Go Driver v2 (`go.mongodb.org/mongo-driver/v2`)
- Requires Go 1.25.0+
- Has zero external dependencies (test-only: testify, mongo-driver)
- Enforces strict query security and idiomatic error handling

## Quick Start & Build Commands

```bash
# Run all tests (starts Docker MongoDB automatically)
./test.sh

# Start MongoDB manually (if needed)
docker compose -f mongodb-compose.yml up -d

# Credentials: admin/password
```

## Core Patterns & Conventions

### 1. Model Requirements

All models must have an `ID` field for public identification:

```go
type User struct {
    ID    string `bson:"id,omitempty"`  // Public ID (string UUID or hex)
    Name  string `bson:"name"`
    Email string `bson:"email"`
}
```

- MongoDB generates the internal `_id` automatically
- The struct's `ID` field maps to the separate BSON `id` field
- Reflection-based ID detection looks for `bson:"id"` tag or field named `ID`

### 2. Generic Repository Pattern

The `Storage[T]` type is the primary abstraction:

```go
userStorage := storage.NewStorage[User](db, "users")
```

- `NewStorage` reflects on the type `T` to detect ID field BSON name
- Works with `string` type for IDs
- All operations are type-safe and compile-time checked

### 3. Error Handling (Critical Convention)

**Always propagate errors using `fmt.Errorf` with `%w`:**

```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

- Use `errors.Is()` and `errors.As()` to check wrapped errors
- Sentinel error: `storage.ErrNotFound` for missing documents
- Never create new error types; wrap with context instead

### 4. Public ID Format

- Public IDs are strictly `string` types (e.g. UUID, custom string keys, or ObjectID hex string representation).
- Do not use BSON ObjectID fields directly for public IDs in the model.
- Keep generic decoding working without custom registries.

### 5. Query Security

Enforce strict equality matching (i.e., `$eq`) on all query fields:

```go
// ✓ Correct: strict equality
bson.M{"email": "user@example.com"}

// ✗ Wrong: operator-based queries (subject to NoSQL injection)
// Never pass user input directly into `$regex`, `$gt`, etc.
```

- See [mongodb-query-safety-and-input-validation](d:\oneweave-storage\.github\skills\mongodb-query-safety-and-input-validation\SKILL.md) for secure query patterns

### 6. Default Find Options

Use `DefaultFindOptions()` for standard batch sizes:

```go
opts := storage.DefaultFindOptions()  // Batch size: 100
// Pass to find operations
```

### 7. Fluent Query Building (QueryBuilder)

Use `QueryBuilder` to construct query filters with built-in input validation:
- Every operator method automatically validates parameters (rejects empty string, nil pointer, empty slices) to prevent NoSQL injection.
- Call `.Build()` to get the `bson.D` query and check the validation error.
- Call `.BuildMap()` to get `bson.M`.
- Consecutive calls for the same field (e.g., `.Gt("age", 18).Lt("age", 30)`) are merged automatically.

## Testing

Tests use Docker Compose for isolated MongoDB:

- `test.sh` auto-starts MongoDB if not running
- Credentials: `admin` / `password`
- Tests clean up Docker resources on exit
- Use `context.Background()` or `context.WithTimeout()` for test contexts

## File Organization

- `storage.go` – Core `Storage[T]` type and CRUD operations
- `query_builder.go` – Fluent MongoDB query builder with validation and merging
- `client.go` – MongoDB connection, transaction support, BSON registry
- `errors.go` – Sentinel errors
- `options.go` – Default MongoDB options
- `test_helper.go`, `storage_test.go`, `query_builder_test.go` – Test utilities and suites
- `mongodb-compose.yml` – Docker test environment
- `test.sh` – Bash test runner

## Style & Structure

Follow these Go conventions:

- **Receivers**: Use pointer receivers for methods that modify state
- **Context**: Always accept `context.Context` as first parameter
- **Package functions**: Exported functions (PascalCase) at package level for connection, registry, options
- **Generics**: Use `[T any]` for maximum compatibility; no constraints unless essential
- **Documentation**: Godoc comments on all exported functions and types
- **Error wrapping**: `fmt.Errorf(..., %w, ...)` for all error returns

## Development Tips

- Run tests frequently: `./test.sh`
- Check for compilation errors before committing
- Verify error handling on all MongoDB driver calls
- Use `reflect` carefully; prefer explicit struct tags over complex reflection
- Avoid external dependencies; keep core library lean
- Test with string ID types

## When to Use Related Skills

- **go-error-handling-enforcer** – When adding/refactoring functions with error returns
- **mongodb-query-safety-and-input-validation** – When writing or changing storage queries with user input
- **go-library-selection-guard** – When considering new dependencies (note: try to keep zero-dependency core)
- **conventional-commit-author-go** – When preparing commits after implementing changes

---

**For more details**, see [README.md](README.md) and inline function documentation.
