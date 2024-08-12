package shortener

import (
	"context"
	"testing"

	"github.com/enleur/shrink/internal/storage"
	"github.com/enleur/shrink/tests"
	"github.com/stretchr/testify/assert"
)

func TestShortenerService(t *testing.T) {
	ctx := context.Background()
	redisContainer, redisAddress := tests.SetupRedis(t)

	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	store, err := storage.NewRedisStore(redisAddress, 0)
	assert.NoError(t, err)

	service := NewService(store)

	t.Run("Shorten and Retrieve URL", func(t *testing.T) {
		longURL := "https://example.com"
		shortCode, err := service.ShortenURL(ctx, longURL)
		assert.NoError(t, err)
		assert.NotEmpty(t, shortCode)

		retrievedURL, err := service.GetLongURL(ctx, shortCode)
		assert.NoError(t, err)
		assert.Equal(t, longURL, retrievedURL)
	})
}
