package main

import (
	"context"
	"github.com/j0lvera/go-double-e/internal/testutils"
	"net/http"
	"runtime"
	"strings"
	"testing"
)

var testServer *testutils.TestServer

func init() {
	t := &testing.T{}
	testServer = testutils.SetupTestServerWithRun(t, run)

	// Register cleanup to run at the program exit
	runtime.SetFinalizer(testServer, func(ts *testutils.TestServer) {
		ts.Cleanup()
	})
}

func TestHealthCheck(t *testing.T) {
	apiUrl := testServer.BaseURL + "/health"

	resp, err := http.Get(apiUrl)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	//defer resp.Body.Close()
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := testutils.ReadResponseBody(t, resp)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := `{"status":"ok"}`

	if body != expected {
		t.Errorf("Expected response body '%s', got '%s'", expected, body)
	}
}

func TestCreateUser(t *testing.T) {
	apiUrl := testServer.BaseURL + "/users"

	testDb, err := testutils.GetTestDB(context.Background())
	if err != nil {
		t.Fatalf("Failed to get a test database connection: %v", err)
	}

	// Test a successful user creation
	user := `{"email": "user@email.com", "password": "123456789012345678901234567890123456789012345678901234567890"}`
	resp, err := http.Post(apiUrl, "application/json", strings.NewReader(user))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	// defer resp.Body.Close()
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	// Check if the user was created in the database
	var userExists bool
	err = testDb.Pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", "user@email.com").Scan(&userExists)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if !userExists {
		t.Errorf("User was not created in the database")
	}

	// Test a validation error
	user = `{"email": "validation@email.com", "password": "123"}`
	resp, err = http.Post(apiUrl, "application/json", strings.NewReader(user))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	// defer resp.Body.Close()
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}()

	// Read the body of the response
	body, err := testutils.ReadResponseBody(t, resp)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	expected := `{"status":400,"message":{"errors":[{"Field":"Password","Message":"This field must be at least 8 characters long"}]}}`
	if body != expected {
		t.Errorf("Expected response body '%s', got '%s'", expected, body)
	}
}
