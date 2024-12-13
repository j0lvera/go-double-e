package testutils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	testDB, err := GetTestDB(ctx)
	if err != nil {
		t.Fatalf("GetTestDB() failed to setup test database: %v", err)
	}

	port := 8081
	os.Setenv("DATABASE_URL", testDB.Pool.Config().ConnString())

	go func() {
		if err := runFn(ctx, os.Stdout, port); err != nil {
			t.Errorf("runFn() failed to start server: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	return &TestServer{
		BaseURL: fmt.Sprintf("http://localhost:%d", port),
		Cleanup: func() {
			testDB.Cleanup()
		},
	}
}