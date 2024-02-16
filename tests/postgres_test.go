package tests

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/kyuff/dbleases/internal/assert"
	"github.com/kyuff/dbleases/internal/schemas/postgres"
)

func TestPostgres(t *testing.T) {

	var (
		ctx               = context.Background()
		newDatabaseSchema = func() string {
			return fmt.Sprintf("schema_%d", rand.Uint64())
		}
		newPrefix = func() string {
			return fmt.Sprintf("prefix_%d", rand.Uint64())
		}

		db = Connect(t)
	)

	t.Run("should create a postgres.Repository with no error", func(t *testing.T) {
		// arrange
		var (
			dbSchema = newDatabaseSchema()
			prefix   = newPrefix()
		)

		// act
		_, err := postgres.New(ctx, db, dbSchema, prefix)

		// assert
		assert.NoError(t, err)
	})

}
