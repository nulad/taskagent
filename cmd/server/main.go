package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/nulad/taskagent/internal/config"
	"github.com/nulad/taskagent/internal/handler"
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
	appStore, err := store.NewStore(cfg.DatabasePath)
	if err != nil {
		return err
	}
	defer appStore.Close()

	projectService := service.NewProjectService(appStore)
	taskService := service.NewTaskService(appStore)

	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)
	authHandler := handler.NewAuthHandler(appStore)

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

	log.Printf("listening on %s", cfg.ListenAddr)
	return http.ListenAndServe(cfg.ListenAddr, mux)
}

func runSeed(cfg config.Config, args []string) error {
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

	appStore, err := store.NewStore(cfg.DatabasePath)
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

	fmt.Fprintln(os.Stdout, rawKey)
	return nil
}
