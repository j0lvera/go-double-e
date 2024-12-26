package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/j0lvera/go-double-e/internal/db"
	"github.com/j0lvera/go-double-e/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func run(ctx context.Context, w io.Writer, port int) error {
	dbUrl := os.Getenv("DATABASE_URL")

	// create a pool configuration
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return fmt.Errorf("error parsing database url: %w", err)
	}

	// pool manual configuration
	config.MaxConns = 25                      // Maximum number of connections
	config.MinConns = 5                       // Minimum number of connections
	config.MaxConnLifetime = time.Hour        // Maximum lifetime of a connection
	config.MaxConnIdleTime = 30 * time.Minute // Maximum idle time for a connection
	config.HealthCheckPeriod = time.Minute    // How often to check connection health

	// create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("error creating connection pool: %w", err)
	}
	defer pool.Close()

	slog.Info("Database connection pool created")

	// verify the connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("error pinging connection pool: %w", err)
	}

	// initialize the client
	client := db.NewClient(pool)

	// initialize the server
	srv := server.NewServer(client)

	// start HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: srv,
	}

	go func() {
		slog.Info("Server is listening", "port", port)
		if err := httpServer.ListenAndServe(); err != nil {
			slog.Error("Could not start the server", "error", err)
		}
	}()

	// wait to interrupt
	<-ctx.Done()

	// gracefully shutdown the server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}

func main() {
	// set up logging
	handler := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
	})
	if os.Getenv("DEBUG") == "true" {
		handler.SetLevel(log.DebugLevel)
	} else {
		handler.SetLevel(log.InfoLevel)
	}
	slog.SetDefault(slog.New(handler))

	// ensure DATABASE_URL is set
	if os.Getenv("DATABASE_URL") == "" {
		slog.Error("DATABASE_URL environment variable is required", "DATABASE_URL", os.Getenv("DATABASE_URL"))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, os.Stdout, 8080); err != nil {
		slog.Error("Error running server", "error", err)
		os.Exit(1)
	}
}
