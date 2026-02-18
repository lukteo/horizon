package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/luketeo/horizon/server/internal/config"
	"github.com/luketeo/horizon/server/internal/logprocessor"
	"github.com/luketeo/horizon/server/internal/messaging"
	"github.com/luketeo/horizon/server/internal/ocsf"
	"github.com/luketeo/horizon/server/internal/route"
	"github.com/luketeo/horizon/server/internal/storage"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	// Set up logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Set up configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		zlog.Warn().Err(err).Msg("Could not read config file, using defaults")
	}
}

func main() {
	// Initialize logger
	appLogger := zlog.With().Logger()

	// Load configuration
	config := config.LoadConfig()

	// Initialize services
	dbService, err := storage.NewDatabaseService(config.DatabaseURL)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer dbService.Close()

	natsService, err := messaging.NewNatsService(config.NatsURL)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to connect to NATS")
	}
	defer natsService.Close()

	blobStorage, err := storage.NewBlobStorageService(
		config.BlobStorageEndpoint,
		config.BlobStorageAccessKey,
		config.BlobStorageSecretKey,
		config.BlobStorageBucket,
		config.BlobStorageSecure,
	)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to connect to blob storage")
	}

	// Initialize services
	normalizer := ocsf.NewNormalizer(dbService)
	logProcessor := logprocessor.NewProcessor(natsService, normalizer, dbService, blobStorage)

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{config.ClientOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Initialize routes
	route.InitRoutes(r, dbService, normalizer)

	// Start log processor in a goroutine
	go func() {
		if err := logProcessor.Start(context.Background()); err != nil {
			appLogger.Error().Err(err).Msg("Log processor failed")
		}
	}()

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	// Run server in a goroutine
	go func() {
		appLogger.Info().Int("port", config.Port).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info().Msg("Shutting down server...")

	// Shutdown gracefully with a timeout of 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zlog.Error().Err(err).Msg("Server forced to shutdown")
	} else {
		zlog.Info().Msg("Server exited properly")
	}
}