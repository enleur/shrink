package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/enleur/shrink/internal/api"
	"github.com/enleur/shrink/internal/shortener"
	"github.com/enleur/shrink/internal/storage"
)

func TestEndToEnd(t *testing.T) {
	ctx := context.Background()
	redisContainer, redisAddress := SetupRedis(t)

	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	store, err := storage.NewRedisStore(redisAddress, 0)
	assert.NoError(t, err)

	shortenerService := shortener.NewService(store)

	logger, _ := zap.NewDevelopment()

	server := api.NewServer(logger, shortenerService)

	router := gin.New()
	api.RegisterHandlers(router, server)

	t.Run("Shorten URL", func(t *testing.T) {
		w := httptest.NewRecorder()
		url := "https://example.com"
		body := api.PostShortenJSONRequestBody{Url: &url}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/shorten", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			ShortUrl string `json:"shortUrl"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.ShortUrl)

		t.Run("Retrieve URL", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/"+response.ShortUrl, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusFound, w.Code)
			assert.Equal(t, "https://example.com", w.Header().Get("Location"))
		})
	})

	t.Run("Shorten Invalid URL", func(t *testing.T) {
		w := httptest.NewRecorder()
		invalidUrl := "not_a_valid_url"
		body := api.PostShortenJSONRequestBody{Url: &invalidUrl}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/shorten", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Retrieve Non-existent URL", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
