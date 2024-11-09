package doublee

import (
	"context"
	"github.com/j0lvera/go-double-e/internal/testutils"
	"log"
	"testing"
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	// setup test database
	db, err := testutils.GetTestDB(ctx)
	if err != nil {
		t.Errorf("Failed to get test database: %v", err)
	}

	// Reset the database before running the test
	if err := testutils.ResetTestData(ctx, db.Pool); err != nil {
		t.Fatalf("Failed to reset test data: %v", err)
	}

	log.Println("Creating new client...")
	client := NewClient(db.Pool)
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
