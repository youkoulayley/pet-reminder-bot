package store

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

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
