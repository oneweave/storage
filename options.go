package storage

import (
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DefaultFindOptions returns standard find options with a default batch size of 100.
// Unlike the previous implementation, it does not project out _id, keeping standard
// MongoDB document behavior.
func DefaultFindOptions() []options.Lister[options.FindOptions] {
	opts := options.Find().SetBatchSize(100)
	return []options.Lister[options.FindOptions]{opts}
}
