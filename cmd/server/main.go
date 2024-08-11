package main

import (
	"context"
	"errors"
	"fmt"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"shrink/internal/api"
	"shrink/internal/config"
	"shrink/internal/shortener"
	"shrink/internal/storage"
	"syscall"
	"time"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := zap.Must(zap.NewDevelopment())
	if os.Getenv("GIN_MODE") == "release" {
		logger = zap.Must(zap.NewProduction())
	}
	defer logger.Sync() //nolint:errcheck

	redis, err := storage.NewRedisStore(conf.Redis.Address, conf.Redis.DB)
	if err != nil {
		logger.Fatal("failed to init redis store", zap.Error(err))
	}
	defer redis.Close()

	short := shortener.NewService(redis)

	r := gin.New()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	server := api.NewServer(logger, short)
	api.RegisterHandlers(r, server)

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf(":%d", conf.Server.Port),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("listen", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown", zap.Error(err))
	}
	<-ctx.Done()
	logger.Info("Server exiting")
}
