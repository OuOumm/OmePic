package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"omepic/backend/internal/cache"
	"omepic/backend/internal/config"
	"omepic/backend/internal/http/handler"
	"omepic/backend/internal/http/router"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/service"
	"omepic/backend/internal/storage"
	"omepic/backend/internal/uid"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := os.MkdirAll("data", 0o755); err != nil {
		logger.Error("failed to create data directory", "error", err.Error())
		os.Exit(1)
	}

	repo, err := repository.New(cfg.DatabasePath)
	if err != nil {
		logger.Error("failed to open sqlite", "error", err.Error())
		os.Exit(1)
	}
	defer repo.Close()

	if err := repo.Migrate(ctx); err != nil {
		logger.Error("migration failed", "error", err.Error())
		os.Exit(1)
	}

	storageCatalog, err := repo.InitializeStorageCatalog(ctx, cfg.DefaultStorageConfig())
	if err != nil {
		logger.Error("failed to initialize storage catalog", "error", err.Error())
		os.Exit(1)
	}

	storageManager, err := storage.NewManager(storageCatalog.StorageConfigs)
	if err != nil {
		logger.Error("failed to init storage", "error", err.Error())
		os.Exit(1)
	}

	uidCodec, err := uid.NewCodec(cfg.UIDPrefix, cfg.UIDEncryptionKey)
	if err != nil {
		logger.Error("failed to init uid codec", "error", err.Error())
		os.Exit(1)
	}

	imageCache, err := cache.New(cfg.RedisURL)
	if err != nil {
		logger.Error("failed to create redis client", "error", err.Error())
		os.Exit(1)
	}
	defer imageCache.Close()

	if err := repo.Ping(ctx); err != nil {
		logger.Error("sqlite ping failed", "error", err.Error())
		os.Exit(1)
	}
	if err := imageCache.Ping(ctx); err != nil {
		logger.Error("redis ping failed", "error", err.Error())
		os.Exit(1)
	}

	imageService := service.NewImageService(repo, imageCache, storageManager, uidCodec.Generate, uidCodec.Validate, logger)
	adminService := service.NewAdminService(repo, storageManager, imageService, cfg.AdminPassword, cfg.JWTSecret)

	if _, err := imageService.Preheat(ctx); err != nil {
		logger.Error("redis preheat failed", "error", err.Error())
		os.Exit(1)
	}

	engine := router.New(router.Dependencies{
		Logger:        logger,
		ImageHandler:  handler.NewImageHandler(imageService, storageManager, logger, cfg.PublicBaseURL),
		AdminHandler:  handler.NewAdminHandler(adminService, logger),
		HealthHandler: handler.NewHealthHandler(repo, imageCache),
		JWTSecret:     cfg.JWTSecret,
	})

	logger.Info("server starting", "addr", cfg.HTTPAddr, "default_storage_key", storageManager.CurrentKey(), "storage_backend", storageManager.CurrentBackend())
	if err := engine.Run(cfg.HTTPAddr); err != nil {
		logger.Error("server stopped", "error", err.Error())
		os.Exit(1)
	}
}
