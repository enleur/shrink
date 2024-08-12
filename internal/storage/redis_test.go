package storage

import (
	"context"
	"testing"
	"time"

	"github.com/enleur/shrink/tests"
	"github.com/stretchr/testify/assert"
)

func TestRedisStore(t *testing.T) {
	ctx := context.Background()
	redisContainer, redisAddress := tests.SetupRedis(t)

	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	store, err := NewRedisStore(redisAddress, 0)
	assert.NoError(t, err)

	t.Run("Set and Get", func(t *testing.T) {
		err := store.Set(ctx, "testKey", "testValue", time.Minute)
		assert.NoError(t, err)

		value, err := store.Get(ctx, "testKey")
		assert.NoError(t, err)
		assert.Equal(t, "testValue", value)
	})

	t.Run("Get Non-Existent Key", func(t *testing.T) {
		_, err := store.Get(ctx, "nonExistentKey")
		assert.Error(t, err)
	})
}
