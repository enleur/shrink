package shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

type Store interface {
	Set(ctx context.Context, key, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ShortenURL(ctx context.Context, longURL string) (string, error) {
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
