package testutils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type TestServer struct {
	BaseURL string
	Handler http.Handler
	Server  *httptest.Server
	Cleanup func()
}

// SetupTestServerWithHandler creates a new test server with an existing http.Handler
func SetupTestServerWithHandler(t *testing.T, handler http.Handler) *TestServer {
	t.Helper()

	testDB, err := GetTestDB(context.Background())
	if err != nil {
		t.Fatalf("GetTestDB() failed to setup test database: %v", err)
	}

	server := httptest.NewServer(handler)

	return &TestServer{
		BaseURL: server.URL,
		Handler: handler,
		Server:  server,
		Cleanup: func() {
			server.Close()
			testDB.Cleanup()
		},
	}
}

// SetupTestServerWithRun creates a test server using a run function
func SetupTestServerWithRun(t *testing.T, runFn func(context.Context, io.Writer, int) error) *TestServer {
	t.Helper()

	ctx := context.Background()

	testDb, err := GetTestDB(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	err = os.Setenv("DATABASE_URL", testDb.Pool.Config().ConnString())
	if err != nil {
		t.Fatalf("Failed to set DATABASE_URL: %v", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	port := 8081

	go func() {
		if err := runFn(ctx, os.Stdout, port); err != nil {
			t.Errorf("runFn() failed to start server: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	return &TestServer{
		BaseURL: fmt.Sprintf("http://localhost:%d", port),
		Cleanup: func() {
			testDb.Cleanup()
		},
	}
}

// ReadResponseBody reads the response body and returns it as a string
func ReadResponseBody(t *testing.T, resp *http.Response) (string, error) {
	t.Helper()

	body := make([]byte, 1024)
	n, err := resp.Body.Read(body)
	if err != nil && err != io.EOF {
		//t.Fatalf("Failed to read response body: %v", err)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Trim the response body to remove any trailing whitespace
	bodyTrimmed := strings.TrimSpace(string(body[:n]))
	return bodyTrimmed, nil
}
