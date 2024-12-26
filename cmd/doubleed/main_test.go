package main

import (
	"github.com/j0lvera/go-double-e/internal/testutils"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	ts := testutils.SetupTestServerWithRun(t, run)
	defer ts.Cleanup()

	resp, err := http.Get(ts.BaseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body := make([]byte, 1024)
	n, err := resp.Body.Read(body)
	if err != nil && err != io.EOF {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := `{"status":"ok"}`
	// Trim the response body to remove any trailing whitespace
	if strings.TrimSpace(string(body[:n])) != expected {
		t.Errorf("Expected response body '%s', got '%s'", expected, strings.TrimSpace(string(body[:n])))
	}
}

func TestCreateUser(t *testing.T) {
	// Set up the server
	ts := testutils.SetupTestServerWithRun(t, run)
	defer ts.Cleanup()

	apiUrl := ts.BaseURL + "/users"

	// Create a test user
	user := `{"email": "user@email.com", "password": "123456789012345678901234567890123456789012345678901234567890"}`
	resp, err := http.Post(apiUrl, "application/json", strings.NewReader(user))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	// Check the validation
	user = `{"email": "user@email.com", "password": "123"}`
	resp, err = http.Post(apiUrl, "application/json", strings.NewReader(user))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

}
