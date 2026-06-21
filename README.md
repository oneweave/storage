# Storage - Generic Go MongoDB Storage Module

`storage` is a lightweight, generic, type-safe MongoDB repository library for Go, built on top of the official MongoDB Go Driver v2.

It abstracts away boilerplate database operations (CRUD, simple querying, and input validation) while providing compile-time type safety through Go generics.

## Features

- **Generic Repository (`Storage[T]`)**: Simplifies database operations for any struct document model.
- **MongoDB-Generated `_id` & Public `id`**: MongoDB automatically generates the database `_id` primary key, while the Go model's `ID` field maps to a separate BSON field (`id`), serving as the public identifier.
- **Strict Query Security**: Enforces strict equality matching (`$eq`) on query fields to protect against NoSQL injection.
- **Robust Error Handling**: Automatically wraps errors using `fmt.Errorf` with `%w` for idiomatic error propagation.
- **Zero-Dependency Core**: Free of dependencies on external config or error modules, making it completely reusable.

---

## Installation

Add the package to your Go module dependencies:

```bash
go get github.com/oneweave/storage
```

---

## Struct Model Requirements

Your models can be any plain Go structs. To map the identifier correctly, the struct must have a field named `ID` or tagged with `bson:"id"`. This field acts as the public ID. MongoDB will automatically generate the database `_id` field.

```go
package domain

type User struct {
	ID    string `bson:"id,omitempty"` // The public ID (e.g. a string UUID or hex)
	Name  string `bson:"name"`
	Email string `bson:"email"`
}
```

---

## Usage Guide

### 1. Connection Setup

Connect to MongoDB using the library's `Connect` function:

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/oneweave/storage"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := storage.Connect(ctx, "mongodb://localhost:27017")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("my_database")

	// Get a storage instance
	userStorage := storage.NewStorage[User](db, "users")
}
```

#### Environment Variables for Connection Configuration

You can configure connection parameters (such as TLS) via environment variables when calling `Connect`:

- `OWV_MONGODB_TLS`: Set to `true` to enable TLS communication by default.
- `OWV_MONGODB_TLS_INSECURE`: Set to `true` to skip TLS certificate verification (useful for self-signed certificates in dev/staging environments).

For test execution, the test suite also reads credentials from:
- `OWV_MONGODB_USERNAME`: The database user (defaults to `admin` in tests).
- `OWV_MONGODB_PASSWORD`: The database password (defaults to `password` in tests).



### 2. CRUD Operations

```go
ctx := context.Background()

// 1. Create (expect ID to be pre-set on the struct)
user := &User{
	ID:    "user-1234",
	Name:  "Jane Doe",
	Email: "jane@example.com",
}
if err := userStorage.Create(ctx, user); err != nil {
	log.Fatalf("Create failed: %v", err)
}

// 2. Get (retrieves by the public ID)
retrieved, err := userStorage.Get(ctx, "user-1234")
if err != nil {
	if err == storage.ErrNotFound {
		log.Println("User not found")
	} else {
		log.Fatalf("Get failed: %v", err)
	}
}

// 3. Exists
exists, err := userStorage.Exists(ctx, "user-1234")
if exists {
	log.Println("User exists!")
}

// 4. Update (completely replaces the document matching public ID)
retrieved.Name = "Jane Smith"
updated, err := userStorage.Update(ctx, retrieved.ID, retrieved)

// 5. UpdateFields (partial update using standard MongoDB update document)
err = userStorage.UpdateFields(ctx, "user-1234", bson.M{"$set": bson.M{"email": "jane.smith@example.com"}})

// 6. Delete
err = userStorage.Delete(ctx, "user-1234")
```

### 3. Generic Querying

Use `Find` or `FindOne` to run custom queries:

```go
import "go.mongodb.org/mongo-driver/v2/mongo/options"

// Find multiple documents
filter := bson.M{"name": "Jane Smith"}
opts := options.Find().SetLimit(10)

users, err := userStorage.Find(ctx, filter, opts)
if err != nil {
	log.Fatalf("Query failed: %v", err)
}

// Find a single document
user, err := userStorage.FindOne(ctx, bson.M{"email": "jane.smith@example.com"})
```### 4. Unique Index Management

To guarantee O(1) query performance and enforce uniqueness on your public ID field at the database level, call `EnsureIndexes`:

```go
// Creates a unique index on the public ID field (e.g. "id")
if err := userStorage.EnsureIndexes(ctx); err != nil {
	log.Fatalf("Index creation failed: %v", err)
}
```

### 5. Transactions

To execute multiple operations atomically across different collections or repositories, use `RunTransaction`:

```go
err := storage.RunTransaction(ctx, client, func(sessCtx context.Context) error {
	// 1. Create a user
	user := &User{ID: "user-999", Name: "Bob"}
	if err := userStorage.Create(sessCtx, user); err != nil {
		return err
	}

	// 2. Perform another action (e.g., in a different repository)
	// if err := logStorage.Create(sessCtx, logEntry); err != nil { return err }

	return nil // Commits transaction automatically
})
if err != nil {
	log.Fatalf("Transaction failed and rolled back: %v", err)
}
```

---



## Errors

The module returns the following package-level error when no documents match an operation:

```go
var ErrNotFound = errors.New("document not found")
```
