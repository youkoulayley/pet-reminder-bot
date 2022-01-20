package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func createStore(t *testing.T, reminds []Remind) *Store {
	t.Helper()

	ctx := context.Background()

	database := fmt.Sprint("petreminder-", time.Now().Nanosecond())
	uri := fmt.Sprintf("mongodb://mongoadmin:secret@127.0.0.1:27017/%s?authSource=admin", database)
	opts := options.Client().ApplyURI(uri)

	client, err := mongo.NewClient(opts)
	require.NoError(t, err)

	err = client.Connect(ctx)
	require.NoError(t, err)

	store := New(client, database)

	err = store.Bootstrap(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = store.client.Database(database).Drop(ctx)
		require.NoError(t, err)
	})

	for _, q := range reminds {
		_, err = store.reminds.InsertOne(ctx, q)
		require.NoError(t, err)
	}

	return store
}
