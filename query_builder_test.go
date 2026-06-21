package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestQueryBuilder_Comparison(t *testing.T) {
	qb := NewQueryBuilder().
		Eq("name", "Alice").
		Ne("status", "inactive").
		Gt("age", 21).
		Gte("score", 90).
		Lt("views", 1000).
		Lte("points", 500)

	filter, err := qb.Build()
	assert.NoError(t, err)

	expected := bson.D{
		{Key: "name", Value: bson.M{"$eq": "Alice"}},
		{Key: "status", Value: bson.M{"$ne": "inactive"}},
		{Key: "age", Value: bson.M{"$gt": 21}},
		{Key: "score", Value: bson.M{"$gte": 90}},
		{Key: "views", Value: bson.M{"$lt": 1000}},
		{Key: "points", Value: bson.M{"$lte": 500}},
	}
	assert.Equal(t, expected, filter)

	filterMap, err := qb.BuildMap()
	assert.NoError(t, err)
	assert.Equal(t, "Alice", filterMap["name"].(bson.M)["$eq"])
	assert.Equal(t, "inactive", filterMap["status"].(bson.M)["$ne"])
	assert.Equal(t, 21, filterMap["age"].(bson.M)["$gt"])
	assert.Equal(t, 90, filterMap["score"].(bson.M)["$gte"])
	assert.Equal(t, 1000, filterMap["views"].(bson.M)["$lt"])
	assert.Equal(t, 500, filterMap["points"].(bson.M)["$lte"])
}

func TestQueryBuilder_ArrayOperators(t *testing.T) {
	roles := []string{"admin", "user"}
	qb := NewQueryBuilder().
		In("role", roles).
		Nin("tags", []string{"banned", "spam"})

	filter, err := qb.Build()
	assert.NoError(t, err)

	expected := bson.D{
		{Key: "role", Value: bson.M{"$in": roles}},
		{Key: "tags", Value: bson.M{"$nin": []string{"banned", "spam"}}},
	}
	assert.Equal(t, expected, filter)
}

func TestQueryBuilder_Exists(t *testing.T) {
	qb := NewQueryBuilder().Exists("deleted_at", false)
	filter, err := qb.Build()
	assert.NoError(t, err)

	expected := bson.D{
		{Key: "deleted_at", Value: bson.M{"$exists": false}},
	}
	assert.Equal(t, expected, filter)
}

func TestQueryBuilder_Regex(t *testing.T) {
	qb := NewQueryBuilder().Regex("email", `.*@example\.com$`, "i")
	filter, err := qb.Build()
	assert.NoError(t, err)

	expected := bson.D{
		{Key: "email", Value: bson.M{"$regex": `.*@example\.com$`, "$options": "i"}},
	}
	assert.Equal(t, expected, filter)
}

func TestQueryBuilder_ElemMatch(t *testing.T) {
	subQB := NewQueryBuilder().Eq("name", "work").Exists("verified", true)
	qb := NewQueryBuilder().ElemMatch("addresses", subQB)

	filter, err := qb.Build()
	assert.NoError(t, err)

	subFilter, err := subQB.Build()
	assert.NoError(t, err)

	expected := bson.D{
		{Key: "addresses", Value: bson.M{"$elemMatch": subFilter}},
	}
	assert.Equal(t, expected, filter)
}

func TestQueryBuilder_Compound(t *testing.T) {
	qb1 := NewQueryBuilder().Eq("status", "active")
	qb2 := NewQueryBuilder().Gt("age", 18)

	qb := NewQueryBuilder().Or(qb1, qb2)
	filter, err := qb.Build()
	assert.NoError(t, err)

	f1, _ := qb1.Build()
	f2, _ := qb2.Build()

	expected := bson.D{
		{Key: "$or", Value: []bson.D{f1, f2}},
	}
	assert.Equal(t, expected, filter)
}

func TestQueryBuilder_Merging(t *testing.T) {
	qb := NewQueryBuilder().
		Gt("age", 18).
		Lt("age", 30)

	filter, err := qb.Build()
	assert.NoError(t, err)

	expected := bson.D{
		{Key: "age", Value: bson.M{"$gt": 18, "$lt": 30}},
	}
	assert.Equal(t, expected, filter)
}

func TestQueryBuilder_ValidationRejection(t *testing.T) {
	t.Run("empty field", func(t *testing.T) {
		_, err := NewQueryBuilder().Eq("", "val").Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field name cannot be empty")
	})

	t.Run("nil value", func(t *testing.T) {
		_, err := NewQueryBuilder().Eq("name", nil).Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value cannot be nil")
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := NewQueryBuilder().Eq("name", "").Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "string value cannot be empty")
	})

	t.Run("empty slice", func(t *testing.T) {
		_, err := NewQueryBuilder().In("roles", []string{}).Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "slice/array value cannot be empty or nil")
	})

	t.Run("nil slice pointer", func(t *testing.T) {
		var nilSlice []string
		_, err := NewQueryBuilder().In("roles", nilSlice).Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "slice/array value cannot be empty or nil")
	})

	t.Run("nil element in slice", func(t *testing.T) {
		vals := []interface{}{"admin", nil}
		_, err := NewQueryBuilder().In("roles", vals).Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("empty string element in slice", func(t *testing.T) {
		vals := []string{"admin", ""}
		_, err := NewQueryBuilder().In("roles", vals).Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be an empty string")
	})

	t.Run("invalid regex", func(t *testing.T) {
		_, err := NewQueryBuilder().Regex("email", "[a-z", "").Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pattern")
	})

	t.Run("non-slice passed to In", func(t *testing.T) {
		_, err := NewQueryBuilder().In("role", "admin").Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a slice or array")
	})

	t.Run("nil builder in compound", func(t *testing.T) {
		_, err := NewQueryBuilder().And(nil).Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sub-builder at index 0 cannot be nil")
	})

	t.Run("failed builder in compound", func(t *testing.T) {
		bad := NewQueryBuilder().Eq("name", "")
		_, err := NewQueryBuilder().And(bad).Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed")
	})

	t.Run("no builders in compound", func(t *testing.T) {
		_, err := NewQueryBuilder().And().Build()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires at least one sub-builder")
	})
}

func TestQueryBuilder_Integration(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	s := NewStorage[UserWithStringID](db, "users_qb_test")
	ctx := context.Background()

	users := []*UserWithStringID{
		{ID: "u1", Name: "Alice", Email: "alice@example.com"},
		{ID: "u2", Name: "Bob", Email: "bob@example.com"},
		{ID: "u3", Name: "Charlie", Email: "charlie@gmail.com"},
	}

	for _, u := range users {
		err := s.Create(ctx, u)
		assert.NoError(t, err)
	}

	t.Run("find by name and email domain", func(t *testing.T) {
		qb := NewQueryBuilder().
			Regex("email", `.*@example\.com$`, "").
			Ne("name", "Bob")

		filter, err := qb.Build()
		assert.NoError(t, err)

		results, err := s.Find(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Alice", results[0].Name)
	})
}
