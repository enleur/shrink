package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func SetupRedis(t *testing.T) (testcontainers.Container, string) {
	ctx := context.Background()
	redisContainer, err := redis.Run(ctx,
		"docker.io/redis:7",
		redis.WithSnapshotting(10, 1),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %s", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("failed to get container port: %s", err)
	}

	return redisContainer, fmt.Sprintf("%s:%s", host, port.Port())
}
