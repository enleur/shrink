package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enleur/shrink/internal/api"
	"github.com/enleur/shrink/internal/api/middleware"
	"github.com/enleur/shrink/internal/config"
	"github.com/enleur/shrink/internal/shortener"
	"github.com/enleur/shrink/internal/storage"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const ServiceName = "shrink-service"

func main() {
	conf, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := initLogger(conf.Server.Mode)
	defer func() { _ = logger.Sync() }()

	tp, err := initTracer(conf.Otel)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Info("error shutting down tracer provider", zap.Error(err))
		}
	}()

	redis, err := storage.NewRedisStore(conf.Redis.Address, conf.Redis.DB)
	if err != nil {
		logger.Fatal("failed to init redis store", zap.Error(err))
	}
	defer func() { _ = redis.Close() }()

	short := shortener.NewService(redis)
	server := api.NewServer(logger, short)

	router := setupRouter(logger, server)

	srv := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%d", conf.Server.Port),
	}

	err = serveHTTP(srv, logger)
	if err != nil {
		logger.Fatal("failed to serve http", zap.Error(err))
	}
}

func initLogger(mode string) *zap.Logger {
	if mode == "release" {
		return zap.Must(zap.NewProduction())
	}
	return zap.Must(zap.NewDevelopment())
}

func initTracer(conf config.OtelConfig) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(conf.Endpoint),
			otlptracegrpc.WithInsecure(),
		),
	)
	if err != nil {
		return nil, err
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ServiceName),
		),
	)

	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(r),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func setupRouter(logger *zap.Logger, server *api.Server) *gin.Engine {
	r := gin.New()
	r.Use(middleware.PrometheusMiddleware())
	r.Use(otelgin.Middleware(ServiceName))
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	r.GET(middleware.MetricsPath, gin.WrapH(promhttp.Handler()))
	api.RegisterHandlers(r, server)
	return r
}

func serveHTTP(srv *http.Server, logger *zap.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to start server: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
		logger.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}
		logger.Info("Server exiting")
	}

	return nil
}
