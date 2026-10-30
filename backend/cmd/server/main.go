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

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/HeroicHorizon/user-management/backend/internal/config"
	"github.com/HeroicHorizon/user-management/backend/internal/handler"
	"github.com/HeroicHorizon/user-management/backend/internal/repository"
	"github.com/HeroicHorizon/user-management/backend/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		lvl = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})))

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI).SetServerSelectionTimeout(5 * time.Second))
	if err != nil {
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	}()

	pingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return errors.New("cannot connect to MongoDB")
	}
	slog.Info("connected to MongoDB", "db", cfg.MongoDB)

	repo := repository.NewMongoUserRepository(client.Database(cfg.MongoDB))
	if err := repo.EnsureIndexes(pingCtx); err != nil {
		return errors.New("cannot ensure MongoDB indexes")
	}

	h := handler.NewUserHandler(service.NewUserService(repo, nil))
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.Chain(h.Routes(), cfg.CORSAllowedOrigin, cfg.MaxBodyBytes),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-sigCtx.Done():
		slog.Info("shutting down")
		shutCtx, c := context.WithTimeout(context.Background(), 10*time.Second)
		defer c()
		return srv.Shutdown(shutCtx)
	}
	return nil
}
