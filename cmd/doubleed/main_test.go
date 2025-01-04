package main

import (
	"context"
	"github.com/j0lvera/go-double-e/internal/testutils"
	is_ "github.com/matryer/is"
	"net/http"
	"runtime"
	"strings"
	"testing"
)

var testServer *testutils.TestServer

func init() {
	t := &testing.T{}
	testServer = testutils.SetupTestServerWithRun(t, run)

	// register cleanup to run at the program exit
	runtime.SetFinalizer(testServer, func(ts *testutils.TestServer) {
		ts.Cleanup()
	})
}

func TestHealthCheck(t *testing.T) {
	is := is_.New(t)

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

	is.Equal(resp.StatusCode, http.StatusOK) // invalid status code

	body, err := testutils.ReadResponseBody(t, resp)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := `{"status":"ok"}`
	is.Equal(body, expected) // invalid response body
}

func TestCreateLedger(t *testing.T) {
	is := is_.New(t)

	apiUrl := testServer.BaseURL + "/ledgers"

	testDb, err := testutils.GetTestDB(context.Background())
	if err != nil {
		t.Fatalf("unable to setup test database: %v", err)
	}

	metadata := `{"user_id": 24}`
	reqBody := `{"name": "Test Ledger", "description": "This is a test ledger", "metadata": ` + metadata + `}`

	t.Run("should return 201 for valid request", func(t *testing.T) {
		resp, err := http.Post(apiUrl, "application/json", strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("unable to make POST request: %v", err)
		}
		// defer resp.Body.Close()
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				t.Fatalf("unable to close response body: %v", err)
			}
		}()

		is.Equal(resp.StatusCode, http.StatusCreated) // invalid status code
	})

	t.Run("should return 400 for invalid request", func(t *testing.T) {
		t.Run("should return 400 for invalid metadata JSON", func(t *testing.T) {
			// create a request with invalid metadata JSON
			invalidBody := `{"name": "Test Ledger", "description": "This is a test ledger", "metadata": "invalid metadata"}`
			resp, err := http.Post(apiUrl, "application/json", strings.NewReader(invalidBody))
			if err != nil {
				t.Fatalf("unable to make POST request: %v", err)
			}
			// defer resp.Body.Close()
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					t.Fatalf("unable to close response body: %v", err)
				}
			}()

			is.Equal(resp.StatusCode, http.StatusBadRequest) // invalid status code
		})

		t.Run("should return 400 for invalid metadata JSON", func(t *testing.T) {
			// create a request without name
			invalidBody := `{"description": "This is a test ledger", "metadata": "{"user_id": 24}"}`
			resp, err := http.Post(apiUrl, "application/json", strings.NewReader(invalidBody))
			if err != nil {
				t.Fatalf("unable to make POST request: %v", err)
			}
			// defer resp.Body.Close()
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					t.Fatalf("unable to close response body: %v", err)
				}
			}()

			is.Equal(resp.StatusCode, http.StatusBadRequest) // invalid status code
		})
	})

	t.Run("should create ledger in database", func(t *testing.T) {
		// check if ledger was created in the database
		var ledgerExists bool
		err = testDb.Pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM ledgers WHERE name = $1)", "Test Ledger").Scan(&ledgerExists)
		if err != nil {
			t.Fatalf("unable to query database: %v", err)
		}

		is.True(ledgerExists) // unable to find ledger in database
	})
}
