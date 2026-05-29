package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"omepic/backend/internal/auth"
	"omepic/backend/internal/cache"
	"omepic/backend/internal/config"
	"omepic/backend/internal/http/clientip"
	"omepic/backend/internal/http/handler"
	"omepic/backend/internal/http/router"
	"omepic/backend/internal/ratelimit"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/secrets"
	"omepic/backend/internal/service"
	"omepic/backend/internal/storage"
	"omepic/backend/internal/uid"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()

	// H-01: enforce required secrets at startup.
	if err := enforceRequiredSecrets(cfg); err != nil {
		logger.Error("startup failed", "error", err.Error())
		os.Exit(1)
	}

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

	secretEncryptor, err := secrets.NewEncryptor(cfg.SecretEncryptionKey)
	if err != nil {
		logger.Error("failed to init secret encryptor", "error", err.Error())
		os.Exit(1)
	}

	// Initialize the storage catalog (may seed default configs from env).
	if _, err := repo.InitializeStorageCatalog(ctx, config.DefaultStorageConfig()); err != nil {
		logger.Error("failed to initialize storage catalog", "error", err.Error())
		os.Exit(1)
	}

	// Encrypt sensitive fields in storage configs and write back to DB.
	if err := service.EncryptStorageCatalogSecrets(ctx, repo, secretEncryptor); err != nil {
		logger.Error("failed to encrypt storage secrets", "error", err.Error())
		os.Exit(1)
	}

	// Reload catalog from DB (now with encrypted values) and decrypt for storage manager.
	encryptedConfigs, err := repo.ListStorageConfigs(ctx)
	if err != nil {
		logger.Error("failed to reload storage configs", "error", err.Error())
		os.Exit(1)
	}
	decryptedForManager := service.DecryptStorageCatalogValues(encryptedConfigs, secretEncryptor)

	storageManager, err := storage.NewManager(decryptedForManager)
	if err != nil {
		logger.Error("failed to init storage", "error", err.Error())
		os.Exit(1)
	}

	uidCodec, err := uid.NewCodec(cfg.UIDPrefix, cfg.UIDEncryptionKey)
	if err != nil {
		logger.Error("failed to init uid codec", "error", err.Error())
		os.Exit(1)
	}

	redisClient, err := cache.NewClient(cfg.RedisURL)
	if err != nil {
		logger.Error("failed to create redis client", "error", err.Error())
		os.Exit(1)
	}
	defer redisClient.Close()
	imageCache := cache.NewWithClient(redisClient)
	rateLimiter := ratelimit.NewRedisLimiter(redisClient)

	if err := repo.Ping(ctx); err != nil {
		logger.Error("sqlite ping failed", "error", err.Error())
		os.Exit(1)
	}
	if err := imageCache.Ping(ctx); err != nil {
		logger.Error("redis ping failed", "error", err.Error())
		os.Exit(1)
	}

	settingsManager := service.NewRuntimeSettingsManager()
	if err := settingsManager.Load(ctx, repo); err != nil {
		logger.Error("failed to load runtime settings", "error", err.Error())
		os.Exit(1)
	}

	// M-03: seed PUBLIC_BASE_URL env into runtime settings and enforce in production.
	if cfg.PublicBaseURL != "" {
		s := settingsManager.Current()
		if s.PublicBaseURL == "" {
			s.PublicBaseURL = cfg.PublicBaseURL
			settingsManager.Reconfigure(s)
		}
	}
	if cfg.IsProduction() && settingsManager.Current().PublicBaseURL == "" {
		logger.Error("PUBLIC_BASE_URL must be configured in production (set PUBLIC_BASE_URL env or configure public_base_url in runtime settings)")
		os.Exit(1)
	}

	revChecker := auth.NewRevocationChecker(redisClient)
	imageService := service.NewImageService(repo, imageCache, storageManager, settingsManager, uidCodec.Generate, uidCodec.Validate, logger)

	adminService := service.NewAdminService(repo, storageManager, settingsManager, imageService, cfg.JWTSecret, revChecker, secretEncryptor, service.AdminEnvMetadata{
		HTTPAddr:         cfg.HTTPAddr,
		DatabasePath:     cfg.DatabasePath,
		RedisURL:         cfg.RedisURL,
		UIDEncryptionKey: cfg.UIDEncryptionKey,
	})
	announcementService := service.NewAnnouncementService(repo)
	healthService := service.NewHealthService(repo, imageCache)
	storageHealthService := service.NewStorageHealthService(repo, storageManager, logger)
	stopStorageHealthHeartbeat := storageHealthService.StartHeartbeat(context.Background(), service.StorageHealthDefaultInterval)
	ipResolver := clientip.NewResolver(nil, func() string {
		return settingsManager.Current().RealIPSource
	})

	if _, err := imageService.Preheat(ctx); err != nil {
		logger.Error("redis preheat failed", "error", err.Error())
		os.Exit(1)
	}

	engine := router.New(router.Dependencies{
		Logger:              logger,
		ImageHandler:        handler.NewImageHandler(imageService, logger, ipResolver),
		AdminHandler:        handler.NewAdminHandler(adminService, logger),
		AnnouncementHandler: handler.NewAnnouncementHandler(announcementService, logger),
		HealthHandler:       handler.NewHealthHandler(healthService),
		Settings:            settingsManager,
		ImageService:        imageService,
		RateLimiter:         rateLimiter,
		IPResolver:          ipResolver,
		JWTSecret:           cfg.JWTSecret,
		RevChecker:          revChecker,
		FrontendDir:         "web",
	})

	// --- Graceful shutdown ---
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      300 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", cfg.HTTPAddr, "default_storage_key", storageManager.CurrentKey(), "storage_backend", storageManager.CurrentBackend())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-serverErr:
		if err != nil {
			logger.Error("server stopped unexpectedly", "error", err.Error())
			os.Exit(1)
		}
	}

	// Phase 1: stop accepting new HTTP requests
	shutdownCtx, shutdownCancel := context.WithTimeout(rootCtx, 15*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err.Error())
	}

	// Phase 2: stop background heartbeat
	stopStorageHealthHeartbeat()

	// Phase 3: close Redis and SQLite (deferred in reverse order)
	logger.Info("server stopped gracefully")
}

func enforceRequiredSecrets(cfg config.AppConfig) error {
	if cfg.JWTSecret == "" || len(cfg.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be set in .env and be at least 32 characters")
	}
	// UID_ENCRYPTION_KEY is an obfuscation key (XOR-based ID encoding), not a cryptographic boundary.
	// The env var name is kept as-is for deployment compatibility.
	if cfg.UIDEncryptionKey == "" || len(cfg.UIDEncryptionKey) < 32 {
		return errors.New("UID_ENCRYPTION_KEY must be set in .env and be at least 32 characters")
	}
	if cfg.SecretEncryptionKey == "" || len(cfg.SecretEncryptionKey) < 32 {
		return errors.New("SECRET_ENCRYPTION_KEY must be set in .env and be at least 32 characters")
	}
	return nil
}
