package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the pgx driver
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type TestDB struct {
	Pool    *pgxpool.Pool
	Cleanup func()
}

var (
	testDB     *TestDB
	setupOnce  sync.Once
	setupError error
)

// GetTestDB returns a singleton database instance for testing.
// The database is created only once and reused for subsequent calls.
// Callers should ensure they call ResetTestData() between tests if needed.
func GetTestDB(ctx context.Context) (*TestDB, error) {
	setupOnce.Do(func() {
		pool, cleanup, err := SetupTestDB(ctx)
		if err != nil {
			setupError = err
			return
		}

		testDB = &TestDB{
			Pool:    pool,
			Cleanup: cleanup,
		}
	})

	if setupError != nil {
		return nil, fmt.Errorf("failed to setup test database: %w", setupError)
	}

	return testDB, nil
}

// ResetTestData truncates all tables in the test database.
// Call this between tests to ensure test isolation.
func ResetTestData(ctx context.Context, db *pgxpool.Pool) error {
	// Get all table names
	rows, err := db.Query(ctx, `
		select tablename
		from pg_catalog.pg_tables
		where schemaname = 'public';
	`)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	// Truncate each table
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}

		_, err := db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", tableName, err)
		}
	}

	return rows.Err()
}

func SetupTestDB(ctx context.Context) (*pgxpool.Pool, func(), error) {
	// get the project root to find migrations
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "..", "..")

	// setup postgres container
	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %s", err)
	}

	// get connection details
	mappedPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		_ = postgresContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get mapped port: %s", err)
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		_ = postgresContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get host: %s", err)
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, mappedPort.Port())
	log.Printf("Database DSN: %s", dsn)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		_ = postgresContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to open DB for migrations: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		_ = postgresContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	if err := goose.Up(db, filepath.Join(projectRoot, "internal/db/migrations")); err != nil {
		_ = db.Close()
		_ = postgresContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	_ = db.Close()

	// Create connection pool for actual usage
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to parse config: %s", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to create pool: %s", err)
	}

	// Test the pool connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		postgresContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to ping database pool: %s", err)
	}

	log.Println("Test database setup completed successfully")

	cleanup := func() {
		log.Println("Cleaning up test database...")
		pool.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate container during cleanup: %s", err)
		}
	}

	return pool, cleanup, nil
}
