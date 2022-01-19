package store

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

// NotFoundError represents a document not found error.
type NotFoundError struct {
	Err error
}

// Error stringifies the error.
func (e NotFoundError) Error() string {
	if e.Err == nil {
		return "resource not found"
	}

	return fmt.Sprintf("not found: %v", e.Err)
}

// Unwrap returns the underlying error.
func (e NotFoundError) Unwrap() error { return e.Err }

func isMongoDBDuplicateError(err error) bool {
	var (
		writeException mongo.WriteException
		commandError   mongo.CommandError
	)

	if errors.As(err, &writeException) {
		for _, we := range writeException.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}

	if errors.As(err, &commandError) {
		return commandError.Code == 11000
	}

	return false
}
