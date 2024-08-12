package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockShortener struct {
	mock.Mock
}

func (m *MockShortener) ShortenURL(ctx context.Context, longURL string) (string, error) {
	args := m.Called(ctx, longURL)
	return args.String(0), args.Error(1)
}

func (m *MockShortener) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	args := m.Called(ctx, shortCode)
	return args.String(0), args.Error(1)
}

func TestPostShorten(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockShortener := new(MockShortener)
	logger, _ := zap.NewDevelopment()
	server := NewServer(logger, mockShortener)

	t.Run("Successful Shortening", func(t *testing.T) {
		mockShortener.On("ShortenURL", mock.Anything, "https://example.com").Return("abc123", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(`{"url":"https://example.com"}`))

		server.PostShorten(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response struct {
			ShortUrl string `json:"shortUrl"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "abc123", response.ShortUrl)
	})
}
