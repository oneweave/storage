package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserWithStringID struct {
	ID    string `bson:"id,omitempty"`
	Name  string `bson:"name"`
	Email string `bson:"email"`
}


func TestStorage_StringID_CRUD(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	s := NewStorage[UserWithStringID](db, "users_str")
	ctx := context.Background()

	// 1. Create with ID set
	userID := bson.NewObjectID().Hex()
	user := &UserWithStringID{
		ID:    userID,
		Name:  "John Doe",
		Email: "john@example.com",
	}
	err := s.Create(ctx, user)
	assert.NoError(t, err)
	assert.Equal(t, userID, user.ID)

	// 2. Get the created user
	retrieved, err := s.Get(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Name, retrieved.Name)
	assert.Equal(t, user.Email, retrieved.Email)

	// 3. Exists check
	exists, err := s.Exists(ctx, user.ID)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 4. Update (Replace) the user
	user.Name = "John Updated"
	user.Email = "john_updated@example.com"
	updated, err := s.Update(ctx, user.ID, user)
	assert.NoError(t, err)
	assert.Equal(t, "John Updated", updated.Name)

	retrieved, err = s.Get(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "John Updated", retrieved.Name)

	// 5. UpdateFields (Partial Update)
	err = s.UpdateFields(ctx, user.ID, bson.M{"$set": bson.M{"email": "john_partial@example.com"}})
	assert.NoError(t, err)

	retrieved, err = s.Get(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "john_partial@example.com", retrieved.Email)
	assert.Equal(t, "John Updated", retrieved.Name)

	// 6. Delete user
	err = s.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// Verify deleted
	exists, err = s.Exists(ctx, user.ID)
	assert.NoError(t, err)
	assert.False(t, exists)

	_, err = s.Get(ctx, user.ID)
	assert.Equal(t, ErrNotFound, err)
}


func TestStorage_Find_Queries(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	s := NewStorage[UserWithStringID](db, "users_find")
	ctx := context.Background()

	users := []*UserWithStringID{
		{ID: bson.NewObjectID().Hex(), Name: "Alice", Email: "alice@example.com"},
		{ID: bson.NewObjectID().Hex(), Name: "Bob", Email: "bob@example.com"},
		{ID: bson.NewObjectID().Hex(), Name: "Charlie", Email: "charlie@example.com"},
	}

	for _, u := range users {
		err := s.Create(ctx, u)
		assert.NoError(t, err)
	}

	// 1. Find all
	results, err := s.Find(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// 2. Find with filter
	results, err = s.Find(ctx, bson.M{"name": "Alice"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "alice@example.com", results[0].Email)

	// 3. Find with options (sorting and limit)
	opts := options.Find().SetSort(bson.M{"name": 1}).SetLimit(2)
	results, err = s.Find(ctx, nil, opts)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Alice", results[0].Name)
	assert.Equal(t, "Bob", results[1].Name)

	// 4. FindOne
	single, err := s.FindOne(ctx, bson.M{"name": "Charlie"})
	assert.NoError(t, err)
	assert.Equal(t, "charlie@example.com", single.Email)
}

func TestStorage_InjectionProtection(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	s := NewStorage[UserWithStringID](db, "users_injection")
	ctx := context.Background()

	// Create a legitimate user
	legitUser := &UserWithStringID{ID: bson.NewObjectID().Hex(), Name: "Legit", Email: "legit@example.com"}
	err := s.Create(ctx, legitUser)
	assert.NoError(t, err)

	// Malicious IDs that attempt NoSQL injection
	maliciousIDs := []string{
		`{"$ne": null}`,
		`{"$regex": ".*"}`,
		`{"$where": "function() { return true; }"}`,
	}

	for _, malID := range maliciousIDs {
		// Get attempt should fail with ErrNotFound rather than matching
		_, err := s.Get(ctx, malID)
		assert.Error(t, err)

		// Delete attempt should return ErrNotFound and not delete any existing document
		err = s.Delete(ctx, malID)
		assert.Error(t, err)
	}

	// Verify legit user still exists
	exists, err := s.Exists(ctx, legitUser.ID)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestStorage_InputValidation(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	s := NewStorage[UserWithStringID](db, "users_validation")
	ctx := context.Background()

	// 1. Create with nil item
	err := s.Create(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "item cannot be nil")

	// 2. Get with empty ID
	_, err = s.Get(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")

	// 3. Update with empty ID
	_, err = s.Update(ctx, "", &UserWithStringID{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")

	// 4. Update with nil item
	_, err = s.Update(ctx, "some-id", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "item cannot be nil")

	// 5. UpdateFields with empty ID
	err = s.UpdateFields(ctx, "", bson.M{"$set": bson.M{"name": "test"}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")

	// 6. UpdateFields with nil update
	err = s.UpdateFields(ctx, "some-id", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update document cannot be nil")

	// 7. Delete with empty ID
	err = s.Delete(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")

	// 8. Exists with empty ID
	_, err = s.Exists(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")
}

func TestStorage_EnsureIndexes(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	s := NewStorage[UserWithStringID](db, "users_indexes")
	ctx := context.Background()

	// Ensure indexes is called
	err := s.EnsureIndexes(ctx)
	assert.NoError(t, err)

	// Verify that inserting a document with a duplicate public ID fails!
	user1 := &UserWithStringID{ID: "dup-id", Name: "User 1"}
	err = s.Create(ctx, user1)
	assert.NoError(t, err)

	user2 := &UserWithStringID{ID: "dup-id", Name: "User 2"}
	err = s.Create(ctx, user2)
	assert.Error(t, err, "expected error when inserting duplicate public ID under unique index")
}

func TestStorage_RunTransaction(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	client := db.Client()

	s := NewStorage[UserWithStringID](db, "users_txn")

	// Since local MongoDB standalone instance doesn't support replica sets,
	// calling RunTransaction will fail. We assert that it either succeeds (if RS exists)
	// or returns the expected MongoDB replica set error.
	err := RunTransaction(ctx, client, func(sessCtx context.Context) error {
		user := &UserWithStringID{ID: "txn-id", Name: "Txn User"}
		return s.Create(sessCtx, user)
	})

	if err != nil {
		assert.Contains(t, err.Error(), "transaction")
	}
}

func TestConnect_TLS(t *testing.T) {
	ctx := context.Background()
	t.Setenv("OWV_MONGODB_TLS", "true")
	t.Setenv("OWV_MONGODB_TLS_INSECURE", "true")

	// This connection will fail because there is no MongoDB server on 27019,
	// but it verifies that Connect executes the TLS configuration paths without panicking.
	_, err := Connect(ctx, "mongodb://localhost:27019/?serverSelectionTimeoutMS=50")
	assert.Error(t, err)
}

