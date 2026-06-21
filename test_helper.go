package storage

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SetupTestDB creates a temporary test database and returns a cleanup function.
// It resolves the MongoDB connection URI from the MONGODB_URI environment variable,
// defaulting to "mongodb://localhost:27017".
func SetupTestDB(t *testing.T) (*mongo.Database, func()) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		username := os.Getenv("OWV_MONGODB_USERNAME")
		password := os.Getenv("OWV_MONGODB_PASSWORD")
		if username != "" && password != "" {
			uri = fmt.Sprintf("mongodb://%s:%s@localhost:27017", username, password)
		} else {
			uri = "mongodb://localhost:27017"
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := Connect(ctx, uri)
	if err != nil {
		t.Fatalf("MongoDB not available for testing at %s: %v", uri, err)
		return nil, nil
	}

	randomNum, _ := rand.Int(rand.Reader, big.NewInt(999999))
	dbName := fmt.Sprintf("test_generic_%d", randomNum.Int64())
	db := client.Database(dbName)

	cleanup := func() {
		if db == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
	}

	return db, cleanup
}
