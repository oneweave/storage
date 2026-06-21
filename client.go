package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	mongoClientInstance *mongo.Client
	mongoClientOnce     sync.Once
	mongoClientError    error
)

// TransactionFunc represents a function to be executed within a database transaction.
type TransactionFunc func(ctx context.Context) error

// Connect initializes and returns a mongo.Client with custom options.
// It supports configuring TLS via the OWV_MONGODB_TLS environment variable:
// - If OWV_MONGODB_TLS="true", it enables TLS.
// - If OWV_MONGODB_TLS_INSECURE="true", it skips TLS certificate verification.
func Connect(ctx context.Context, uri string, opts ...*options.ClientOptions) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(uri)

	// Configure TLS if specified by environment variable
	if os.Getenv("OWV_MONGODB_TLS") == "true" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: os.Getenv("OWV_MONGODB_TLS_INSECURE") == "true",
		}
		clientOpts.SetTLSConfig(tlsConfig)
	}

	// Combine our clientOpts and user provided opts
	allOpts := append([]*options.ClientOptions{clientOpts}, opts...)

	// Connect in MongoDB v2 does not accept a context parameter, only client options.
	client, err := mongo.Connect(allOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// Ping the server to test connection (using the context for timeout/cancel)
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	return client, nil
}

// InitMongoClient connects to a global client instance (similar to the legacy setup)
func InitMongoClient(uri string) (*mongo.Client, error) {
	mongoClientOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		mongoClientInstance, mongoClientError = Connect(ctx, uri)
	})
	return mongoClientInstance, mongoClientError
}

// RunTransaction executes the provided function inside a MongoDB transaction.
// It handles session creation, transaction initiation, commit, and abort automatically.
// The provided TransactionFunc must use the transaction context (sessCtx) passed into it
// for all database operations to participate in the transaction.
func RunTransaction(ctx context.Context, client *mongo.Client, fn TransactionFunc) error {
	if client == nil {
		return fmt.Errorf("refusing to run transaction: client cannot be nil")
	}
	if fn == nil {
		return fmt.Errorf("refusing to run transaction: transaction function cannot be nil")
	}

	session, err := client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// WithTransaction automatically commits the transaction if the function returns nil,
	// aborts it if an error is returned, and retries transient transaction failures.
	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (interface{}, error) {
		err := fn(sessCtx)
		if err != nil {
			return nil, fmt.Errorf("transaction execution error: %w", err)
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}
