package main

import (
	"context"
	"github.com/j0lvera/go-double-e/internal/testutils"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// get the test database instance
	db, err := testutils.GetTestDB(ctx)
	if err != nil {
		log.Fatalf("GetTestDB() failed to setup test database: %v", err)
	}

	// verify the database connection before running any tests
	if err := db.Pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database during setup: %v", err)
	}
	log.Println("Test database setup successful!")

	// set the DATABASE_URL environment variable
	err = os.Setenv("DATABASE_URL", db.Pool.Config().ConnString())
	if err != nil {
		log.Fatalf("Failed to set DATABASE_URL: %v", err)
	}

	// run all tests
	code := m.Run()

	// cleanup after all tests complete
	db.Cleanup()

	os.Exit(code)
}

func TestDBSetup(t *testing.T) {
	ctx := context.Background()

	db, err := testutils.GetTestDB(ctx)
	if err != nil {
		t.Fatalf("GetTestDB() failed to get database: %v", err)
	}

	// Test the connection
	if err := db.Pool.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Test that we can reset the database
	if err := testutils.ResetTestData(ctx, db.Pool); err != nil {
		t.Fatalf("Failed to reset test data: %v", err)
	}

	t.Log("Database setup successful!")
}
