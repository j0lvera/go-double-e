package main

import (
	"github.com/j0lvera/go-double-e/internal/testutils"
	"net/http"
	"runtime"
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
