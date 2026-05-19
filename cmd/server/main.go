package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nulad/taskagent/internal/config"
	"github.com/nulad/taskagent/internal/handler"
	"github.com/nulad/taskagent/internal/logging"
	"github.com/nulad/taskagent/internal/middleware"
	"github.com/nulad/taskagent/internal/service"
	"github.com/nulad/taskagent/internal/store"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	cfg := config.Load()

	if len(args) > 0 && args[0] == "seed" {
		return runSeed(cfg, args[1:])
	}

	return runServer(cfg)
}

func runServer(cfg config.Config) error {
	logLevel := logging.ParseLevel(cfg.LogLevel)
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(logHandler)

	appStore, err := store.NewStore(cfg.DatabasePath, logger)
	if err != nil {
		return err
	}

	projectService := service.NewProjectService(appStore, logger)
	taskService := service.NewTaskService(appStore, logger)

	projectHandler := handler.NewProjectHandler(projectService, logger)
	taskHandler := handler.NewTaskHandler(taskService, logger)
	authHandler := handler.NewAuthHandler(appStore, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	protectedMux := http.NewServeMux()
	handler.RegisterProjectRoutes(protectedMux, projectHandler)
	handler.RegisterTaskRoutes(protectedMux, taskHandler)
	handler.RegisterAuthRoutes(protectedMux, authHandler)

	protectedAPI := middleware.AuthMiddleware(appStore)(protectedMux)
	mux.Handle("/", protectedAPI)

	// Wrap the mux with CORS to apply it to all routes, including /health and protectedAPI
	corsHandler := middleware.CORSMiddleware(cfg.CORSOrigins)(mux)

	finalHandler := middleware.RequestIDMiddleware()(
		middleware.RequestLoggingMiddleware(logger)(corsHandler),
	)

	server := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: finalHandler,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	logger.Info("listening on", "address", cfg.ListenAddr)

	select {
	case err := <-serverErr:
		if closeErr := appStore.Close(); closeErr != nil {
			logger.Error("database close failed", "error", closeErr)
			if err == nil {
				return closeErr
			}
		}
		return err
	case sig := <-sigChan:
		logger.Info("shutting down...", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("shutdown timed out", "error", err)
			if closeErr := server.Close(); closeErr != nil {
				logger.Error("server close failed", "error", closeErr)
			}
			if closeErr := appStore.Close(); closeErr != nil {
				logger.Error("database close failed", "error", closeErr)
			}
			return err
		}

		if closeErr := appStore.Close(); closeErr != nil {
			logger.Error("database close failed", "error", closeErr)
		}
		return fmt.Errorf("shutdown server: %w", err)
	}

	if err := <-serverErr; err != nil {
		if closeErr := appStore.Close(); closeErr != nil {
			logger.Error("database close failed", "error", closeErr)
		}
		return err
	}

	if err := appStore.Close(); err != nil {
		logger.Error("database close failed", "error", err)
		return err
	}

	logger.Info("shutdown complete")
	return nil
}

func runSeed(cfg config.Config, args []string) error {
	logLevel := logging.ParseLevel(cfg.LogLevel)
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(logHandler)

	flags := flag.NewFlagSet("seed", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	userName := flags.String("user", "admin", "user name to create or reuse")
	label := flags.String("label", "bootstrap", "API key label")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}

	appStore, err := store.NewStore(cfg.DatabasePath, logger)
	if err != nil {
		return err
	}
	defer appStore.Close()

	ctx := context.Background()

	existing, err := appStore.ListApiKeys(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return fmt.Errorf("seed refused: %d API key(s) already exist; use POST /auth/keys to mint additional keys", len(existing))
	}

	user, err := appStore.GetUserByName(ctx, *userName)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			return err
		}

		user, err = appStore.CreateUser(ctx, *userName, true)
		if err != nil {
			return err
		}
	}

	_, rawKey, err := appStore.CreateApiKey(ctx, *label, user.ID)
	if err != nil {
		return err
	}

	logger.Info("seed completed", "user", *userName, "key", rawKey)
	return nil
}
