package storage

import (
	"fmt"
	"reflect"
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// QueryBuilder is a fluent, type-safe builder for constructing MongoDB query filters.
type QueryBuilder struct {
	filter bson.D
	err    error
}

// NewQueryBuilder initializes a new QueryBuilder instance.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		filter: bson.D{},
	}
}

// validateFieldAndValue ensures that the query field is not empty, the value is not nil,
// and any strings/slices/arrays are not empty, helping prevent NoSQL injection and invalid queries.
func validateFieldAndValue(field string, value interface{}) error {
	if field == "" {
		return fmt.Errorf("field name cannot be empty")
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil for field %q", field)
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		if val.String() == "" {
			return fmt.Errorf("string value cannot be empty for field %q", field)
		}
	case reflect.Slice, reflect.Array:
		if val.IsNil() || val.Len() == 0 {
			return fmt.Errorf("slice/array value cannot be empty or nil for field %q", field)
		}
	case reflect.Pointer:
		if val.IsNil() {
			return fmt.Errorf("pointer value cannot be nil for field %q", field)
		}
		// Recursively validate the dereferenced value
		return validateFieldAndValue(field, val.Elem().Interface())
	}
	return nil
}

// append adds or merges a query key and its operator value in the internal bson.D filter.
// If the key already exists and both the existing and new values are bson.M, it merges them.
func (qb *QueryBuilder) append(key string, val interface{}) {
	for i, elem := range qb.filter {
		if elem.Key == key {
			existingMap, ok1 := elem.Value.(bson.M)
			newMap, ok2 := val.(bson.M)
			if ok1 && ok2 {
				for k, v := range newMap {
					existingMap[k] = v
				}
				qb.filter[i].Value = existingMap
				return
			}
		}
	}
	qb.filter = append(qb.filter, bson.E{Key: key, Value: val})
}

// Eq adds an equality condition: {field: {"$eq": value}}
func (qb *QueryBuilder) Eq(field string, value interface{}) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if err := validateFieldAndValue(field, value); err != nil {
		qb.err = fmt.Errorf("Eq validation failed: %w", err)
		return qb
	}
	qb.append(field, bson.M{"$eq": value})
	return qb
}

// Ne adds a not-equal condition: {field: {"$ne": value}}
func (qb *QueryBuilder) Ne(field string, value interface{}) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if err := validateFieldAndValue(field, value); err != nil {
		qb.err = fmt.Errorf("Ne validation failed: %w", err)
		return qb
	}
	qb.append(field, bson.M{"$ne": value})
	return qb
}

// Gt adds a greater-than condition: {field: {"$gt": value}}
func (qb *QueryBuilder) Gt(field string, value interface{}) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if err := validateFieldAndValue(field, value); err != nil {
		qb.err = fmt.Errorf("Gt validation failed: %w", err)
		return qb
	}
	qb.append(field, bson.M{"$gt": value})
	return qb
}

// Gte adds a greater-than-or-equal condition: {field: {"$gte": value}}
func (qb *QueryBuilder) Gte(field string, value interface{}) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if err := validateFieldAndValue(field, value); err != nil {
		qb.err = fmt.Errorf("Gte validation failed: %w", err)
		return qb
	}
	qb.append(field, bson.M{"$gte": value})
	return qb
}

// Lt adds a less-than condition: {field: {"$lt": value}}
func (qb *QueryBuilder) Lt(field string, value interface{}) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if err := validateFieldAndValue(field, value); err != nil {
		qb.err = fmt.Errorf("Lt validation failed: %w", err)
		return qb
	}
	qb.append(field, bson.M{"$lt": value})
	return qb
}

// Lte adds a less-than-or-equal condition: {field: {"$lte": value}}
func (qb *QueryBuilder) Lte(field string, value interface{}) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if err := validateFieldAndValue(field, value); err != nil {
		qb.err = fmt.Errorf("Lte validation failed: %w", err)
		return qb
	}
	qb.append(field, bson.M{"$lte": value})
	return qb
}

// In adds an in-array condition: {field: {"$in": [values]}}
// It validates that values is a non-empty slice or array, and that no element is nil or empty string.
func (qb *QueryBuilder) In(field string, values interface{}) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if err := validateFieldAndValue(field, values); err != nil {
		qb.err = fmt.Errorf("In validation failed: %w", err)
		return qb
	}

	val := reflect.ValueOf(values)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		qb.err = fmt.Errorf("In validation failed: values must be a slice or array, got %s", val.Kind())
		return qb
	}

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i).Interface()
		if elem == nil {
			qb.err = fmt.Errorf("In validation failed: element at index %d cannot be nil", i)
			return qb
		}
		if str, ok := elem.(string); ok && str == "" {
			qb.err = fmt.Errorf("In validation failed: element at index %d cannot be an empty string", i)
			return qb
		}
	}

	qb.append(field, bson.M{"$in": values})
	return qb
}

// Nin adds a not-in-array condition: {field: {"$nin": [values]}}
// It validates that values is a non-empty slice or array, and that no element is nil or empty string.
func (qb *QueryBuilder) Nin(field string, values interface{}) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if err := validateFieldAndValue(field, values); err != nil {
		qb.err = fmt.Errorf("Nin validation failed: %w", err)
		return qb
	}

	val := reflect.ValueOf(values)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		qb.err = fmt.Errorf("Nin validation failed: values must be a slice or array, got %s", val.Kind())
		return qb
	}

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i).Interface()
		if elem == nil {
			qb.err = fmt.Errorf("Nin validation failed: element at index %d cannot be nil", i)
			return qb
		}
		if str, ok := elem.(string); ok && str == "" {
			qb.err = fmt.Errorf("Nin validation failed: element at index %d cannot be an empty string", i)
			return qb
		}
	}

	qb.append(field, bson.M{"$nin": values})
	return qb
}

// Exists adds a field existence condition: {field: {"$exists": exists}}
func (qb *QueryBuilder) Exists(field string, exists bool) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if field == "" {
		qb.err = fmt.Errorf("Exists validation failed: field name cannot be empty")
		return qb
	}
	qb.append(field, bson.M{"$exists": exists})
	return qb
}

// Regex adds a regular expression condition: {field: {"$regex": pattern, "$options": options}}
// It validates that the pattern compiles successfully.
func (qb *QueryBuilder) Regex(field string, pattern string, options string) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if field == "" {
		qb.err = fmt.Errorf("Regex validation failed: field name cannot be empty")
		return qb
	}
	if pattern == "" {
		qb.err = fmt.Errorf("Regex validation failed: pattern cannot be empty for field %q", field)
		return qb
	}
	if _, err := regexp.Compile(pattern); err != nil {
		qb.err = fmt.Errorf("Regex validation failed: invalid pattern %q for field %q: %w", pattern, field, err)
		return qb
	}

	qb.append(field, bson.M{"$regex": pattern, "$options": options})
	return qb
}

// ElemMatch adds an element matching condition for array fields: {field: {"$elemMatch": subFilter}}
func (qb *QueryBuilder) ElemMatch(field string, builder *QueryBuilder) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if field == "" {
		qb.err = fmt.Errorf("ElemMatch validation failed: field name cannot be empty")
		return qb
	}
	if builder == nil {
		qb.err = fmt.Errorf("ElemMatch validation failed: sub-builder cannot be nil for field %q", field)
		return qb
	}

	subFilter, err := builder.Build()
	if err != nil {
		qb.err = fmt.Errorf("ElemMatch validation failed: sub-builder for field %q failed: %w", field, err)
		return qb
	}

	qb.append(field, bson.M{"$elemMatch": subFilter})
	return qb
}

// And combines multiple sub-builders with a logical AND: {"$and": [subFilters...]}
func (qb *QueryBuilder) And(builders ...*QueryBuilder) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if len(builders) == 0 {
		qb.err = fmt.Errorf("And validation failed: requires at least one sub-builder")
		return qb
	}

	var subFilters []bson.D
	for i, subQB := range builders {
		if subQB == nil {
			qb.err = fmt.Errorf("And validation failed: sub-builder at index %d cannot be nil", i)
			return qb
		}
		subFilter, err := subQB.Build()
		if err != nil {
			qb.err = fmt.Errorf("And validation failed: sub-builder at index %d failed: %w", i, err)
			return qb
		}
		subFilters = append(subFilters, subFilter)
	}

	qb.append("$and", subFilters)
	return qb
}

// Or combines multiple sub-builders with a logical OR: {"$or": [subFilters...]}
func (qb *QueryBuilder) Or(builders ...*QueryBuilder) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if len(builders) == 0 {
		qb.err = fmt.Errorf("Or validation failed: requires at least one sub-builder")
		return qb
	}

	var subFilters []bson.D
	for i, subQB := range builders {
		if subQB == nil {
			qb.err = fmt.Errorf("Or validation failed: sub-builder at index %d cannot be nil", i)
			return qb
		}
		subFilter, err := subQB.Build()
		if err != nil {
			qb.err = fmt.Errorf("Or validation failed: sub-builder at index %d failed: %w", i, err)
			return qb
		}
		subFilters = append(subFilters, subFilter)
	}

	qb.append("$or", subFilters)
	return qb
}

// Build returns the constructed bson.D query filter.
// It returns an error if any of the builder method validation checks failed.
func (qb *QueryBuilder) Build() (bson.D, error) {
	if qb.err != nil {
		return nil, qb.err
	}
	return qb.filter, nil
}

// BuildMap returns the constructed bson.M query filter.
// It returns an error if any of the builder method validation checks failed.
func (qb *QueryBuilder) BuildMap() (bson.M, error) {
	if qb.err != nil {
		return nil, qb.err
	}
	m := bson.M{}
	for _, elem := range qb.filter {
		m[elem.Key] = elem.Value
	}
	return m, nil
}
