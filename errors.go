package storage

import "errors"

// ErrNotFound is returned when a document is not found in the storage.
var ErrNotFound = errors.New("document not found")
