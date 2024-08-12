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
	"github.com/enleur/shrink/internal/config"
	"github.com/enleur/shrink/internal/shortener"
	"github.com/enleur/shrink/internal/storage"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
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

func main() {
	conf, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := zap.Must(zap.NewDevelopment())
	if conf.Server.Mode == "release" {
		logger = zap.Must(zap.NewProduction())
	}
	defer logger.Sync() //nolint:errcheck

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
	defer redis.Close()

	short := shortener.NewService(redis)

	r := gin.New()
	r.Use(otelgin.Middleware("shrink-service"))
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
			semconv.ServiceName("shrink-service"),
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
