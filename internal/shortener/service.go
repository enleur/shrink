package shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Store interface {
	Set(ctx context.Context, key, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type Shortener interface {
	ShortenURL(ctx context.Context, longURL string) (string, error)
	GetLongURL(ctx context.Context, shortCode string) (string, error)
}

type Service struct {
	store  Store
	tracer trace.Tracer
}

func NewService(store Store) *Service {
	return &Service{
		store:  store,
		tracer: otel.Tracer("shrink-service"),
	}
}

func (s *Service) ShortenURL(ctx context.Context, longURL string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "ShortenURL")
	defer span.End()

	parsedURL, err := url.Parse(longURL)
	if err != nil {
		return "", InvalidURLError{Reason: err.Error()}
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", InvalidURLError{Reason: "missing scheme or host"}
	}

	shortCode, err := generateShortCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate short code: %w", err)
	}

	err = s.store.Set(ctx, shortCode, longURL, 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to store URL: %w", err)
	}

	return shortCode, nil
}

func (s *Service) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "GetLongURL")
	defer span.End()

	longURL, err := s.store.Get(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve long URL: %w", err)
	}

	return longURL, nil
}

func generateShortCode() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b)[:6], nil
}
