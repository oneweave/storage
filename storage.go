package storage

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Storage is a generic repository implementation for MongoDB collections.
// T is any struct that represents a collection document.
type Storage[T any] struct {
	collection *mongo.Collection
	bsonName   string
}

// NewStorage creates a new generic Storage instance for the given type T.
// It reflects on T to find the field annotated with bson:"id" or named "ID"
// to detect its BSON field name.
func NewStorage[T any](db *mongo.Database, collectionName string) *Storage[T] {
	s := &Storage[T]{
		collection: db.Collection(collectionName),
	}

	// Reflect to find ID field BSON name
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Struct {
		// 1. Check BSON tag "id"
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			bsonTag := field.Tag.Get("bson")
			parts := strings.Split(bsonTag, ",")
			if len(parts) > 0 && parts[0] == "id" {
				s.bsonName = "id"
				break
			}
		}

		// 2. Fallback to field name "ID"
		if s.bsonName == "" {
			field, found := t.FieldByName("ID")
			if found {
				// Determine BSON name for "ID"
				bsonTag := field.Tag.Get("bson")
				parts := strings.Split(bsonTag, ",")
				if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
					s.bsonName = parts[0]
				} else {
					s.bsonName = "id" // default lowercased
				}
			}
		}
	}

	if s.bsonName == "" {
		s.bsonName = "id" // fallback default
	}

	return s
}

// GetCollection returns the raw mongo.Collection instance.
func (s *Storage[T]) GetCollection() *mongo.Collection {
	return s.collection
}

// EnsureIndexes creates a unique index on the public ID field to guarantee
// O(1) query performance and enforce uniqueness at the database level.
func (s *Storage[T]) EnsureIndexes(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("refusing to ensure indexes: context cannot be nil")
	}

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: s.bsonName, Value: int32(1)}},
		Options: options.Index().SetUnique(true).SetName("unique_public_id_" + s.bsonName),
	}

	_, err := s.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create unique index on field %q: %w", s.bsonName, err)
	}

	return nil
}

// Create inserts a document into MongoDB.
func (s *Storage[T]) Create(ctx context.Context, item *T) error {
	if item == nil {
		return fmt.Errorf("refusing to create document: item cannot be nil")
	}

	_, err := s.collection.InsertOne(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}
	return nil
}

// Get retrieves a document by its string ID representation.
func (s *Storage[T]) Get(ctx context.Context, id string) (*T, error) {
	if id == "" {
		return nil, fmt.Errorf("refusing to get document: ID cannot be empty")
	}

	var result T
	filter := bson.M{s.bsonName: bson.M{"$eq": id}}
	err := s.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get document from collection: %w", err)
	}

	return &result, nil
}

// Update replaces the entire document with the given ID.
func (s *Storage[T]) Update(ctx context.Context, id string, item *T) (*T, error) {
	if id == "" {
		return nil, fmt.Errorf("refusing to update document: ID cannot be empty")
	}
	if item == nil {
		return nil, fmt.Errorf("refusing to update document: item cannot be nil")
	}

	filter := bson.M{s.bsonName: bson.M{"$eq": id}}
	result, err := s.collection.ReplaceOne(ctx, filter, item)
	if err != nil {
		return nil, fmt.Errorf("failed to replace document: %w", err)
	}
	if result.MatchedCount == 0 {
		return nil, ErrNotFound
	}

	return item, nil
}

// UpdateFields performs a partial update of specific fields using standard update syntax.
func (s *Storage[T]) UpdateFields(ctx context.Context, id string, update interface{}) error {
	if id == "" {
		return fmt.Errorf("refusing to update document fields: ID cannot be empty")
	}
	if update == nil {
		return fmt.Errorf("refusing to update document fields: update document cannot be nil")
	}

	filter := bson.M{s.bsonName: bson.M{"$eq": id}}
	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document fields in collection: %w", err)
	}
	if result.MatchedCount == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete removes a document by its ID.
func (s *Storage[T]) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("refusing to delete document: ID cannot be empty")
	}

	filter := bson.M{s.bsonName: bson.M{"$eq": id}}
	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete document from collection: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

// Exists checks whether a document with the given ID exists.
func (s *Storage[T]) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, fmt.Errorf("refusing to check document existence: ID cannot be empty")
	}

	filter := bson.M{s.bsonName: bson.M{"$eq": id}}
	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check document existence in collection: %w", err)
	}

	return count > 0, nil
}

// Find executes a query and returns multiple documents.
func (s *Storage[T]) Find(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) ([]*T, error) {
	if filter == nil {
		filter = bson.M{}
	}

	cursor, err := s.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*T
	for cursor.Next(ctx) {
		var item T
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		results = append(results, &item)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error during find operation: %w", err)
	}

	return results, nil
}

// FindOne executes a query and returns a single matching document.
func (s *Storage[T]) FindOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOneOptions]) (*T, error) {
	if filter == nil {
		filter = bson.M{}
	}

	var result T
	err := s.collection.FindOne(ctx, filter, opts...).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find document matching filter: %w", err)
	}

	return &result, nil
}
