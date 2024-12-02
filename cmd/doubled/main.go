package main

import (
	"context"
	"fmt"
	"github.com/j0lvera/go-double-e/internal/server"
	"github.com/j0lvera/go-double-e/pkg/doublee"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func run(ctx context.Context, w io.Writer) error {
	// create a pool configuration
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		_ = fmt.Errorf("error parsing database url: %w", err)
		return err
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
		_ = fmt.Errorf("error creating connection pool: %w", err)
		return err
	}
	defer pool.Close()

	// verify the connection
	if err := pool.Ping(ctx); err != nil {
		_ = fmt.Errorf("error pinging connection pool: %w", err)
		return err
	}

	// initialize the client
	client := doublee.NewClient(pool)

	// initialize the server
	srv := server.NewServer(client)

	// start HTTP server
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: srv,
	}

	go func() {
		fmt.Fprintf(w, "Server listening on port 8080\n")
		if err := httpServer.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "error starting the server: %v\n", err)
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
	// ensure DATABASE_URL is set
	if os.Getenv("DATABASE_URL") == "" {
		fmt.Fprintf(os.Stderr, "DATABASE_URL environment variable is required\n")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error running server: %v\n", err)
		os.Exit(1)
	}
}
