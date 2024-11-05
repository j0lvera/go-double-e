package doublee

import (
	"context"
	"fmt"
	"github.com/j0lvera/go-double-e/internal/testutils"
	"log"
	"testing"
)

func TestDBCreation(t *testing.T) {
	ctx := context.Background()
	pool, cleanup, err := testutils.SetupTestDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Database setup successful!")
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	// setup test database
	pool, cleanup, err := testutils.SetupTestDB(ctx)
	if err != nil {
		t.Errorf("SetupTestDB() failed: %v", err)
	}
	defer cleanup()

	if pool == nil {
		t.Fatalf("SetupTestDB() failed: got nil, want pool")
	}
	log.Println("Database pool created successfully")

	// Test database connection
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("cannot ping database: %v", err)
	}
	log.Println("Database ping successful")

	log.Println("Creating new client...")
	client := NewClient(pool)
	if client == nil {
		t.Fatal("client is nil after creation")
	}
	log.Println("Client created successfully")

	if client.queries == nil {
		t.Fatal("client.queries is nil - SQLC queries not initialized")
	}
	log.Println("SQLC queries initialized successfully")

	log.Println("Attempting to create user...")
	user, err := client.CreateUser(ctx, "test@example.com", "123456789012345678901234567890123456789012345678901234567890")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("CreateUser() email = %v, want test@example.com", user.Email)
	}
}
